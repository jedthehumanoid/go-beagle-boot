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

func makeUdpHeader(source uint16, dest uint16, length uint16) udpHeader {
	var ret udpHeader

	ret.Source = source
	ret.Dest = dest
	ret.Length = length + 8

	return ret
}

type tftpData struct {
	Opcode      uint16
	BlockNumber uint16
}
