// +build !is_tweak

package commands

import (
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

func (sender *CommandSender) UpdateCommandBlock(x int32,y int32,z int32,d *types.CommandBlockData) {
	sender.getConn().WritePacket(&packet.CommandBlockUpdate{
		Block: true,
		Position: protocol.BlockPos {x,y,z},
		Mode: d.Mode,
		NeedsRedstone: d.NeedRedstone,
		Conditional: d.Conditional,
		Command: d.Command,
		LastOutput: d.LastOutput,
		Name: d.CustomName,
		TickDelay: d.TickDelay,
		ExecuteOnFirstTick: d.ExecuteOnFirstTick,
	})
}