package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
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

func processBOOTP(data []byte) []byte {
	var request struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
		Bootp bootpMessage
	}

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &request)
	check(err)

	// Replace rndis separately because little endian
	inbuf = bytes.NewReader(data)
	err = binary.Read(inbuf, binary.LittleEndian, &request.Rndis)
	check(err)

	rndisResp := makeRndis(fullSize - rndisSize)
	etherResp := etherHeader{request.Ether.Source, serverHwaddr, 0x800}
	ipResp := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+bootpSize, 0, ipUDP)
	udpResp := udpHeader{request.Udp.Dest, request.Udp.Source, bootpSize + 8, 0}

	bootpResp := makeBootpPacket("BEAGLEBOOT", request.Bootp.Xid,
		request.Ether.Source, bbIP, serverIP, filename)

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndisResp)
	check(err)

	packets := struct {
		etherHeader
		ipv4Datagram
		udpHeader
		bootpMessage
	}{etherResp, ipResp, udpResp, bootpResp}
	err = binary.Write(buf, binary.BigEndian, packets)
	check(err)

	return buf.Bytes()
}

func processARP(data []byte) []byte {
	var request struct {
		Rndis rndisMessage
		Ether etherHeader
		Arp   arpMessage
	}

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &request)
	check(err)

	// Replace rndis separately because little endian
	inbuf = bytes.NewReader(data)
	err = binary.Read(inbuf, binary.LittleEndian, &request.Rndis)
	check(err)

	arp := makeARPMessage(2, serverHwaddr, request.Arp.TargetProtocolAddr, request.Arp.SenderHardwareAddr, request.Arp.SenderProtocolAddr)
	rndisResp := makeRndis(etherSize + arpSize)
	etherResp := etherHeader{request.Ether.Source, serverHwaddr, 0x806}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndisResp)
	check(err)

	packets := struct {
		etherHeader
		arpMessage
	}{etherResp, arp}
	err = binary.Write(buf, binary.BigEndian, packets)
	check(err)

	return buf.Bytes()
}

func processTFTP(data []byte) []byte {
	var request struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
	}

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &request)
	check(err)

	// Replace rndis separately because little endian
	inbuf = bytes.NewReader(data)
	err = binary.Read(inbuf, binary.LittleEndian, &request.Rndis)
	check(err)

	dat, err := ioutil.ReadFile("spl")
	check(err)
	blocks := math.Ceil(float64(len(dat)) / 512.0)
	fmt.Println("Blocks, ", blocks)

	rndis := makeRndis(etherSize + ipSize + udpSize + tftpSize + 512)
	etherResp := etherHeader{request.Ether.Source, serverHwaddr, 0x800}
	ip := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+tftpSize+512, 0, ipUDP)
	udpResp := udpHeader{request.Udp.Dest, request.Udp.Source, tftpSize + 512 + 8, 0}
	tftp := tftpData{3, 1}
	filedata := dat[:512]

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndis)
	check(err)

	err = binary.Write(buf, binary.BigEndian, etherResp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, ip)
	check(err)
	err = binary.Write(buf, binary.BigEndian, udpResp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, tftp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, filedata)
	check(err)

	//	fmt.Println(rndis, etherResp, ip, udpResp, tftp)
	return buf.Bytes()
}

func processTFTPData(data []byte) []byte {
	var request struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
		Tftp  tftpData
	}

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &request)
	check(err)

	// Replace rndis separately because little endian
	inbuf = bytes.NewReader(data)
	err = binary.Read(inbuf, binary.LittleEndian, &request.Rndis)
	check(err)

	dat, err := ioutil.ReadFile("spl")
	check(err)
	blocks := uint64(math.Ceil(float64(len(dat)) / 512.0))

	bn := uint64(request.Tftp.BlockNumber + 1)
	fmt.Println(request.Tftp.BlockNumber, blocks)
	blocksize := 512
	if bn == blocks {
		blocksize = len(dat[(bn-1)*512:])
	}
	if request.Tftp.BlockNumber == uint16(blocks) {
		return []byte("")
	}

	rndis := makeRndis(etherSize + ipSize + udpSize + tftpSize + uint32(blocksize))
	etherResp := etherHeader{request.Ether.Source, serverHwaddr, 0x800}
	ip := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+tftpSize+uint16(blocksize), 0, ipUDP)
	udpResp := udpHeader{request.Udp.Dest, request.Udp.Source, tftpSize + uint16(blocksize) + 8, 0}
	tftp := tftpData{3, request.Tftp.BlockNumber + 1}

	start := (bn - 1) * 512

	filedata := dat[start : start+uint64(blocksize)]

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndis)
	check(err)

	err = binary.Write(buf, binary.BigEndian, etherResp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, ip)
	check(err)
	err = binary.Write(buf, binary.BigEndian, udpResp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, tftp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, filedata)
	check(err)

	//	fmt.Println(rndis, etherResp, ip, udpResp, tftp)
	return buf.Bytes()
}
