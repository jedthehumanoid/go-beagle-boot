package main

type bootpMessage struct {
	Opcode             uint8
	HardwareType       uint8
	HardwareAddrLength uint8
	Hops               uint8
	Xid                uint32
	Seconds            uint16
	Flags              uint16
	ClientIPAddr       [4]byte
	YourIPAddr         [4]byte
	ServerIPAddr       [4]byte
	GatewayIPAddr      [4]byte
	ClientHardwareAddr [16]byte
	ServerName         [64]byte
	BootFilename       [128]byte
	VendorSpecific     [64]byte
}

type specialbootpMessage struct {
	Opcode             uint8
	HardwareType       uint8
	HardwareAddrLength uint8
	Hops               uint8
	Xid                uint32
	Seconds            uint16
	Flags              uint16
	ClientIPAddr       [4]byte
	YourIPAddr         [4]byte
	ServerIPAddr       [4]byte
	GatewayIPAddr      [4]byte
	ClientHardwareAddr [16]byte
	ServerName         [64]byte
	BootFilename       [128]byte
}

func makeBootpPacket(servername string, id uint32, clientHwAddr [6]byte, yourIp [4]byte, serverIp [4]byte, bootfile string) bootpMessage {
	var ret bootpMessage
	ret.Opcode = 2
	ret.HardwareType = 1
	ret.HardwareAddrLength = 6

	copy(ret.ClientHardwareAddr[:], clientHwAddr[:])

	ret.Xid = id
	ret.YourIPAddr = yourIp
	ret.ServerIPAddr = serverIp
	ret.GatewayIPAddr = serverIp

	ret.VendorSpecific = [64]byte{99, 130, 83, 99, 1, 4, 255, 255, 255, 0, 3, 4, 192, 168, 1, 9, 0xFF}

	copy(ret.ServerName[:], servername)
	copy(ret.BootFilename[:], bootfile)

	return ret
}
