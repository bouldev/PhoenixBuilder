package lua_utils

import (
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

func ToLuaValue(L *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}
	if lval, ok := value.(lua.LValue); ok {
		return lval
	}
	switch val := reflect.ValueOf(value); val.Kind() {
	case reflect.Bool:
		return lua.LBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(float64(val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(float64(val.Uint()))
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(val.Float())
	case reflect.Chan:
		if val.IsNil() {
			return lua.LNil
		}
		fallthrough
	case reflect.Map:
		ud := L.NewTable()
		// iter and fill
		r := val.MapRange()
		for r.Next() {
			ud.RawSetString(r.Key().String(), ToLuaValue(L, r.Value().Interface()))
		}
		return ud
	case reflect.Array, reflect.Slice:
		ud := L.NewTable()
		items := val.Len()
		for idx := 0; idx < items; idx++ {
			ud.RawSetInt(idx+1, ToLuaValue(L, val.Index(idx).Interface()))
		}
		return ud
	case reflect.Ptr:
		if val.IsNil() {
			return lua.LNil
		}
		return ToLuaValue(L, val.Elem().Interface())
	case reflect.Struct:
		ud := L.NewTable()
		// iter and fill
		for i := 0; i < val.NumField(); i++ {
			ud.RawSetString(val.Type().Field(i).Name, ToLuaValue(L, val.Field(i).Interface()))
		}
		return ud
	case reflect.Func:
		if val.IsNil() {
			return lua.LNil
		}
		fallthrough
	case reflect.String:
		return lua.LString(val.String())
	default:
		ud := L.NewUserData()
		ud.Value = val.Interface()
		return ud
	}
}

func UserDataToLuaValue(L *lua.LState) int {
	ud := L.CheckUserData(1)
	L.Push(ToLuaValue(L, ud.Value))
	return 1
}
