package packet

type EvalPBCommandPacket struct {
	Command string
}

func (pkt EvalPBCommandPacket) Marshal() {
	return []byte(pkt.Command)
}

func (pkt EvalPBCommandPacket) Parse(cont []byte) bool {
	pkt.Command=string(cont)
	return true
}

func (_ EvalPBCommandPacket) ID() uint8 {
	return 5
}

func (_ EvalPBCommandPacket) Name() string {
	return "EvalPBCommandPacket"
}

type GameCommandPacket struct {
	Command string
}

func (pkt GameCommandPacket) Marshal() {
	return []byte(pkt.Command)
}

func (pkt GameCommandPacket) Parse(cont []byte) bool {
	pkt.Command=string(cont)
	return true
}

func (_ GameCommandPacket) ID() uint8 {
	return 6
}

func (_ GameCommandPacket) Name() string {
	return "GameCommandPacket"
}