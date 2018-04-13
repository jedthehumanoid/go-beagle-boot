package main

import (
	"bytes"
	"encoding/binary"
)

func parseEtherHeader(data [14]byte) etherHeader {
	var etherheader etherHeader
	buf := bytes.NewReader(data[:])
	err := binary.Read(buf, binary.BigEndian, &etherheader)
	check(err)

	return etherheader
}

func parseUdpPacket(data [8]byte) udpPacket {
	var ret udpPacket
	buf := bytes.NewReader(data[:])
	err := binary.Read(buf, binary.BigEndian, &ret)
	check(err)

	return ret
}

func makeRndis(length uint32) rndisPacket {
	var rndis rndisPacket

	rndis.MsgType = 1
	rndis.MsgLength = length + 44
	rndis.DataOffset = 0x24
	rndis.DataLength = length

	return rndis

}

func identifyRequest(buf []byte, length int) string {
	val := int(buf[4])
	switch val {
	case 0xc2, 0x6c:
		return "BOOTP"
	case 0x56:
		return "ARP"
	case 0x5f + length, 0x76 + length:
		return "TFTP"
	case 0x5a:
		return "TFTP_Data"
	default:
		return "notIdentified"
	}
}
