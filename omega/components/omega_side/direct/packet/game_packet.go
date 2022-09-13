package packet

type GamePacket struct {
	Content []byte
}

func (pkt *GamePacket) Marshal() []byte {
	return pkt.Content
}

func (pkt *GamePacket) Parse(cont []byte) bool {
	pkt.Content = cont
	return true
}

func (_ *GamePacket) ID() uint8 {
	return IDGamePacket
}

func (_ *GamePacket) Name() string {
	return "GamePacket"
}

type GamePacketReducePacket struct {
	PacketID uint8
	DropBy   uint8
}

func (pkt *GamePacketReducePacket) Marshal() []byte {
	return []byte{pkt.PacketID, pkt.DropBy}
}

func (pkt *GamePacketReducePacket) Parse(cont []byte) bool {
	if len(cont) != 2 {
		return false
	}
	pkt.PacketID = cont[0]
	pkt.DropBy = cont[1]
	return true
}

func (_ *GamePacketReducePacket) ID() uint8 {
	return IDGamePacketReducePacket
}

func (_ *GamePacketReducePacket) Name() string {
	return "GamePacketReducePacket"
}
