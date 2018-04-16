package main

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

const romvid gousb.ID = 0x0451
const rompid gousb.ID = 0x6141
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
	bytesread, err := ep.Read(buf)
	check(err)
	buf = buf[:bytesread]
	fmt.Printf("Receiving: --%d-- % x\n", len(buf), buf)
	return buf
}

func sendUSB(ep *gousb.OutEndpoint, data []byte) {
	_, err := ep.Write(data)
	check(err)
	fmt.Printf("Sending: --%d-- % x\n", len(data), data)
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

		if request == "BOOTP" {
			data := processBOOTP(in)
			sendUSB(oep, data)
		} else if request == "ARP" {
			data := processARP(in)
			sendUSB(oep, data)
		} else if request == "TFTP" {
			data := processTFTP(in)
			fmt.Println(data)
		}
	}

	return true
}

func main() {

	ctx = gousb.NewContext()
	defer ctx.Close()
	for sendSPL() == false {
		time.Sleep(time.Second)
	}

}
