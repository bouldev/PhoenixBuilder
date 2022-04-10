package packet

const (
	IDPingPacket = iota + 1
	IDPongPacket
	IDByePacket
	IDPacketViolationWarningPacket
	IDEvalPBCommandPacket
	IDGameCommandPacket
	IDGamePacket
	IDUQHolderRequestPacket
	IDUQHolderResponsePacket
	IDGamePacketReducePacket
)

var PacketPool map[uint8]func() Packet = map[uint8]func() Packet{
	IDPingPacket:                   func() Packet { return &PingPacket{} },
	IDPongPacket:                   func() Packet { return &PongPacket{} },
	IDByePacket:                    func() Packet { return &ByePacket{} },
	IDPacketViolationWarningPacket: func() Packet { return &PacketViolationWarningPacket{} },
	IDEvalPBCommandPacket:          func() Packet { return &EvalPBCommandPacket{} },
	IDGameCommandPacket:            func() Packet { return &GameCommandPacket{} },
	IDGamePacket:                   func() Packet { return &GamePacket{} },
	IDUQHolderRequestPacket:        func() Packet { return &UQHolderRequestPacket{} },
	IDUQHolderResponsePacket:       func() Packet { return &UQHolderResponsePacket{} },
	IDGamePacketReducePacket:       func() Packet { return &GamePacketReducePacket{} },
}
