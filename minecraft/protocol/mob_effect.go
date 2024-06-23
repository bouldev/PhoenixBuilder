/*
PhoenixBuilder specific.
Author: Liliya233
*/
package protocol

import "github.com/go-gl/mathgl/mgl32"

// Netease
type MobEffectV2UnknownEntry struct {
	Unknown1  int32
	Posistion mgl32.Vec3
	Unknown2  string
	Unknown3  string
}

// Netease
func (m *MobEffectV2UnknownEntry) Marshal(io IO) {
	io.Varint32(&m.Unknown1)
	io.Vec3(&m.Posistion)
	io.String(&m.Unknown2)
	io.String(&m.Unknown3)
}
