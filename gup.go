package main

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

const romvid gousb.ID = 0x0451
const rompid gousb.ID = 0x6141

const rndisSize = 44
const etherSize = 14
const arp_Size = 28
const ipSize = 20
const udpSize = 8
const bootpSize = 300
const tftpSize = 4
const fullSize = 386

const maxbuf = 450

var ctx *gousb.Context

type ether_header struct {
	h_dest   [6]byte
	h_source [6]byte
	h_proto  uint16
}

type udp_packet struct {
	udpSrc  uint16
	udpDest uint16
	udpLen  uint16
	chkSum  uint16
}

func processBOOTP(data []byte) {
	fmt.Println("processBOOTP")

	var ether_buf [etherSize]byte
	var udp_buf [udpSize]byte

	fmt.Println(data)

	copy(ether_buf[:], data[rndisSize:])
	copy(udp_buf[:], data[rndisSize+etherSize+ipSize:])

	etherheader := parseEtherHeader(ether_buf)
	udppacket := parseUdpPacket(udp_buf)

	rndis := makeRndis(fullSize - rndisSize)

	fmt.Println(etherheader)
	fmt.Println(udppacket)

	fmt.Println(rndis)

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
