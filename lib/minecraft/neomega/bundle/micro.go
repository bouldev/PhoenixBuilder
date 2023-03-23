package bundle

import (
	minecraft "fastbuilder-core/lib/minecraft/gophertunnel"
	"fastbuilder-core/lib/minecraft/neomega/decouple/cmdsender"
	"fastbuilder-core/lib/minecraft/neomega/decouple/core"
	"fastbuilder-core/lib/minecraft/neomega/decouple/infosender"
	"fastbuilder-core/lib/minecraft/neomega/omega"
	"fastbuilder-core/lib/minecraft/neomega/uqholder"
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
	return &MicroOmega{
		reactable,
		interactCore,
		infoSender,
		cmdSender,
		botBasicInfoHolder,
	}
}

func (o *MicroOmega) GetGameControl() interface {
	omega.InteractCore
	omega.CmdSender
	omega.InfoSender
} {
	return o
}

func (o *MicroOmega) GetGameListener() omega.PacketDispatcher {
	return o
}
