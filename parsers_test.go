package main

import (
	"fmt"
	"testing"
)

func TestParseEtherHeader(t *testing.T) {

	resp := parseEtherHeader([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14})

	if !byteSliceEquals(resp.Dest[:], []byte{1, 2, 3, 4, 5, 6}) {
		t.Error("Expected {1,2,3,4,5,6}")
	}
	if !byteSliceEquals(resp.Source[:], []byte{7, 8, 9, 10, 11, 12}) {
		t.Error("Expected {7,8,9,10,11,12}")
	}

	if resp.Proto != 3342 {
		t.Error("Got ", resp.Proto)
	}
}

func testParseUdpPacket(t *testing.T) {

	resp := parseUdpPacket([]byte{1, 2, 3, 4, 5, 6, 7, 8})

	fmt.Println(resp)

}

func TestMakeIpv4Packet(t *testing.T) {
	p := makeIpv4Packet([4]byte{1, 2, 3, 4}, [4]byte{5, 6, 7, 8}, 20, 30, 40)
	fmt.Printf("%+v", p)
	if p.VerHl != 69 || p.TotalLength != 20 || p.ID != 30 || p.ChkSum != 27281 ||
		p.SourceAddr[0] != 1 || p.DestAddr[0] != 5 {
		t.Error()
	}

}

func TestMakeRndis(t *testing.T) {
	rndis := makeRndis(342)

	if rndis.MsgType != 1 {
		t.Error("Wrong type")
	}
	if rndis.MsgLength != 386 {
		t.Error("Wrong message length")
	}
	if rndis.DataOffset != 0x24 {
		t.Error("Wrong data offset")
	}
	if rndis.DataLength != 342 {
		t.Error("Wrong data length")
	}
}

func TestIdentifyRequest(t *testing.T) {
	if identifyRequest([]byte{0, 0, 0, 0, 0xC2}, 0) != "BOOTP" {
		t.Error("Expected BOOTP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0x6C}, 0) != "BOOTP" {
		t.Error("Expected BOOTP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0x56}, 0) != "ARP" {
		t.Error("Expected ARP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0x60}, 1) != "TFTP" {
		t.Error("Expected TFTP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0x77}, 1) != "TFTP" {
		t.Error("Expected TFTP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0x5a}, 0) != "TFTP_Data" {
		t.Error("Expected TFTP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0}, 0) != "notIdentified" {
		t.Error("Expected notIdentified")
	}
}

func TestCalculateCheksum(t *testing.T) {
	sum := calculateChecksum([]byte{0x45, 0x00, 0x00, 0x73, 0x00, 0x00, 0x40, 0x00,
		0x40, 0x11, 0x00, 0x00, 0xc0, 0xa8, 0x00, 0x01, 0xc0, 0xa8, 0x00, 0xc7})
	if sum != 47201 {
		t.Error("Wrong checksum, got", sum)
	}
}
func TestCalculateCheksum2(t *testing.T) {
	sum := calculateChecksum([]byte{0x45, 0x00, 0x00, 0x3c, 0x1c, 0x46, 0x40, 0x00,
		0x40, 0x06, 0x00, 0x00, 0xac, 0x10, 0x0a, 0x63, 0xac, 0x10, 0x0a, 0x0c})
	if sum != 45542 {
		t.Error("Wrong checksum, got", sum)
	}
}
