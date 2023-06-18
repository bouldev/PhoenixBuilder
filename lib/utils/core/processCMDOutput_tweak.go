//go:build is_tweak

package utils_core

import (
	"phoenixbuilder/io/commands"
	"phoenixbuilder/minecraft/protocol/packet"
)

func ProcessCommandOutput(sender *commands.CommandSender, resp *packet.CommandOutput) {
	pr, ok := sender.UUIDMap.LoadAndDelete(resp.CommandOrigin.UUID.String())
	if ok {
		pu := pr.(chan *packet.CommandOutput)
		pu <- resp
	}
}
