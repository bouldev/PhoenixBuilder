package packets_utils

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// 包装了 LuaGoPackets 接口的 OmegaPacketsModule 结构体
// 存储了mc包名到go与lua的id和string的字典
type OmegaPacketsModule struct {
	goImplements          LuaGoPackets           // 实现了 LuaGoPackets 接口的对象
	MCPacketNameIDMapping map[string]uint32      // Minecraft 包名到包 ID 的映射
	MCPacketIDToLuaName   map[uint32]lua.LString // Minecraft 包 ID 到 Lua 包名的映射
	MCPacketIDToLuaInt    map[uint32]lua.LNumber // Minecraft 包 ID 到 Lua 包 ID 的映射
}

// 创建一个新的 OmegaPacketsModule 对象
func NewOmegaPacketsModule(goImplements LuaGoPackets) *OmegaPacketsModule {
	//获取游戏包 的Id与对应的string的对应表
	mapping := goImplements.GetMCPacketNameIDMapping()
	//初始化对象
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

// 生成 Lua 中的包表
func (m *OmegaPacketsModule) MakeLValue(L *lua.LState) lua.LValue {
	packetModule := L.NewTable()
	//goId表示在go中的包数字代号 Name同理
	for goId, goName := range m.MCPacketIDToLuaName {
		//获取对应在lua代码中的id与名字
		luaID := m.MCPacketIDToLuaInt[goId]
		luaName := m.MCPacketIDToLuaName[goId]
		//看上去是以三种形式存入:{包名:包名}是索要包名本身
		//{no包名:!包名}是在all模式之下不要xxxx包
		//{ID包名:包id}是索要这个包的id
		L.SetTable(packetModule, luaName, luaName)
		L.SetTable(packetModule, lua.LString("no"+goName), lua.LString("!"+goName))
		L.SetTable(packetModule, lua.LString("ID"+goName), luaID)
	}
	//所有包
	L.SetTable(packetModule, lua.LString("all"), lua.LString("all"))
	allNames := make([]string, 0, len(m.MCPacketNameIDMapping))
	for name := range m.MCPacketNameIDMapping {
		allNames = append(allNames, name)
	}
	L.SetTable(packetModule, lua.LString("all_names"), lua.LString(
		"["+strings.Join(allNames, ",")+"]",
	))
	//同理获得包对应的json数据 并且存入to_json_string_slow这个键值对中
	// L.SetTable(packetModule, lua.LString("to_json_string_slow"), L.NewFunction(m.luaGoPacketsToJSONString))
	// L.SetTable(packetModule, lua.LString("to_user_data"), L.NewFunction(m.luaGoPacketsToUserData))
	registerGamePacket(L)
	return packetModule
}

// // packet.to_json_string(pk:packet)
// func (m *OmegaPacketsModule) luaGoPacketsToJSONString(L *lua.LState) int {
// 	pk := checkGamePacket(L)
// 	jsonBytes, _ := json.Marshal(pk.goPacket)
// 	L.Push(lua.LString(jsonBytes))
// 	return 1
// }

// // 将包转化为 userdata
// func (m *OmegaPacketsModule) luaGoPacketsToUserData(L *lua.LState) int {
// 	pk := checkGamePacket(L)
// 	luaUserData := luar.New(L, pk.goPacket)
// 	L.Push(luaUserData)
// 	return 1
// }
