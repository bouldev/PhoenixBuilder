package packet

type EvalPBCommandPacket struct {
	Command string
}

func (pkt EvalPBCommandPacket) Marshal() []byte {
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

