package fastbuilder

import (
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/function"
	fbauth "phoenixbuilder/fastbuilder/pv4"
	script_bridge "phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/script_engine/bridge/script_holder"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/mirror/io/global"
	"phoenixbuilder/mirror/io/lru"
)

func create_environment() *environment.PBEnvironment {
	clientOption := fbauth.MakeDefaultClientOptions()
	clientOption.AuthServer = args.AuthServer
	clientOption.RespondUserOverride = args.CustomGameName
	env := &environment.PBEnvironment{
		ClientOptions: clientOption,
	}
	env.UQHolder = nil
	env.Resources = nil
	env.ActivateTaskStatus = make(chan bool)
	env.TaskHolder = fbtask.NewTaskHolder()
	functionHolder := function.NewFunctionHolder(env)
	env.FunctionHolder = functionHolder
	env.Destructors = []func(){}
	hostBridgeGamma := &script_bridge.HostBridgeGamma{}
	hostBridgeGamma.Init()
	hostBridgeGamma.HostQueryExpose = map[string]func() string{
		"server_code": func() string {
			return env.LoginInfo.ServerCode
		},
		"fb_version": func() string {
			return args.FBVersion
		},
		"uc_username": func() string {
			return env.FBAuthClient.(*fbauth.Client).FBUCUsername
		},
	}
	for _, key := range args.CustomSEUndefineConsts {
		_, found := hostBridgeGamma.HostQueryExpose[key]
		if found {
			delete(hostBridgeGamma.HostQueryExpose, key)
		}
	}
	for key, val := range args.CustomSEConsts {
		hostBridgeGamma.HostQueryExpose[key] = func() string { return val }
	}
	env.ScriptBridge = hostBridgeGamma
	scriptHolder := script_holder.InitScriptHolder(env)
	env.ScriptHolder = scriptHolder
	if args.StartupScript != "" {
		scriptHolder.LoadScript(args.StartupScript, env)
	}
	env.Destructors = append(env.Destructors, func() {
		scriptHolder.Destroy()
	})
	hostBridgeGamma.HostRemoveBlock()
	env.LRUMemoryChunkCacher = lru.NewLRUMemoryChunkCacher(12, false)
	env.ChunkFeeder = global.NewChunkFeeder()
	return env
}

// Shouldn't be called when running a debug client
func ConfigRealEnvironment(token string, server_code string, server_password string, username string, password string) *environment.PBEnvironment {
	env := create_environment()
	env.LoginInfo = environment.LoginInfo{
		Token:          token,
		ServerCode:     server_code,
		ServerPasscode: server_password,
		Username: username,
		Password: password,
	}
	env.FBAuthClient = fbauth.CreateClient(env.ClientOptions)
	return env
}

func ConfigDebugEnvironment() *environment.PBEnvironment {
	env := create_environment()
	env.IsDebug = true
	env.LoginInfo = environment.LoginInfo{
		ServerCode: "[DEBUG]",
	}
	return env
}

func DestroyEnv(env *environment.PBEnvironment) {
	env.Stop()
	env.WaitStopped()
	env.Connection.(*minecraft.Conn).Close()
}
