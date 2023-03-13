package embed

import (
	"crypto/md5"
	"fmt"
	blockNBT_API "phoenixbuilder/fastbuilder/bdump/blockNBT/API"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/mainframe"
	"strings"
	"time"
)

type EmbeddedAdaptor struct {
	env              *environment.PBEnvironment
	BackendCmdFeeder chan string
	PacketFeeder     chan *defines.CombinedPacket
	ChunkDataFeeder  chan *mirror.ChunkData
}

func (ea *EmbeddedAdaptor) FeedBackendCommand(cmd string) {
	ea.BackendCmdFeeder <- cmd
}

func (ea *EmbeddedAdaptor) GetBackendCommandFeeder() chan string {
	return ea.BackendCmdFeeder
}

func (ea *EmbeddedAdaptor) FeedPacketAndByte(pkt mc_packet.Packet, data []byte) {
	ea.PacketFeeder <- &defines.CombinedPacket{P: pkt, D: data}
}

func (ea *EmbeddedAdaptor) GetPacketFeeder() chan *defines.CombinedPacket {
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

func (rc *EmbeddedAdaptor) Write(pkt mc_packet.Packet) (err error) {
	return rc.env.Connection.(*minecraft.Conn).WritePacket(pkt)
}

func (rc *EmbeddedAdaptor) WriteBytes(data []byte) (err error) {
	_, err = rc.env.Connection.(*minecraft.Conn).Write(data)
	return err
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

func (ea *EmbeddedAdaptor) QuerySensitiveInfo(key defines.SensitiveInfoType) (result string, err error) {
	rawVal := ""
	switch key {
	case defines.SENSITIVE_INFO_SERVER_CODE_HASH:
		rawVal = ea.env.ServerCode
	case defines.SENSITIVE_INFO_USERNAME_HASH:
		_frags := strings.Split(ea.env.FBUCUsername, "|")
		if len(_frags) > 0 {
			rawVal = _frags[0]
		}
	}
	if rawVal == "" {
		return "", fmt.Errorf("no result")
	} else {
		cvt := func(in [16]byte) []byte {
			return in[:16]
		}
		hashedBytes := cvt(md5.Sum([]byte(rawVal)))
		return fmt.Sprintf("%x", hashedBytes), nil
	}
}

func EnableOmegaSystem(env *environment.PBEnvironment) (*EmbeddedAdaptor, func()) {
	ea := &EmbeddedAdaptor{
		env:              env,
		BackendCmdFeeder: make(chan string, 1024),
		PacketFeeder:     make(chan *defines.CombinedPacket, 1024),
		ChunkDataFeeder:  make(chan *mirror.ChunkData, 1024),
	}

	omega := mainframe.NewOmega()
	omega.NewUQHolder = env.NewUQHolder.(*blockNBT_API.PacketHandleResult)
	omega.Bootstrap(ea)
	env.OmegaHolder = omega
	env.OmegaAdaptorHolder = ea
	env.Destructors = append(env.Destructors, func() {
		omega.Stop()
	})
	return ea, func() {
		fmt.Println("Starting Omega in a second")
		time.Sleep(time.Second)
		omega.Activate()
	}
}
