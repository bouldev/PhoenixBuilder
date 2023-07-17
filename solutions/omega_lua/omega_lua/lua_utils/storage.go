package lua_utils

import (
	"fmt"
	"os"
	"path"
	"phoenixbuilder/fastbuilder/lib/utils/file_wrapper"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type StorageModule struct {
	basePath string
}

func NewStorageModule(basePath string) *StorageModule {
	s := &StorageModule{
		basePath: basePath,
	}
	os.MkdirAll(basePath, 0755)
	return s
}

func (s *StorageModule) MakeLValue(L *lua.LState) lua.LValue {
	luaModule := L.NewTable()
	luaModule = L.SetFuncs(luaModule, map[string]lua.LGFunction{
		"read":   s.read,
		"save":   s.save,
		"remove": s.remove,
	})
	return luaModule
}

func (s *StorageModule) cleanPath(L *lua.LState) string {
	p := L.CheckString(1)
	p = path.Clean(p)
	if path.IsAbs(p) {
		L.RaiseError("path must be relative")
	}
	p = path.Join(s.basePath, p)
	return p
}

func (s *StorageModule) read(L *lua.LState) int {
	fp := s.cleanPath(L)
	var data any
	if err := file_wrapper.GetJsonData(fp, &data); err != nil {
		L.Push(lua.LNil)
		return 1
	}
	ld := ToLuaValue(L, data)
	L.Push(ld)
	return 1
}

func CleanMapKey(v interface{}) any {
	if mv, ok := v.(map[interface{}]interface{}); ok {
		r := map[string]interface{}{}
		for k, v := range mv {
			r[fmt.Sprintf("%v", k)] = CleanMapKey(v)
		}
		return r
	} else {
		return v
	}
}

func (s *StorageModule) save(L *lua.LState) int {
	fp := s.cleanPath(L)
	// make sure the dir exists
	os.MkdirAll(path.Dir(fp), 0755)
	data := L.Get(2)
	goData := gluamapper.ToGoValue(data, gluamapper.Option{
		NameFunc: func(s string) string { return s },
	})
	goData = CleanMapKey(goData)
	err := file_wrapper.WriteJsonData(fp, goData)
	if err != nil {
		L.RaiseError(err.Error())
	}
	return 0
}

func (s *StorageModule) remove(L *lua.LState) int {
	fp := s.cleanPath(L)
	os.Remove(fp)
	return 0
}
