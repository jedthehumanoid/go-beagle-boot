package main

import (
	"testing"
)

func TestIdentifyRequst(t *testing.T) {
	if identifyRequest([]byte{0, 0, 0, 0, 0xC2}, 0) != "BOOTP" {
		t.Error("Expected BOOTP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0x6C}, 0) != "BOOTP" {
		t.Error("Expected BOOTP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0x56}, 0) != "ARP" {
		t.Error("Expected ARP")
	}
	if identifyRequest([]byte{0, 0, 0, 0, 0}, 0) != "notIdentified" {
		t.Error("Expected notIdentified")
	}

}
