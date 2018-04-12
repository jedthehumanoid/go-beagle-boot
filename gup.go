package main

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

const romvid gousb.ID = 0x0451
const rompid gousb.ID = 0x6141

var ctx *gousb.Context

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func identifyRequest(buf []byte, length int) string {

	val := buf[4]

	if val == 0xc2 || val == 0x6c {
		return "BOOTP"
	}
	if val == 0x56 {
		return "ARP"
	}

	if int(val) == (0x5f+length) || int(val) == (0x76+length) {
		return "TFTP"
	}

	if val == 0x5a {
		return "TFTP_Data"
	}

	return "notIdentified"
}

func processBOOTP(data []byte) {
	fmt.Println("processBOOTP")
}
func sendSPL() bool {

	dev, err := ctx.OpenDeviceWithVIDPID(romvid, rompid)
	if dev == nil {
		return false
	}
	fmt.Println("Sending SPL...")
	check(err)
	defer dev.Close()

	dev.SetAutoDetach(true)

	config, err := dev.Config(1)
	check(err)
	intf, err := config.Interface(1, 0)
	check(err)
	/*
		intf, done, err := dev.DefaultInterface()
		if err != nil {
			log.Fatalf("%s.DefaultInterface(): %v", dev, err)
		}
		defer done()
	*/

	iep, err := intf.InEndpoint(1)
	check(err)

	_, err = intf.OutEndpoint(2)
	check(err)

	buf := make([]byte, 10*iep.Desc.MaxPacketSize)
	_, err = iep.Read(buf)
	check(err)

	request := identifyRequest(buf, 0)
	if request == "BOOTP" {
		processBOOTP(buf)
	}
	time.Sleep(time.Second * 5)
	return true
}

func main() {

	ctx = gousb.NewContext()
	defer ctx.Close()
	for sendSPL() == false {
		time.Sleep(time.Second)
	}

}
