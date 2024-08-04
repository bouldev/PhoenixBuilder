/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import (
	"phoenixbuilder/minecraft/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// Netease packet
type LevelSoundEventV1 struct {
	SoundType uint8
	Posistion mgl32.Vec3
	ExtraData int32
	Pitch     int32
	IsBabyMob bool
	IsGlobal  bool
}

// ID ...
func (*LevelSoundEventV1) ID() uint32 {
	return IDLevelSoundEventV1
}

func (pk *LevelSoundEventV1) Marshal(io protocol.IO) {
	io.Uint8(&pk.SoundType)
	io.Vec3(&pk.Posistion)
	io.Varint32(&pk.ExtraData)
	io.Varint32(&pk.Pitch)
	io.Bool(&pk.IsBabyMob)
	io.Bool(&pk.IsGlobal)
}
