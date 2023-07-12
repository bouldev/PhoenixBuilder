package backend

import lua "github.com/yuin/gopher-lua"

// backend module: user_input/log/print
type OmegaBackendModule struct {
	goImplements LuaGoBackends
}

type LuaGoBackends interface {
	Log(string)
}

func NewOmegaBackendModule(goImplements LuaGoBackends) *OmegaBackendModule {
	return &OmegaBackendModule{
		goImplements: goImplements,
	}
}

func (m *OmegaBackendModule) MakeLValue(L *lua.LState) lua.LValue {
	return L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"log":   m.luaGoBackendLog,
		"print": m.luaGoBackendLog,
	})
}

// backend.log(info:string)
func (m *OmegaBackendModule) luaGoBackendLog(L *lua.LState) int {
	m.goImplements.Log(L.ToString(-1))
	return 0
}
