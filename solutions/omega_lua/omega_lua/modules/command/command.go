package command

import (
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/omega_lua/omega_lua/pollers"

	lua "github.com/yuin/gopher-lua"
)

func init() {
	if false {
		func(pollers.EventDataProvider) {}(&CmdRespHandler{})
	}
}

type CmdSender interface {
	SendWOCmd(cmd string)
	SendWSCmd(cmd string)
	SendPlayerCmd(cmd string)
	SendWSCmdAndInvokeOnResponse(string, func(output *packet.CommandOutput))
	SendPlayerCmdAndInvokeOnResponseWithFeedback(string, func(output *packet.CommandOutput))
}

type CmdModule struct {
	sender        CmdSender
	eventChan     pollers.EventChan
	packetWrapper func(goPacket packet.Packet, L *lua.LState) lua.LValue
}

func NewCmdModule(sender CmdSender,
	packetWrapper func(goPacket packet.Packet, L *lua.LState) lua.LValue) *CmdModule {
	eventChan := make(pollers.EventChan)
	return &CmdModule{sender: sender, eventChan: eventChan, packetWrapper: packetWrapper}
}

func (m *CmdModule) MakeLValue(L *lua.LState, luaAsyncInvoker pollers.LuaAsyncInvoker) lua.LValue {
	luaModule := L.NewTable()
	luaModule = L.SetFuncs(luaModule, map[string]lua.LGFunction{
		"send_wo_cmd":               m.sendWoCmd,
		"send_ws_cmd":               m.sendWsCmd,
		"send_player_cmd":           m.sendPlayerCmd,
		"send_ws_cmd_with_resp":     m.sendWsCmdAndInvokeOnResponse,
		"send_player_cmd_with_resp": m.sendPlayerCmdAndInvokeOnResponseWithFeedback,
	})
	registerCmdRespHandler(L)
	respHandler := NewCmdRespHandler(m.eventChan, luaAsyncInvoker)
	L.SetField(luaModule, "resp", respHandler.MakeLValue(L))
	return luaModule
}

func (m *CmdModule) sendWsCmd(L *lua.LState) int {
	cmd := L.CheckString(1)
	m.sender.SendWSCmd(cmd)
	return 0
}

func (m *CmdModule) sendWoCmd(L *lua.LState) int {
	cmd := L.CheckString(1)
	m.sender.SendWOCmd(cmd)
	return 0
}

func (m *CmdModule) sendPlayerCmd(L *lua.LState) int {
	cmd := L.CheckString(1)
	m.sender.SendPlayerCmd(cmd)
	return 0
}

func (m *CmdModule) sendWsCmdAndInvokeOnResponse(L *lua.LState) int {
	cmd := L.CheckString(1)
	attachment := L.Get(2)
	event := L.NewTable()
	m.sender.SendWSCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		go func() {
			L.SetTable(event, lua.LString("output"), m.packetWrapper(output, L))
			L.SetTable(event, lua.LString("cb"), attachment)
			m.eventChan <- event
		}()
	})
	return 0
}

func (m *CmdModule) sendPlayerCmdAndInvokeOnResponseWithFeedback(L *lua.LState) int {
	cmd := L.CheckString(1)
	attachment := L.Get(2)
	event := L.NewTable()
	m.sender.SendPlayerCmdAndInvokeOnResponseWithFeedback(cmd, func(output *packet.CommandOutput) {
		go func() {
			L.SetTable(event, lua.LString("output"), m.packetWrapper(output, L))
			L.SetTable(event, lua.LString("cb"), attachment)
			m.eventChan <- event
		}()
	})
	return 0
}
