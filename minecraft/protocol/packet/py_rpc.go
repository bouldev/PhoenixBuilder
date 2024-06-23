/*
PhoenixBuilder specific NEMC packet.
Author: LNSSPsd, Liliya233, CMA2401PT, Happy2018new
*/
package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

const (
	PyRpcOperationTypeSend = 0x05db23ae
	PyRpcOperationTypeRecv = 0x0094d408
)

type PyRpc struct {
	// Value from/to msgpack format
	Value any
	// OperationType is a fixed number
	OperationType uint32
}

// ID ...
func (*PyRpc) ID() uint32 {
	return IDPyRpc
}

// Marshal ...
func (pk *PyRpc) Marshal(io protocol.IO) {
	io.MsgPack(&pk.Value)
	io.Uint32(&pk.OperationType)
}
