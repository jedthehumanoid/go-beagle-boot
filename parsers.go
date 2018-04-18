package main

import (
	"bytes"
	"encoding/binary"
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

func processBOOTP(data []byte, filename string) []byte {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
		Bootp incompleteBootpMessage
	}

	inbuf := bytes.NewReader(data)
	binRead(inbuf, binary.BigEndian, &req)
	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	binRead(inbuf, binary.LittleEndian, &req.Rndis)

	rndisResp := makeRndis(fullSize - rndisSize)
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x800}
	ipResp := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+bootpSize, 0, ipUDP)
	udpResp := makeUdpHeader(req.Udp.Dest, req.Udp.Source, bootpSize)
	bootpResp := makeBootpPacket("BEAGLEBOOT", req.Bootp.Xid,
		req.Ether.Source, bbIP, serverIP, filename)

	buf := new(bytes.Buffer)
	binWrite(buf, binary.LittleEndian, rndisResp)
	binWrite(buf, binary.BigEndian, etherResp)
	binWrite(buf, binary.BigEndian, ipResp)
	binWrite(buf, binary.BigEndian, udpResp)
	binWrite(buf, binary.BigEndian, bootpResp)

	return buf.Bytes()
}

func processARP(data []byte) []byte {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Arp   arpMessage
	}

	inbuf := bytes.NewReader(data)
	binRead(inbuf, binary.BigEndian, &req)
	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	binRead(inbuf, binary.LittleEndian, &req.Rndis)

	arp := makeARPMessage(2, serverHwaddr, req.Arp.TargetProtocolAddr,
		req.Arp.SenderHardwareAddr, req.Arp.SenderProtocolAddr)
	rndisResp := makeRndis(etherSize + arpSize)
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x806}

	buf := new(bytes.Buffer)
	binWrite(buf, binary.LittleEndian, rndisResp)
	binWrite(buf, binary.BigEndian, etherResp)
	binWrite(buf, binary.BigEndian, arp)

	return buf.Bytes()
}

func processTFTP(data []byte, filename string) []byte {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
	}

	var blocksize uint16 = 512

	inbuf := bytes.NewReader(data)
	binRead(inbuf, binary.BigEndian, &req)
	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	binRead(inbuf, binary.LittleEndian, &req.Rndis)

	dat, err := ioutil.ReadFile(filename)
	check(err)

	rndis := makeRndis(etherSize + ipSize + udpSize + tftpSize + uint32(blocksize))
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x800}
	ip := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+tftpSize+blocksize, 0, ipUDP)
	udpResp := makeUdpHeader(req.Udp.Dest, req.Udp.Source, tftpSize+blocksize)
	tftp := tftpData{3, 1}
	filedata := dat[:blocksize]

	buf := new(bytes.Buffer)
	binWrite(buf, binary.LittleEndian, rndis)
	binWrite(buf, binary.BigEndian, etherResp)
	binWrite(buf, binary.BigEndian, ip)
	binWrite(buf, binary.BigEndian, udpResp)
	binWrite(buf, binary.BigEndian, tftp)
	binWrite(buf, binary.BigEndian, filedata)

	return buf.Bytes()
}

func processTFTPData(data []byte, filename string) []byte {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
		Tftp  tftpData
	}

	var blocksize uint16 = 512

	inbuf := bytes.NewReader(data)
	binRead(inbuf, binary.BigEndian, &req)
	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	binRead(inbuf, binary.LittleEndian, &req.Rndis)

	dat, err := ioutil.ReadFile(filename)
	check(err)
	blocks := uint16(math.Ceil(float64(len(dat)) / float64(blocksize)))

	bn := req.Tftp.BlockNumber + 1
	if bn == blocks { // Last block
		blocksize = uint16(len(dat[(bn-1)*blocksize:]))
	} else if bn == blocks+1 { //Finished
		return []byte("")
	}

	rndis := makeRndis(etherSize + ipSize + udpSize + tftpSize + uint32(blocksize))
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x800}
	ip := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+tftpSize+blocksize, 0, ipUDP)
	udpResp := makeUdpHeader(req.Udp.Dest, req.Udp.Source, tftpSize+blocksize)
	tftp := tftpData{3, bn}

	start := (uint64(bn) - 1) * uint64(512)

	filedata := dat[start : start+uint64(blocksize)]

	buf := new(bytes.Buffer)
	binWrite(buf, binary.LittleEndian, rndis)
	binWrite(buf, binary.BigEndian, etherResp)
	binWrite(buf, binary.BigEndian, ip)
	binWrite(buf, binary.BigEndian, udpResp)
	binWrite(buf, binary.BigEndian, tftp)
	binWrite(buf, binary.BigEndian, filedata)

	return buf.Bytes()
}
