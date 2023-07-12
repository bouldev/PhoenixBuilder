package submodule_holder

import lua "github.com/yuin/gopher-lua"

type SubModuleHolder struct {
	subModules map[string]lua.LValue
}

func NewSubModuleHolder(subModules map[string]lua.LValue) *SubModuleHolder {
	return &SubModuleHolder{
		subModules: subModules,
	}
}

func (h *SubModuleHolder) Loader(L *lua.LState) int {
	omegaModule := L.NewTable()
	for name, module := range h.subModules {
		L.SetTable(omegaModule, lua.LString(name), module)
	}
	L.Push(omegaModule)
	return 1
}
