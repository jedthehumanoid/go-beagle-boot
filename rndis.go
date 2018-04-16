package main

type rndisMessage struct {
	MsgType          uint32
	MsgLength        uint32
	DataOffset       uint32
	DataLength       uint32
	BandOffset       uint32
	BandLen          uint32
	OutBandElements  uint32
	PacketOffset     uint32
	PacketInfoLength uint32
	ReservedFirst    uint32
	ReservedSecond   uint32
}

func makeRndis(length uint32) rndisMessage {
	var rndis rndisMessage

	rndis.MsgType = 1
	rndis.MsgLength = length + 44
	rndis.DataOffset = 0x24
	rndis.DataLength = length

	return rndis
}
