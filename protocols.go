package main

type etherHeader struct {
	Dest   [6]byte
	Source [6]byte
	Proto  uint16
}

type udpHeader struct {
	Source uint16
	Dest   uint16
	Length uint16
	ChkSum uint16
}

type tftpData struct {
	Opcode      uint16
	BlockNumber uint16
}
