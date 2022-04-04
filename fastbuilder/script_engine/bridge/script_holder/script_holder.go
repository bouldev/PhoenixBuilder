package script_holder

import (
	"fmt"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/script_engine/bridge/kickstarter"
	"phoenixbuilder/fastbuilder/environment"
)

type ScriptHolder struct {
	allScripts map[string]func()
}

func InitScriptHolder(env *environment.PBEnvironment) *ScriptHolder {
	sh:=&ScriptHolder {}
	sh.allScripts=map[string]func() {}
	fh:=env.FunctionHolder.(*function.FunctionHolder)
	fh.RegisterFunction(&function.Function {
		Name: "script",
		OwnedKeywords: []string {"script"},
		FunctionType: function.FunctionTypeSimple,
		SFArgumentTypes: []byte { function.SimpleFunctionArgumentMessage },
		SFMinSliceLen: 1,
		FunctionContent: func(env *environment.PBEnvironment,args []interface{}) {
			str:=args[0].(string)
			sh.LoadScript(str, env)
		},
	})
	return sh
}

func (sh *ScriptHolder) Destroy() {
	for _, scb := range sh.allScripts {
		scb()
	}
}

func (sh *ScriptHolder) LoadScript(script_path string, env *environment.PBEnvironment) bool {
	// TODO: Normalize script_path
	hostbridge:=env.ScriptBridge.(bridge.HostBridge)
	stopCb, found := sh.allScripts[script_path]
	if found {
		fmt.Printf("Reloading script %s\n",script_path)
		stopCb()
	}
	tStop, err:=script_kickstarter.LoadScript(script_path, hostbridge)
	if(err!=nil) {
		fmt.Printf("Failed to load script %s: %v.\n",script_path,err)
		if(found) {
			delete(sh.allScripts, script_path)
		}
		return false
	}
	sh.allScripts[script_path]=tStop
	return true
}

