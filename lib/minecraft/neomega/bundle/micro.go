package bundle

import (
	"phoenixbuilder/lib/minecraft/neomega/decouple/block/placer"
	"phoenixbuilder/lib/minecraft/neomega/decouple/cmdsender"
	"phoenixbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/lib/minecraft/neomega/decouple/infosender"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/lib/minecraft/neomega/uqholder"
	"phoenixbuilder/minecraft"
)

func init() {
	if false {
		func(omega omega.MicroOmega) {}(&MicroOmega{})
	}
}

type MicroOmega struct {
	omega.ReactCore
	omega.InteractCore
	omega.InfoSender
	omega.CmdSender
	omega.BotBasicInfoHolder
	omega.BlockPlacer
}

func (o *MicroOmega) GetBotInfo() omega.BotBasicInfoHolder {
	return o.BotBasicInfoHolder
}

type MicroOmegaOption struct {
	CmdSenderOptions       cmdsender.Options
	PrintUQHolderDebugInfo bool
}

func NewMicroOmega(conn *minecraft.Conn, options MicroOmegaOption) omega.MicroOmega {
	reactable := core.NewReactCore()
	interactCore := core.NewInteractCore(conn)
	cmdSender := cmdsender.NewCmdSender(reactable, interactCore, options.CmdSenderOptions)
	botBasicInfoHolder := uqholder.NewBotInfoHolder(conn, options.PrintUQHolderDebugInfo)
	infoSender := infosender.NewInfoSender(interactCore, cmdSender, botBasicInfoHolder)
	blockPlacer := placer.NewBlockPlacer(reactable, cmdSender, interactCore)
	return &MicroOmega{
		reactable,
		interactCore,
		infoSender,
		cmdSender,
		botBasicInfoHolder,
		blockPlacer,
	}
}

func (o *MicroOmega) GetGameControl() interface {
	omega.InteractCore
	omega.CmdSender
	omega.InfoSender
	omega.BlockPlacer
} {
	return o
}

func (o *MicroOmega) GetGameListener() omega.PacketDispatcher {
	return o
}
