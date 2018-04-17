package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
)

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

	fmt.Println("----------")
	fmt.Println(request.Ipv4.ID, request.Ipv4.ID+1024)

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

func processBOOTP2(data []byte) []byte {
	var request struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
		Bootp specialbootpMessage
	}

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &request)
	check(err)

	// Replace rndis separately because little endian
	inbuf = bytes.NewReader(data)
	err = binary.Read(inbuf, binary.LittleEndian, &request.Rndis)
	check(err)

	fmt.Printf("rndis: %+v\n", request.Rndis)
	fmt.Printf("ether: %+v\n", request.Ether)
	fmt.Printf("ipv4: %+v\n", request.Ipv4)
	fmt.Printf("udp: %+v\n", request.Udp)
	fmt.Printf("bootp: %+v\n", request.Bootp)

	fmt.Printf("-%d- % x", len(data[86:]), data[86:])

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

	return []byte("")
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

	arp := makeARPMessage(2, serverHwaddr, request.Arp.TargetProtocolAddr,
		request.Arp.SenderHardwareAddr, request.Arp.SenderProtocolAddr)
	rndisResp := makeRndis(etherSize + arpSize)
	etherResp := etherHeader{request.Ether.Source, serverHwaddr, 0x806}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndisResp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, etherResp)
	check(err)
	err = binary.Write(buf, binary.BigEndian, arp)
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

	dat, err := ioutil.ReadFile(filename)
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

	dat, err := ioutil.ReadFile(filename)
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
