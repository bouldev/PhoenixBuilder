package lua_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type CommonLuaGoObject struct {
	data         interface{}
	luaSelf      lua.LValue
	injectLaters map[string]lua.LValue
}

func NewCommonLuaGoObject(data interface{}, injectLaters map[string]lua.LValue) *CommonLuaGoObject {
	return &CommonLuaGoObject{
		data:         data,
		injectLaters: injectLaters,
	}
}

func (o *CommonLuaGoObject) MakeLValue(L *lua.LState) lua.LValue {
	luaGamePacket := L.NewUserData()
	luaGamePacket.Value = o
	mt := L.GetTypeMetatable("common_go_object")
	for k, v := range o.injectLaters {
		L.SetField(mt, k, v)
	}
	L.SetMetatable(luaGamePacket, mt)
	o.luaSelf = luaGamePacket
	return luaGamePacket
}

func (o *CommonLuaGoObject) ToLuaTable(L *lua.LState) lua.LValue {
	return luar.New(L, o.data)
}

func (o *CommonLuaGoObject) ToJsonStr() string {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err := enc.Encode(o.data)
	if err != nil {
		return fmt.Sprintf("data cannot present as a json str: %v", err.Error())
	}
	return buf.String()
}

func (o *CommonLuaGoObject) FromLuaTable(lt *lua.LTable) error {
	return gluamapper.Map(lt, &o.data)
}

func RegisterCommonLuaGoObject(L *lua.LState) {
	mt := L.NewTypeMetatable("common_go_object")
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"lua_table":      commonGoObjectToLuaTable,
		"from_lua_table": commonGoObjectFromLuaTable,
		"json_str":       commonGoObjectToJsonStr,
	}))
}

func checkCommonGoObject(L *lua.LState) *CommonLuaGoObject {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*CommonLuaGoObject); ok {
		return v
	}
	L.ArgError(1, "go object")
	return nil
}

func commonGoObjectToLuaTable(L *lua.LState) int {
	g := checkCommonGoObject(L)
	L.Push(g.ToLuaTable(L))
	return 1
}

func commonGoObjectToJsonStr(L *lua.LState) int {
	g := checkCommonGoObject(L)
	L.Push(lua.LString(g.ToJsonStr()))
	return 1
}

func commonGoObjectFromLuaTable(L *lua.LState) int {
	g := checkCommonGoObject(L)
	lt := L.ToTable(2)
	if err := g.FromLuaTable(lt); err != nil {
		L.RaiseError(err.Error())
	}
	L.Push(g.luaSelf)
	return 1
}
