package game_packet

import (
	"phoenixbuilder/minecraft/protocol/packet"

	lua "github.com/yuin/gopher-lua"
)

type GamePacket struct {
	goPacket packet.Packet
	luaSelf  lua.LValue
	luaName  lua.LString
	luaID    lua.LNumber
}

func NewGamePacket(
	goPacket packet.Packet,
	luaName lua.LString,
	luaID lua.LNumber,
) *GamePacket {
	return &GamePacket{
		goPacket: goPacket,
		luaName:  luaName,
		luaID:    luaID,
	}
}

func (g *GamePacket) MakeLValue(L *lua.LState) lua.LValue {
	luaGamePacket := L.NewUserData()
	luaGamePacket.Value = g
	L.SetMetatable(luaGamePacket, L.GetTypeMetatable("game_packet"))
	g.luaSelf = luaGamePacket
	return luaGamePacket
}

func registerGamePacket(L *lua.LState) {
	mt := L.NewTypeMetatable("game_packet")
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"id":   gamePacketGetID,
		"name": gamePacketGetName,
	}))
}

func checkGamePacket(L *lua.LState) *GamePacket {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GamePacket); ok {
		return v
	}
	L.ArgError(1, "game packet expected")
	return nil
}

// game_packet:id()
func gamePacketGetID(L *lua.LState) int {
	g := checkGamePacket(L)
	L.Push(g.luaID)
	return 1
}

// game_packet:name()
func gamePacketGetName(L *lua.LState) int {
	g := checkGamePacket(L)
	L.Push(g.luaName)
	return 1
}
