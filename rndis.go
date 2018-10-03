package main

import (
	"bytes"
	"encoding/binary"
)

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

type rndisInitMsg struct {
	MsgType         uint32
	MsgLength       uint32
	RequestID       uint32
	MajorVersion    uint32
	MinorVersion    uint32
	MaxTransferSize uint32
}

type rndisSetMsg struct {
	MsgType   uint32
	MsgLength uint32
	RequestID uint32
	OID       uint32
	Length    uint32
	Offset    uint32
	Reserved  uint32
	OIDp      uint32
}

func makeRndis(length uint32) rndisMessage {
	var rndis rndisMessage

	rndis.MsgType = 1
	rndis.MsgLength = length + 44
	rndis.DataOffset = 0x24
	rndis.DataLength = length

	return rndis
}

func (msg rndisMessage) bytes() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}
