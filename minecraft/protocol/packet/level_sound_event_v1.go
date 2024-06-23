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
	Unknown1  uint8
	Posistion mgl32.Vec3
	Unknown2  int32
	Unknown3  int32
	Unknown4  bool
	Unknown5  bool
}

// ID ...
func (*LevelSoundEventV1) ID() uint32 {
	return IDLevelSoundEventV1
}

func (pk *LevelSoundEventV1) Marshal(io protocol.IO) {
	io.Uint8(&pk.Unknown1)
	io.Vec3(&pk.Posistion)
	io.Varint32(&pk.Unknown2)
	io.Varint32(&pk.Unknown3)
	io.Bool(&pk.Unknown4)
	io.Bool(&pk.Unknown5)
}
