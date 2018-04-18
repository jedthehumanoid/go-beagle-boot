package main

import (
	"fmt"

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

var debug = false

var serverHwaddr = [6]byte{0x9a, 0x1f, 0x85, 0x1c, 0x3d, 0x0e}
var serverIP = [4]byte{192, 168, 1, 9}
var bbIP = [4]byte{192, 168, 1, 3}

const maxbuf = 450

var ctx *gousb.Context

func sendSPL() bool {

	dev, err := ctx.OpenDeviceWithVIDPID(romvid, rompid)
	if dev == nil {
		return false
	}
	check(err)
	defer dev.Close()

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

	transfer(iep, oep, "spl")
	return true
}

func findSPL() bool {

	dev, err := ctx.OpenDeviceWithVIDPID(splvid, splpid)
	if dev == nil {
		return false
	}
	check(err)
	defer dev.Close()

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

	transfer(iep, oep, "uboot")
	return true
}

func transfer(in *gousb.InEndpoint, out *gousb.OutEndpoint, filename string) {
	for {
		in := readUSB(in)
		request := identifyRequest(in, len(filename))
		var data []byte

		if request == "BOOTP" {
			fmt.Print("bootp")
			data = processBOOTP(in, filename)
		} else if request == "ARP" {
			fmt.Print(", arp")
			data = processARP(in)
		} else if request == "TFTP" {
			fmt.Println(", tftp\n")
			data = processTFTP(in, filename)
		} else if request == "TFTP_Data" {
			fmt.Print(".")
			data = processTFTPData(in, filename)
			if string(data) == "" {
				fmt.Print("\n")
				return
			}
		}
		if string(data) != "" {

			sendUSB(out, data)
		}
	}
}

func main() {
	ctx = gousb.NewContext()
	defer ctx.Close()

	fmt.Println("Connect Beaglebone")

	for onAttach(ctx) != "0451 6141" {
	}
	fmt.Println("Found Beaglebone in ROM mode, sending SPL")

	sendSPL()

	for onAttach(ctx) != "0525 a4a2" {
	}
	fmt.Println("Found Beaglebone in SPL mode, sending UBOOT")

	findSPL()
	fmt.Println("Done!")

}
