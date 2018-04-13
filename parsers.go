package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

func write(buf io.Writer, order binary.ByteOrder, data interface{}) {
	err := binary.Write(buf, order, data)
	check(err)
}

func parseEtherHeader(data [14]byte) ether_header {
	var ret ether_header
	for i, _ := range ret.h_dest {
		ret.h_dest[i] = data[i]
	}

	for i, _ := range ret.h_dest {
		ret.h_source[i] = data[i+6]
	}

	var h_proto uint16
	buf := bytes.NewReader(data[12:])
	err := binary.Read(buf, binary.BigEndian, &h_proto)
	check(err)
	ret.h_proto = h_proto

	return ret
}

func parseUdpPacket(data [8]byte) udp_packet {
	var ret udp_packet
	var val uint16

	buf := bytes.NewReader(data[:2])
	err := binary.Read(buf, binary.BigEndian, &val)
	check(err)
	ret.udpSrc = val

	buf = bytes.NewReader(data[2:4])
	err = binary.Read(buf, binary.BigEndian, &val)
	check(err)
	ret.udpDest = val

	buf = bytes.NewReader(data[4:6])
	err = binary.Read(buf, binary.BigEndian, &val)
	check(err)
	ret.udpLen = val

	buf = bytes.NewReader(data[6:8])
	err = binary.Read(buf, binary.BigEndian, &val)
	check(err)
	ret.chkSum = val

	return ret
}

func makeRndis(length int) []byte {
	buf := new(bytes.Buffer)

	var msg_type uint32 = 1
	var msg_len uint32 = uint32(length + 44)
	var data_offset uint32 = 0x24
	var data_length uint32 = uint32(length)

	write(buf, binary.LittleEndian, msg_type)
	write(buf, binary.LittleEndian, msg_len)
	write(buf, binary.LittleEndian, data_offset)
	write(buf, binary.LittleEndian, data_length)
	write(buf, binary.LittleEndian, uint32(0))
	write(buf, binary.LittleEndian, uint32(0))
	write(buf, binary.LittleEndian, uint32(0))
	write(buf, binary.LittleEndian, uint32(0))
	write(buf, binary.LittleEndian, uint32(0))
	write(buf, binary.LittleEndian, uint32(0))
	write(buf, binary.LittleEndian, uint32(0))

	return buf.Bytes()

}

func makeEther(source [6]byte, dest [6]byte, proto uint16) []byte {
	buf := new(bytes.Buffer)

	write(buf, binary.LittleEndian, source)
	write(buf, binary.LittleEndian, dest)
	write(buf, binary.LittleEndian, proto)

	return buf.Bytes()
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
