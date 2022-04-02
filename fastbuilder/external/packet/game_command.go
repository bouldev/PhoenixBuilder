package packet

import "github.com/google/uuid"
import "bytes"

const (
	CommandTypeNormal = 0
	CommandTypeWebsocket = 1
	CommandTypeSettings = 2 // w/ SuppressOutput=true
)

type GameCommandPacket struct {
	UUID uuid.UUID
	CommandType uint8
	Command string
}

func (pkt GameCommandPacket) Marshal() []byte {
	adb:=[]byte{}
	if(pkt.CommandType!=CommandTypeSettings) {
		bb, _:=pkt.UUID.MarshalBinary()
		adb=bb
	}
	return bytes.Join([][]byte{
		[]byte{pkt.CommandType},
		adb,
		[]byte(pkt.Command),
	},[]byte{})
}

func (pkt GameCommandPacket) Parse(cont []byte) bool {
	pkt.CommandType=cont[0]
	if(pkt.CommandType!=CommandTypeSettings) {
		err:=pkt.UUID.UnmarshalBinary(cont[1:17])
		if(err!=nil) {
			return false
		}
	}
	pkt.Command=string(cont[17:])
	return true
}

func (_ GameCommandPacket) ID() uint8 {
	return 6
}

func (_ GameCommandPacket) Name() string {
	return "GameCommandPacket"
}