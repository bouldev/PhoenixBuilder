package packet

type GamePacket struct {
	Content []byte
}

func (pkt GamePacket) Marshal() {
	return pkt.Content
}

func (pkt GamePacket) Parse(cont []byte) bool {
	pkt.Command=cont
	return true
}

func (_ GamePacket) ID() uint8 {
	return 7
}

func (_ GamePacket) Name() string {
	return "GamePacket"
}
