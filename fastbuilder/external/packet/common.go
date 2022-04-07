package packet

// Packets with a small size won't be used as pointer

type PingPacket struct {
}

func (_ *PingPacket) Marshal() []byte {
	return []byte{}
}

func (_ *PingPacket) Parse(_ []byte) bool {
	return true
}

func (_ *PingPacket) ID() uint8 {
	return IDPingPacket
}

func (_ *PingPacket) Name() string {
	return "PingPacket"
}

type PongPacket struct {
}

func (_ *PongPacket) Marshal() []byte {
	return []byte{}
}

func (_ *PongPacket) Parse(_ []byte) bool {
	return true
}

func (_ *PongPacket) ID() uint8 {
	return IDPongPacket
}

func (_ *PongPacket) Name() string {
	return "PongPacket"
}

type ByePacket struct {
}

func (_ *ByePacket) Marshal() []byte {
	return []byte{}
}

func (_ *ByePacket) Parse(_ []byte) bool {
	return true
}

func (_ *ByePacket) ID() uint8 {
	return IDByePacket
}

func (_ *ByePacket) Name() string {
	return "ByePacket"
}

type PacketViolationWarningPacket struct {
	Text string
}

func (pkt *PacketViolationWarningPacket) Marshal() []byte {
	return []byte(pkt.Text)
}

func (pkt *PacketViolationWarningPacket) Parse(cont []byte) bool {
	pkt.Text = string(cont)
	return true
}

func (_ *PacketViolationWarningPacket) ID() uint8 {
	return IDPacketViolationWarningPacket
}

func (_ *PacketViolationWarningPacket) Name() string {
	return "PacketViolationWarningPacket"
}

type UQHolderRequestPacket struct {
	QueryString []byte
}

func (pkt *UQHolderRequestPacket) Marshal() []byte {
	return pkt.QueryString
}

func (pkt *UQHolderRequestPacket) Parse(cont []byte) bool {
	pkt.QueryString = cont
	return true
}

func (_ *UQHolderRequestPacket) ID() uint8 {
	return IDUQHolderRequestPacket
}

func (_ *UQHolderRequestPacket) Name() string {
	return "UQHolderRequestPacket"
}

type UQHolderResponsePacket struct {
	Content []byte
}

func (pkt *UQHolderResponsePacket) Marshal() []byte {
	return pkt.Content
}

func (pkt *UQHolderResponsePacket) Parse(cont []byte) bool {
	pkt.Content = cont
	return true
}

func (_ *UQHolderResponsePacket) ID() uint8 {
	return IDUQHolderResponsePacket
}

func (_ *UQHolderResponsePacket) Name() string {
	return "UQHolderResponsePacket"
}
