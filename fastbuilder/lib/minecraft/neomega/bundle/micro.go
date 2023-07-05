package bundle

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/decouple/block/placer"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/decouple/cmdsender"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/decouple/infosender"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
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
	omega.MicroUQHolder
	omega.BlockPlacer
}

func (o *MicroOmega) GetMicroUQHolder() omega.MicroUQHolder {
	return o.MicroUQHolder
}

type MicroOmegaOption struct {
	CmdSenderOptions *cmdsender.Options
}

func MakeDefaultMicroOmegaOption() *MicroOmegaOption {
	return &MicroOmegaOption{CmdSenderOptions: cmdsender.MakeDefaultCmdSenderOption()}
}

func NewMicroOmega(interactCore omega.InteractCore, getMicroUQHolder func() omega.MicroUQHolder, options *MicroOmegaOption) *MicroOmega {
	if options == nil {
		options = MakeDefaultMicroOmegaOption()
	}
	reactable := core.NewReactCore()
	//interactCore := core.NewInteractCore(conn)
	//conn.ReadPacketAndBytes()
	cmdSender := cmdsender.NewCmdSender(reactable, interactCore, options.CmdSenderOptions)
	microUQHolder := getMicroUQHolder()
	infoSender := infosender.NewInfoSender(interactCore, cmdSender, microUQHolder.GetBotBasicInfo())
	blockPlacer := placer.NewBlockPlacer(reactable, cmdSender, interactCore)
	return &MicroOmega{
		reactable,
		interactCore,
		infoSender,
		cmdSender,
		microUQHolder,
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
