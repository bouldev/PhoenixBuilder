package script_holder

import (
	"fmt"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/script_engine/bridge/kickstarter"
)

type ScriptHolder struct {
	allScripts map[string]func()
}

func InitScriptHolder(hostbridge bridge.HostBridge) *ScriptHolder {
	sh:=&ScriptHolder {}
	function.RegisterFunction(&function.Function {
		Name: "script",
		OwnedKeywords: []string {"script"},
		FunctionType: function.FunctionTypeSimple,
		SFArgumentTypes: []byte { function.SimpleFunctionArgumentMessage },
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			str:=args[0].(string)
			sh.LoadScript(str, hostbridge)
		},
	})
	return sh
}

func (sh *ScriptHolder) Destroy() {
	for _, scb := range sh.allScripts {
		scb()
	}
}

func (sh *ScriptHolder) LoadScript(script_path string,hostbridge bridge.HostBridge) bool {
	// TODO: Normalize script_path
	
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

