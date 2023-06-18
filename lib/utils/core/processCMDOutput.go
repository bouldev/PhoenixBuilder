//go:build !is_tweak

package utils_core

import (
	"phoenixbuilder/io/commands"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 此实现在非 is_tweak 构建下已被迁移至
// package GlobalAPI(pheonixbuilder/GameControl/GlobalAPI)
func ProcessCommandOutput(sender *commands.CommandSender, resp *packet.CommandOutput) {}
