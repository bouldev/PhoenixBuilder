/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// Netease packet
type WithdrawFurnaceXp struct {
	Position protocol.BlockPos
}

// ID ...
func (*WithdrawFurnaceXp) ID() uint32 {
	return IDWithdrawFurnaceXp
}

func (pk *WithdrawFurnaceXp) Marshal(io protocol.IO) {
	io.BlockPos(&pk.Position)
}
