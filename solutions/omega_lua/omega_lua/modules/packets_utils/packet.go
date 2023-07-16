package packets_utils

import (
	"bytes"
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"

	luar "layeh.com/gopher-luar"

	lua "github.com/yuin/gopher-lua"
)

// 游戏包在lua中的体现
type GamePacket struct {
	goPacket packet.Packet
	luaSelf  lua.LValue
	luaName  lua.LString
	luaID    lua.LNumber
}

// 初始化一个lua中的游戏包
// 你需要传入包本身 在lua中的名字 在lua中的id
// 也就是goPacket luaName luaID
// 会返回给你这个新的包的指针
func (g *OmegaPacketsModule) NewGamePacket(goPacket packet.Packet, L *lua.LState) lua.LValue {
	return (&GamePacket{
		goPacket: goPacket,
		luaName:  g.MCPacketIDToLuaName[goPacket.ID()],
		luaID:    g.MCPacketIDToLuaInt[goPacket.ID()],
	}).MakeLValue(L)
}

// 继承game_packet内容
// 方便用户访问
func (g *GamePacket) MakeLValue(L *lua.LState) lua.LValue {
	luaGamePacket := L.NewUserData()
	luaGamePacket.Value = g
	L.SetMetatable(luaGamePacket, L.GetTypeMetatable("game_packet"))
	g.luaSelf = luaGamePacket
	return luaGamePacket
}

// 注册一个gamePacket包 并且设置两个函数
// 一个为id 一个为name
// 以方便子表调用
func registerGamePacket(L *lua.LState) {
	mt := L.NewTypeMetatable("game_packet")
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"id":        gamePacketGetID,
		"name":      gamePacketGetName,
		"user_data": gamePacketToUserData,
		"json_str":  gamePacketToJsonStr,
	}))
}

// 检查这个包是否为gamePacket包
func checkGamePacket(L *lua.LState) *GamePacket {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GamePacket); ok {
		return v
	}
	L.ArgError(1, "game packet expected")
	return nil
}

// 在lua代码中 game_packet:id() 即可获取这个包的id
func gamePacketGetID(L *lua.LState) int {
	g := checkGamePacket(L)
	L.Push(g.luaID)
	return 1
}

// 在lua代码中game_packet:name()  即可获取这个包的name
func gamePacketGetName(L *lua.LState) int {
	g := checkGamePacket(L)
	L.Push(g.luaName)
	return 1
}

func gamePacketToUserData(L *lua.LState) int {
	g := checkGamePacket(L)
	L.Push(luar.New(L, g.goPacket))
	return 1
}

func gamePacketToJsonStr(L *lua.LState) int {
	g := checkGamePacket(L)
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err := enc.Encode(g.goPacket)
	if err != nil {
		L.RaiseError(err.Error())
	}
	L.Push(lua.LString(buf.String()))
	return 1
}
