package main

import (
	"bufio"
	"context"
	_ "embed"
	lua "github.com/yuin/gopher-lua"
	"os"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/omega_lua/omega_lua"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"phoenixbuilder/solutions/omega_lua/omega_lua/mux_pumper"
	"phoenixbuilder/solutions/remote_omega/transfer"
)

func CreateLuaEnv(ctx context.Context, omegaCore omega.MicroOmega) (ac concurrent.AsyncCtrl, L *lua.LState) {
	L = lua.NewState()
	ac = concurrent.NewAsyncCtrl(ctx)
	// go implements
	// 1. monk system
	goInputPumper := mux_pumper.NewInputPumperMux()
	goSystem := NewLuaGoSystem(goInputPumper)
	go func() {
		for {
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				panic(err)
			}
			goInputPumper.PumpInput(input)
		}
	}()
	goGamePacketPumper := mux_pumper.NewGamePacketPumperMux()
	omegaCore.GetGameListener().SetOnAnyPacketCallBack(func(pk packet.Packet) {
		goGamePacketPumper.PumpGamePacket(pk)
	})
	goPackets := mux_pumper.NewPacketDispatcher(128, goGamePacketPumper)
	goCmdSender := omegaCore.GetGameControl()
	return omega_lua.CreateOmegaLuaEnv(ctx, &omega_lua.GoImplements{
		GoSystem:         goSystem,
		GoPackets:        goPackets,
		GoPacketProvider: goPackets,
		GoCmdSender:      goCmdSender,
	})
}

//go:embed test.lua
var luaCode []byte

func main() {
	omegaCore, err := transfer.ConnectToAccessPoint(nil, nil, nil)
	if err != nil {
		panic(err)
	}
	ac, L := CreateLuaEnv(context.Background(), omegaCore)
	errChan := concurrent.FireLuaCodeInGoRoutine(ac, L, string(luaCode))
	// wait for lua code to finish
	err = <-errChan
	if err != nil {
		panic(err)
	}
}
