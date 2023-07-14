package access_helper

import (
	"fmt"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/bundle"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

type ReadLoopFunction func(conn *minecraft.Conn, deadReason chan<- error, omega omega.ReactCore)

var DefaultReadLoopFunction = func(conn *minecraft.Conn, deadReason chan<- error, omega omega.ReactCore) {
	var pkt packet.Packet
	var err error
	for {
		pkt, err = conn.ReadPacket()
		if err != nil {
			deadReason <- fmt.Errorf("%v: %v", ErrRentalServerDisconnected, err)
		}
		omega.HandlePacket(pkt)
	}
}

type Options struct {
	AuthServer                 string
	ServerConnectionTimeout    time.Duration
	ChallengeSolvingTimeout    time.Duration
	ServerConnectRetryTimes    int
	FBUsername                 string
	FBUserPassword             string
	FBUserToken                string
	ServerCode                 string
	ServerPassword             string
	WriteBackToken             bool
	MicroOmegaOption           *bundle.MicroOmegaOption
	PrintUQHolderDebugInfo     bool
	MakeBotCreative            bool
	DisableCommandBlock        bool
	MaximumWaitTime            time.Duration
	DieOnLosingOpPrivilege     bool
	OpPrivilegeRemovedCallBack func()
	ReadLoopFunction           ReadLoopFunction
	ReasonWithPrivilegeStuff   bool
}

func DefaultOptions() *Options {
	return &Options{
		AuthServer:                 "",
		ServerConnectionTimeout:    time.Minute,
		ChallengeSolvingTimeout:    time.Minute,
		ServerConnectRetryTimes:    0,
		FBUsername:                 "",
		FBUserPassword:             "",
		FBUserToken:                "",
		ServerCode:                 "",
		ServerPassword:             "",
		MicroOmegaOption:           bundle.MakeDefaultMicroOmegaOption(),
		PrintUQHolderDebugInfo:     false,
		MakeBotCreative:            true,
		DisableCommandBlock:        true,
		MaximumWaitTime:            time.Minute * 3,
		DieOnLosingOpPrivilege:     true,
		OpPrivilegeRemovedCallBack: nil,
		ReadLoopFunction:           DefaultReadLoopFunction,
		ReasonWithPrivilegeStuff:   true,
		WriteBackToken:             true,
	}
}
