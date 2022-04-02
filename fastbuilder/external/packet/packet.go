package packet

type Packet interface {
	Marshal() []byte
	Parse([]byte) bool
	ID() uint8
	Name() string
}

func (pkt Packet) Serialize() []byte {
	return append([]byte{pkt.ID()}, pkt.Serialize())
}

func Deserialize(content []byte) (Packet, bool) {
	packetGenerator, found:=PacketPool[content[0]]
	if !found {
		return nil, found
	}
	packet:=packetGenerator()
	parsed:=packet.Parse(content[1:])
	if !parsed {
		return nil, parsed
	}
	return packet, true
}