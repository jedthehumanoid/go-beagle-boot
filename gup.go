package main

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

const romvid gousb.ID = 0x0451
const rompid gousb.ID = 0x6141

const splvid gousb.ID = 0x525
const splpid gousb.ID = 0xa4a2

const ipUDP = 17
const rndisSize = 44
const etherSize = 14
const arpSize = 28
const ipSize = 20
const udpSize = 8
const bootpSize = 300
const tftpSize = 4
const fullSize = 386

var serverHwaddr = [6]byte{0x9a, 0x1f, 0x85, 0x1c, 0x3d, 0x0e}
var serverIP = [4]byte{192, 168, 1, 9}
var bbIP = [4]byte{192, 168, 1, 3}

const maxbuf = 450

var filename = "spl"

var ctx *gousb.Context

func readUSB(ep *gousb.InEndpoint) []byte {

	buf := make([]byte, 10*ep.Desc.MaxPacketSize)
	fmt.Println("Reading...")
	bytesread, err := ep.Read(buf)
	check(err)
	buf2 := buf[:bytesread]
	fmt.Printf("Receiving: --%d-- % x\n", len(buf2), buf2)
	return buf2
}

func sendUSB(ep *gousb.OutEndpoint, data []byte) {
	byteswritten, err := ep.Write(data)
	check(err)
	fmt.Println(byteswritten, len(data))
	fmt.Printf("Sending: --%d-- % x\n", len(data), data)
}

func listDevices() []string {
	var ret []string
	_, _ = ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		ret = append(ret, fmt.Sprintf("%s %s", desc.Vendor, desc.Product))
		return false
	})
	return ret
}

func onAttach(ctx *gousb.Context) string {
	oldDevices := listDevices()
	for {
		newDevices := listDevices()
		//fmt.Println(difference(newDevices, oldDevices))

		attached := difference(newDevices, oldDevices)
		if len(attached) > 0 {
			return attached[0]
		}
		oldDevices = newDevices
		time.Sleep(time.Millisecond * 100)
	}
}

func sendSPL() bool {

	dev, err := ctx.OpenDeviceWithVIDPID(romvid, rompid)
	if dev == nil {
		return false
	}
	check(err)
	defer dev.Close()
	fmt.Println("Sending SPL...")

	dev.SetAutoDetach(true)

	config, err := dev.Config(1)
	check(err)
	defer config.Close()

	intf, err := config.Interface(1, 0)
	check(err)
	defer intf.Close()

	iep, err := intf.InEndpoint(1)
	check(err)

	oep, err := intf.OutEndpoint(2)
	check(err)

	for {
		in := readUSB(iep)
		request := identifyRequest(in, len(filename))

		fmt.Println(request)
		var data []byte
		if request == "BOOTP" {
			data = processBOOTP(in)
		} else if request == "ARP" {
			//	data = []byte("")
			data = processARP(in)
		} else if request == "TFTP" {
			data = processTFTP(in)
		} else if request == "TFTP_Data" {
			data = processTFTPData(in)
			if string(data) == "" {
				return true
			}
		}
		if string(data) != "" {
			sendUSB(oep, data)
		}
	}

}

func findSPL() bool {

	filename = "uboot"

	dev, err := ctx.OpenDeviceWithVIDPID(splvid, splpid)
	if dev == nil {
		return false
	}
	check(err)
	defer dev.Close()
	fmt.Println("Sending UBOOT...")

	dev.SetAutoDetach(true)

	config, err := dev.Config(2)
	check(err)
	defer config.Close()

	intf, err := config.Interface(1, 0)
	check(err)
	defer intf.Close()

	iep, err := intf.InEndpoint(1)
	check(err)

	oep, err := intf.OutEndpoint(1)
	check(err)

	for {
		in := readUSB(iep)

		request := identifyRequest(in, len(filename))

		fmt.Println(request)
		var data []byte
		if request == "BOOTP" {
			data = processBOOTP(in)
		} else if request == "ARP" {
			data = processARP(in)
		} else if request == "TFTP" {
			data = processTFTP(in)
		} else if request == "TFTP_Data" {
			data = processTFTPData(in)
			if string(data) == "" {
				return true
			}
		}
		if string(data) != "" {

			sendUSB(oep, data)
		}
	}
}

func main() {

	//sigs := make(chan os.Signal, 1)

	ctx = gousb.NewContext()
	defer ctx.Close()

	for onAttach(ctx) != "0451 6141" {
	}

	for sendSPL() == false {
		time.Sleep(time.Millisecond * 500)
	}

	fmt.Println("I can haz uboot")

	for findSPL() == false {
		time.Sleep(time.Second)
	}
	fmt.Println("oke now")

}
