package game_packet

import (
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type OmegaPacketsModule struct {
	goImplements          LuaGoPackets
	MCPacketNameIDMapping map[string]uint32
	MCPacketIDToLuaName   map[uint32]lua.LString
	MCPacketIDToLuaInt    map[uint32]lua.LNumber
}

func NewOmegaPacketsModule(goImplements LuaGoPackets) *OmegaPacketsModule {
	mapping := goImplements.GetMCPacketNameIDMapping()
	m := &OmegaPacketsModule{
		goImplements:          goImplements,
		MCPacketNameIDMapping: mapping,
	}
	m.MCPacketIDToLuaName = make(map[uint32]lua.LString)
	m.MCPacketIDToLuaInt = make(map[uint32]lua.LNumber)
	for name, id := range m.MCPacketNameIDMapping {
		m.MCPacketIDToLuaName[id] = lua.LString(name)
		m.MCPacketIDToLuaInt[id] = lua.LNumber(id)
	}
	return m
}

func (m *OmegaPacketsModule) MakeLValue(L *lua.LState) lua.LValue {
	packetModule := L.NewTable()
	for goId, goName := range m.MCPacketIDToLuaName {
		luaID := m.MCPacketIDToLuaInt[goId]
		luaName := m.MCPacketIDToLuaName[goId]
		L.SetTable(packetModule, luaName, luaName)
		L.SetTable(packetModule, lua.LString("no"+goName), lua.LString("!"+goName))
		L.SetTable(packetModule, lua.LString("ID"+goName), luaID)
	}
	L.SetTable(packetModule, lua.LString("all"), lua.LString("all"))
	allNames := make([]string, 0, len(m.MCPacketNameIDMapping))
	for name := range m.MCPacketNameIDMapping {
		allNames = append(allNames, name)
	}
	L.SetTable(packetModule, lua.LString("all_names"), lua.LString(
		"["+strings.Join(allNames, ",")+"]",
	))
	L.SetTable(packetModule, lua.LString("to_json_string_slow"), L.NewFunction(m.luaGoPacketsToJSONString))
	L.SetTable(packetModule, lua.LString("to_lua_table"), L.NewFunction(m.luaGoPacketsToLuaTable))
	registerGamePacket(L)
	return packetModule
}

func (m *OmegaPacketsModule) WrapPacketToLuaPacket(pk packet.Packet) *GamePacket {
	pkID := pk.ID()
	luaID := m.MCPacketIDToLuaInt[pkID]
	luaName := m.MCPacketIDToLuaName[pkID]
	return NewGamePacket(
		pk,
		luaName,
		luaID,
	)
}

// packet.to_json_string(pk:packet)
func (m *OmegaPacketsModule) luaGoPacketsToJSONString(L *lua.LState) int {
	pk := checkGamePacket(L)
	jsonBytes, _ := json.Marshal(pk.goPacket)
	L.Push(lua.LString(jsonBytes))
	return 1
}

// packet.to_lua_table(pk:packet)
func (m *OmegaPacketsModule) luaGoPacketsToLuaTable(L *lua.LState) int {
	pk := checkGamePacket(L)
	luaTable := luar.New(L, pk.goPacket)
	L.Push(luaTable)
	return 1
}
