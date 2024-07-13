package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 头颅
type Skull struct {
	general.BlockActor `mapstructure:",squash"`
	DoingAnimation     byte    `mapstructure:"DoingAnimation"` // * TAG_Byte(1) = 0
	MouthTickCount     int32   `mapstructure:"MouthTickCount"` // TAG_Int(4) = 0
	Rotation           float32 `mapstructure:"Rotation"`       // TAG_Float(6) = 0
	SkullType          byte    `mapstructure:"SkullType"`      // TAG_Byte(1) = 0
}

// ID ...
func (*Skull) ID() string {
	return IDSkull
}

func (s *Skull) Marshal(io protocol.IO) {
	protocol.Single(io, &s.BlockActor)
	protocol.NBTInt(&s.SkullType, io.Varuint16)
	io.Float32(&s.Rotation)
	io.Uint8(&s.DoingAnimation)
	protocol.NBTInt(&s.MouthTickCount, io.Varuint16)
}
