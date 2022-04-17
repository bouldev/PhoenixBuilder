package embed

import (
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/mainframe"
)

type EmbeddedAdaptor struct {
	env              *environment.PBEnvironment
	BackendCmdFeeder chan string
	PacketFeeder     chan mc_packet.Packet
}

func (ea *EmbeddedAdaptor) FeedBackendCommand(cmd string) {
	ea.BackendCmdFeeder <- cmd
}

func (ea *EmbeddedAdaptor) GetBackendCommandFeeder() chan string {
	return ea.BackendCmdFeeder
}

func (ea *EmbeddedAdaptor) FeedPacket(pkt mc_packet.Packet) {
	ea.PacketFeeder <- pkt
}

func (ea *EmbeddedAdaptor) Read() mc_packet.Packet {
	return <-ea.PacketFeeder
}

func (rc *EmbeddedAdaptor) GetInitUQHolderCopy() *uqHolder.UQHolder {
	origHolder := rc.env.UQHolder.(*uqHolder.UQHolder)
	holderBytes := origHolder.Marshal()
	newHolder := uqHolder.NewUQHolder(origHolder.BotRuntimeID)
	err := newHolder.UnMarshal(holderBytes)
	if err != nil {
		return nil
	}
	return newHolder
}

func (rc *EmbeddedAdaptor) Write(pkt mc_packet.Packet) {
	rc.env.Connection.(*minecraft.Conn).WritePacket(pkt)
}

func (rc *EmbeddedAdaptor) FBEval(cmd string) {
	rc.env.FunctionHolder.(*function.FunctionHolder).Process(cmd)
}

func EnableOmegaSystem(env *environment.PBEnvironment) *EmbeddedAdaptor {
	ea := &EmbeddedAdaptor{
		env:              env,
		BackendCmdFeeder: make(chan string, 1024),
		PacketFeeder:     make(chan mc_packet.Packet, 1024),
	}
	omega := mainframe.NewOmega()
	omega.SetRoot("omega_storage")
	omega.Bootstrap(ea)
	env.OmegaHolder = omega
	env.OmegaAdaptorHolder = ea
	env.Destructors = append(env.Destructors, func() {
		omega.Stop()
	})
	go omega.Activate()
	return ea
}
