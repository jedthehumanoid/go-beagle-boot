package main

import (
	"bytes"
	"encoding/binary"
)

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

type ipv4Datagram struct {
	VersionHL    uint8
	Tos          uint8
	TotalLength  uint16
	ID           uint16
	FlagsFragOff uint16
	TTL          uint8
	Protocol     uint8
	ChkSum       uint16
	SourceAddr   [4]byte
	DestAddr     [4]byte
}

func makeIpv4Packet(sourceAddr [4]byte, destAddr [4]byte, length uint16, id uint16, protocol uint8) ipv4Datagram {
	var ret ipv4Datagram

	ret.VersionHL = 69
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
	ret.ChkSum = calculateChecksum(buf.Bytes())

	return ret
}
