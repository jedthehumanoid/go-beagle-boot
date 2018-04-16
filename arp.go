package main

type arpMessage struct {
	HardwareType       uint16
	ProtocolType       uint16
	HardwareAddrLength uint8
	ProtocolAddrLength uint8
	Opcode             uint16
	SenderHardwareAddr [6]byte
	SenderProtocolAddr [4]byte
	TargetHardwareAddr [6]byte
	TargetProtocolAddr [4]byte
}

func makeARPMessage(opcode uint16, senderHwAddr [6]byte, senderIp [4]byte,
	targetHwAddr [6]byte, targetIp [4]byte) arpMessage {
	var m arpMessage

	m.HardwareType = 1
	m.ProtocolType = 0x800
	m.HardwareAddrLength = 6
	m.ProtocolAddrLength = 4
	m.Opcode = opcode
	m.SenderHardwareAddr = senderHwAddr
	m.SenderProtocolAddr = senderIp
	m.TargetHardwareAddr = targetHwAddr
	m.TargetProtocolAddr = targetIp

	return m
}
