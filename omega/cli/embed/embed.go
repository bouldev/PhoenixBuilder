package embed

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/omega/mainframe"
	"time"
)

type EmbeddedAdaptor struct {
	env              *environment.PBEnvironment
	BackendCmdFeeder chan string
	PacketFeeder     chan mc_packet.Packet
	ChunkDataFeeder  chan *mirror.ChunkData
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

func (ea *EmbeddedAdaptor) GetPacketFeeder() chan mc_packet.Packet {
	return ea.PacketFeeder
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

func (ea *EmbeddedAdaptor) FeedChunkData(cd *mirror.ChunkData) {
	ea.ChunkDataFeeder <- cd
}

func (ea *EmbeddedAdaptor) GetChunkFeeder() chan *mirror.ChunkData {
	return ea.ChunkDataFeeder
}

func EnableOmegaSystem(env *environment.PBEnvironment) *EmbeddedAdaptor {
	ea := &EmbeddedAdaptor{
		env:              env,
		BackendCmdFeeder: make(chan string, 1024),
		PacketFeeder:     make(chan mc_packet.Packet, 1024),
		ChunkDataFeeder:  make(chan *mirror.ChunkData, 1024),
	}
	fmt.Println("Starting Omega in 1 Seconds")
	time.Sleep(time.Millisecond * 10)
	omega := mainframe.NewOmega()
	omega.Bootstrap(ea)
	env.OmegaHolder = omega
	env.OmegaAdaptorHolder = ea
	env.Destructors = append(env.Destructors, func() {
		omega.Stop()
	})
	go omega.Activate()
	return ea
}
