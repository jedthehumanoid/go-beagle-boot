package main

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"math"
	"os"
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

func processBOOTP(data []byte, filename string) ([]byte, error) {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
		Bootp incompleteBootpMessage
	}

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &req)
	if err != nil {
		return []byte{}, err
	}
	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	err = binary.Read(inbuf, binary.LittleEndian, &req.Rndis)
	if err != nil {
		return []byte{}, err
	}

	rndisResp := makeRndis(fullSize - rndisSize)
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x800}
	ipResp := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+bootpSize, 0, ipUDP)
	udpResp := makeUdpHeader(req.Udp.Dest, req.Udp.Source, bootpSize)
	bootpResp := makeBootpPacket("BEAGLEBOOT", req.Bootp.Xid,
		req.Ether.Source, bbIP, serverIP, filename)

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndisResp)
	if err != nil {
		return []byte{}, err
	}
	err = writeMulti(buf, binary.BigEndian, []interface{}{
		etherResp,
		ipResp,
		udpResp,
		bootpResp})
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func processARP(data []byte) ([]byte, error) {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Arp   arpMessage
	}

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &req)
	if err != nil {
		return []byte{}, err
	}

	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	err = binary.Read(inbuf, binary.LittleEndian, &req.Rndis)
	if err != nil {
		return []byte{}, err
	}

	arp := makeARPMessage(2, serverHwaddr, req.Arp.TargetProtocolAddr,
		req.Arp.SenderHardwareAddr, req.Arp.SenderProtocolAddr)
	rndisResp := makeRndis(etherSize + arpSize)
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x806}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndisResp)
	if err != nil {
		return []byte{}, err
	}

	err = writeMulti(buf, binary.BigEndian, []interface{}{
		etherResp,
		arp})
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func processTFTP(data []byte, filename string) ([]byte, error) {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
	}

	var blocksize uint16 = 512

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &req)
	if err != nil {
		return []byte{}, err
	}

	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	err = binary.Read(inbuf, binary.LittleEndian, &req.Rndis)
	if err != nil {
		return []byte{}, err
	}

	dat, err := ioutil.ReadFile(binPath + string(os.PathSeparator) + filename)
	check(err)

	rndis := makeRndis(etherSize + ipSize + udpSize + tftpSize + uint32(blocksize))
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x800}
	ip := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+tftpSize+blocksize, 0, ipUDP)
	udpResp := makeUdpHeader(req.Udp.Dest, req.Udp.Source, tftpSize+blocksize)
	tftp := tftpData{3, 1}
	filedata := dat[:blocksize]

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndis)
	if err != nil {
		return []byte{}, err
	}

	err = writeMulti(buf, binary.BigEndian, []interface{}{
		etherResp,
		ip,
		udpResp,
		tftp,
		filedata})
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func processTFTPData(data []byte, filename string) ([]byte, error) {
	var req struct {
		Rndis rndisMessage
		Ether etherHeader
		Ipv4  ipv4Datagram
		Udp   udpHeader
		Tftp  tftpData
	}

	var blocksize uint16 = 512

	inbuf := bytes.NewReader(data)
	err := binary.Read(inbuf, binary.BigEndian, &req)
	if err != nil {
		return []byte{}, err
	}

	inbuf = bytes.NewReader(data) // Reset and read rndis again, in little endian
	err = binary.Read(inbuf, binary.LittleEndian, &req.Rndis)
	if err != nil {
		return []byte{}, err
	}

	dat, err := ioutil.ReadFile(binPath + string(os.PathSeparator) + filename)
	check(err)
	blocks := uint16(math.Ceil(float64(len(dat)) / float64(blocksize)))

	bn := req.Tftp.BlockNumber + 1
	if bn == blocks { // Last block
		blocksize = uint16(len(dat[(bn-1)*blocksize:]))
	} else if bn == blocks+1 { //Finished
		return []byte{}, nil
	}

	rndis := makeRndis(etherSize + ipSize + udpSize + tftpSize + uint32(blocksize))
	etherResp := etherHeader{req.Ether.Source, serverHwaddr, 0x800}
	ip := makeIpv4Packet(serverIP, bbIP, ipSize+udpSize+tftpSize+blocksize, 0, ipUDP)
	udpResp := makeUdpHeader(req.Udp.Dest, req.Udp.Source, tftpSize+blocksize)
	tftp := tftpData{3, bn}

	start := (uint64(bn) - 1) * uint64(512)

	filedata := dat[start : start+uint64(blocksize)]

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndis)
	if err != nil {
		return []byte{}, err
	}

	err = writeMulti(buf, binary.BigEndian, []interface{}{
		etherResp,
		ip,
		udpResp,
		tftp,
		filedata})
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}
