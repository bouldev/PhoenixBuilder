package packet

var PacketPool map[uint8]func()Packet=map[uint8]func()Packet {
	1: func() Packet { return PingPacket{} },
	2: func() Packet { return PongPacket{} },
	3: func() Packet { return ByePacket{} },
	4: func() Packet { return PacketViolationWarningPacket{} },
	5: func() Packet { return EvalPBCommandPacket{} },
	6: func() Packet { return GameCommandPacket{} },
	7: func() Packet { return GamePacket{} },
}