package main

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/omega_lua/omega_lua"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"phoenixbuilder/solutions/omega_lua/omega_lua/lua_utils"
	"phoenixbuilder/solutions/omega_lua/omega_lua/mux_pumper"
	"phoenixbuilder/solutions/remote_omega/transfer"

	lua "github.com/yuin/gopher-lua"
)

func CreateLuaEnv(ctx context.Context, omegaCore omega.MicroOmega, config *lua_utils.LuaConfigRaw) (ac concurrent.AsyncCtrl, L *lua.LState) {
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
	}, config, "./storage")
}

//go:embed test.lua
var luaCode []byte

func main() {
	omegaCore, err := transfer.ConnectToAccessPoint(nil, nil, nil)
	if err != nil {
		panic(err)
	}
	config := map[string]interface{}{
		"Version":   "0.0.1",
		"SomeEntry": "SomeData",
		"Users": map[string]interface{}{
			"2401PT": "architecture",
			"343GS":  "somebody",
		},
	}
	ac, L := CreateLuaEnv(context.Background(), omegaCore, &lua_utils.LuaConfigRaw{
		Config: config,
		OnConfigUpdate: func(newConfig interface{}) {
			fmt.Printf("config upgrade to %v\n", newConfig)
		},
	})
	errChan := concurrent.FireLuaCodeInGoRoutine(ac, L, string(luaCode))
	// wait for lua code to finish
	err = <-errChan
	if err != nil {
		panic(err)
	}
}
