package main

import (
	"fmt"
	"testing"
)

func TestParseEtherHeader(t *testing.T) {

	resp := parseEtherHeader([14]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14})

	if !byteSliceEquals(resp.h_dest[:], []byte{1, 2, 3, 4, 5, 6}) {
		t.Error("Expected {1,2,3,4,5,6}")
	}
	if !byteSliceEquals(resp.h_source[:], []byte{7, 8, 9, 10, 11, 12}) {
		t.Error("Expected {7,8,9,10,11,12}")
	}

	if resp.h_proto != 3342 {
		t.Error("Expected ", resp.h_proto)
	}
}

func TestParseUdpPacket(t *testing.T) {

	resp := parseUdpPacket([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	fmt.Println(resp)

}

func TestMakeRndis(t *testing.T) {
	ret := toHexString(makeRndis(342))

	if ret != "01 00 00 00 82 01 00 00 24 00 00 00 56 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 " {
		t.Error("wrong")
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
