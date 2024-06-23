/*
PhoenixBuilder specific NEMC packet.
Author: CMA2401PT, Liliya233
*/
package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

type NemcWhatever struct {
	Value any
}

// ID ...
func (*NemcWhatever) ID() uint32 {
	return IDPyRpc
}

// Marshal ...
func (pk *NemcWhatever) Marshal(io protocol.IO) {}
