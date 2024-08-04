/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease Packet
type AnimateEmoteEntity struct {
	Animation             string
	NextState             string
	StopExpression        string
	StopExpressionVersion int32
	Controller            string
	BlendOutTime          float32
	RuntimeEntityIds      []uint64
}

// ID ...
func (*AnimateEmoteEntity) ID() uint32 {
	return IDAnimateEmoteEntity
}

func (pk *AnimateEmoteEntity) Marshal(io protocol.IO) {
	io.String(&pk.Animation)
	io.String(&pk.NextState)
	io.String(&pk.StopExpression)
	io.Int32(&pk.StopExpressionVersion)
	io.String(&pk.Controller)
	io.Float32(&pk.BlendOutTime)
	protocol.FuncSliceVarint32Length(io, &pk.RuntimeEntityIds, io.Varuint64)
}
