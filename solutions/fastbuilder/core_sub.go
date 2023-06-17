//go:build !is_tweak

package fastbuilder

import (
	"phoenixbuilder/io/commands"
	"phoenixbuilder/minecraft/protocol/packet"
)

func SubFunc(sender *commands.CommandSender, resp *packet.CommandOutput) {}
