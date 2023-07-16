package lua_utils

type LuaConfigRaw struct {
	Config         interface{}
	OnConfigUpdate func(interface{})
}
