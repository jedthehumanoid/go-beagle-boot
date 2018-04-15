package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"

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

var ctx *gousb.Context

type etherHeader struct {
	Dest   [6]byte
	Source [6]byte
	Proto  uint16
}

type udpPacket struct {
	Src    uint16
	Dest   uint16
	Len    uint16
	ChkSum uint16
}

type rndisPacket struct {
	MsgType          uint32
	MsgLength        uint32
	DataOffset       uint32
	DataLength       uint32
	BandOffset       uint32
	BandLen          uint32
	OutBandElements  uint32
	PacketOffset     uint32
	PacketInfoLength uint32
	ReservedFirst    uint32
	ReservedSecond   uint32
}

type ipv4Packet struct {
	VerHl       uint8
	Tos         uint8
	TotalLength uint16
	ID          uint16
	FragOff     uint16
	TTL         uint8
	Protocol    uint8
	ChkSum      uint16
	SourceAddr  [4]byte
	DestAddr    [4]byte
}

type bootpPacket struct {
	Opcode     uint8
	Hw         uint8
	HwLength   uint8
	Hopcount   uint8
	Xid        uint32
	Secs       uint16
	Flags      uint16
	Ciaddr     [4]byte
	Yiaddr     [4]byte
	ServerIP   [4]byte
	BootpGWIp  [4]byte
	Hwaddr     [16]byte
	ServerName [64]byte
	BootFile   [128]byte
	Vendor     [64]byte
}

type arpMessage struct {
	HardwareType       uint16
	ProtocolType       uint16
	HardwareAddrLength uint8
	ProtocolAddrLength uint8
	Opcode             uint16
}

func processBOOTP(data []byte) []byte {
	fmt.Println("processBOOTP")

	var request struct {
		Rndis rndisPacket
		Ether etherHeader
		Ipv4  ipv4Packet
		Udp   udpPacket
		Bootp bootpPacket
	}

	fmt.Println(unsafe.Sizeof(request))

	fmt.Println(unsafe.Sizeof(request.Rndis))
	fmt.Println(unsafe.Sizeof(request.Ether))
	fmt.Println(unsafe.Sizeof(request.Ipv4))
	fmt.Println(unsafe.Sizeof(request.Udp))
	fmt.Println(unsafe.Sizeof(request.Bootp))

	udpPos := rndisSize + etherSize + ipSize
	bootpPos := udpPos + udpSize

	etherheader := parseEtherHeader(data[rndisSize : rndisSize+etherSize])
	udppacket := parseUdpPacket(data[udpPos : udpPos+udpSize])
	bootp := parseBootpPacket(data[bootpPos : bootpPos+bootpSize])

	rndisResp := makeRndis(fullSize - rndisSize)
	etherResp := etherHeader{etherheader.Source, serverHwaddr, 0x800}
	ipResp := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+bootpSize, 0, ipUDP)
	udpResp := udpPacket{udppacket.Dest, udppacket.Src, bootpSize + 8, 0}

	var ehs [16]byte
	copy(ehs[:], etherheader.Source[:])
	bootpResp := makeBootpPacket("BEAGLEBOOT", bootp.Xid, ehs, bbIP, serverIP, "spl")

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, rndisResp)
	check(err)

	packets := struct {
		etherHeader
		ipv4Packet
		udpPacket
		bootpPacket
	}{etherResp, ipResp, udpResp, bootpResp}
	err = binary.Write(buf, binary.BigEndian, packets)
	check(err)

	return buf.Bytes()
}

func processARP(data []byte) []byte {
	var request struct {
		Rndis rndisPacket
		Ether etherHeader
		Arp   arpMessage
	}

	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &request)
	check(err)

	// Replace rndis separately because little endian
	buf = bytes.NewReader(data)
	err = binary.Read(buf, binary.LittleEndian, &request.Rndis)
	check(err)

	fmt.Println(request.Rndis, request.Ether, request.Arp)
	return []byte("")
}

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

		request := identifyRequest(in, 0)

		fmt.Println(request)

		if request == "BOOTP" {
			data := processBOOTP(in)
			sendUSB(oep, data)
		} else if request == "ARP" {
			data := processARP(in)
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
