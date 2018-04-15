package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func parseEtherHeader(data []byte) etherHeader {
	var etherheader etherHeader
	buf := bytes.NewReader(data[:])
	err := binary.Read(buf, binary.BigEndian, &etherheader)
	check(err)

	return etherheader
}

func parseUdpPacket(data []byte) udpPacket {
	var ret udpPacket
	buf := bytes.NewReader(data[:])
	err := binary.Read(buf, binary.BigEndian, &ret)
	check(err)

	return ret
}

func parseBootpPacket(data []byte) bootpPacket {
	var ret bootpPacket
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

func calculateChecksum(bytes []byte) uint16 {
	var sum uint32

	for i := range bytes {
		if i%2 != 0 {
			continue
		}
		a := uint16(bytes[i])
		b := uint16(bytes[i+1])
		sum += uint32(b + (a << 8))
	}

	for sum > 0xffff {
		carry := sum >> 16
		sum = sum & 0xffff
		sum += carry
	}
	sum = ^sum
	return uint16(sum)
}

func makeIpv4Packet(sourceAddr [4]byte, destAddr [4]byte, length uint16, id uint16, protocol uint8) ipv4Packet {
	var ret ipv4Packet

	ret.VerHl = 69
	ret.TotalLength = length
	ret.ID = id
	ret.TTL = 64
	ret.Protocol = protocol
	ret.SourceAddr = sourceAddr
	ret.DestAddr = destAddr

	// Checksum calculation

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, ret)
	check(err)

	fmt.Printf("%+v", ret)

	ret.ChkSum = calculateChecksum(buf.Bytes())

	return ret
}

func makeBootpPacket(servername string, id uint32, source [16]byte, dest [4]byte, server [4]byte, bootfile string) bootpPacket {
	var ret bootpPacket
	ret.Opcode = 2
	ret.Hw = 1
	ret.HwLength = 6

	ret.Hwaddr = source
	ret.Xid = id
	ret.Yiaddr = dest
	ret.ServerIP = server
	ret.BootpGWIp = server

	ret.Vendor = [64]byte{99, 130, 83, 99, 1, 4, 255, 255, 255, 0, 3, 4, 192, 168, 1, 9, 0xFF}

	copy(ret.ServerName[:], servername)
	copy(ret.BootFile[:], bootfile)

	return ret

}

func identifyRequest(buf []byte, length int) string {
	switch int(buf[4]) {
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
