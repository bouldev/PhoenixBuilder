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
type LevelSoundEventV2 struct {
	SoundType        uint8
	Posistion        mgl32.Vec3
	ExtraData        int32
	EntityIdentifier string
	IsBabyMob        bool
	IsGlobal         bool
}

// ID ...
func (*LevelSoundEventV2) ID() uint32 {
	return IDLevelSoundEventV2
}

func (pk *LevelSoundEventV2) Marshal(io protocol.IO) {
	io.Uint8(&pk.SoundType)
	io.Vec3(&pk.Posistion)
	io.Varint32(&pk.ExtraData)
	io.String(&pk.EntityIdentifier)
	io.Bool(&pk.IsBabyMob)
	io.Bool(&pk.IsGlobal)
}
