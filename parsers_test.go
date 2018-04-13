package main

import (
	"fmt"
	"testing"
)

func TestParseEtherHeader(t *testing.T) {

	resp := parseEtherHeader([14]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14})

	if !byteSliceEquals(resp.Dest[:], []byte{1, 2, 3, 4, 5, 6}) {
		t.Error("Expected {1,2,3,4,5,6}")
	}
	if !byteSliceEquals(resp.Source[:], []byte{7, 8, 9, 10, 11, 12}) {
		t.Error("Expected {7,8,9,10,11,12}")
	}

	if resp.Proto != 3342 {
		t.Error("Expected ", resp.Proto)
	}
}

func TestParseUdpPacket(t *testing.T) {

	resp := parseUdpPacket([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	fmt.Println(resp)

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
