/*
PhoenixBuilder specific NEMC packet.
Author: LNSSPsd, Liliya233, CMA2401PT, Happy2018new
*/
package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

type PyRpc struct {
	Value any
}

// ID ...
func (*PyRpc) ID() uint32 {
	return IDPyRpc
}

// Marshal ...
func (pk *PyRpc) Marshal(io protocol.IO) {
	io.MsgPack(&pk.Value)
}
