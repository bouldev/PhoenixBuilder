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
	indexTable := L.GetField(mt, "__index")
	for k, v := range o.injectLaters {
		L.SetField(indexTable, k, v)
	}
	L.SetMetatable(luaGamePacket, mt)
	o.luaSelf = luaGamePacket
	return luaGamePacket
}

func (o *CommonLuaGoObject) ToUserData(L *lua.LState) lua.LValue {
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

func (o *CommonLuaGoObject) FromLuaValue(lt lua.LValue) error {
	o.data = gluamapper.ToGoValue(lt, gluamapper.Option{
		NameFunc: func(s string) string { return s },
	})
	return nil
}

func (o *CommonLuaGoObject) GetData() interface{} {
	return o.data
}

func RegisterCommonLuaGoObject(L *lua.LState) {
	mt := L.NewTypeMetatable("common_go_object")
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"user_data":      CommonGoObjectToUserData,
		"from_user_data": CommonGoObjectFromUserData,
		"json_str":       CommonGoObjectToJsonStr,
	}))
}

func CheckCommonGoObject(L *lua.LState) *CommonLuaGoObject {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*CommonLuaGoObject); ok {
		return v
	}
	L.ArgError(1, "go object")
	return nil
}

func CommonGoObjectToUserData(L *lua.LState) int {
	g := CheckCommonGoObject(L)
	L.Push(g.ToUserData(L))
	return 1
}

func CommonGoObjectToJsonStr(L *lua.LState) int {
	g := CheckCommonGoObject(L)
	L.Push(lua.LString(g.ToJsonStr()))
	return 1
}

func CommonGoObjectFromUserData(L *lua.LState) int {
	g := CheckCommonGoObject(L)
	v := L.Get(2)
	if ud, ok := v.(*lua.LUserData); ok {
		g.data = ud.Value
	} else {
		if err := g.FromLuaValue(v); err != nil {
			L.RaiseError(err.Error())
		}
	}

	L.Push(g.luaSelf)
	return 1
}
