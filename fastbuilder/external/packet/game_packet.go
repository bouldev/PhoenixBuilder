package packet

type GamePacket struct {
	Content []byte
}

func (pkt GamePacket) Marshal() []byte {
	return pkt.Content
}

func (pkt GamePacket) Parse(cont []byte) bool {
	pkt.Content=cont
	return true
}

func (_ GamePacket) ID() uint8 {
	return 7
}

func (_ GamePacket) Name() string {
	return "GamePacket"
}
