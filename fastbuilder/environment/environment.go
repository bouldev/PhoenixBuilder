package environment

// This package imports nothing to avoid import cycle.

type LoginInfo struct {
	Token string
	Version string
	ServerCode string
	ServerPasscode string
}

type PBEnvironment struct {
	LoginInfo
	IsDebug bool
	ScriptBridge interface{}
	ScriptHolder interface{}
	FunctionHolder interface{}
	FBUCUsername string
	WorldChatChannel chan []string
	FBAuthClient interface{}
	GlobalFullConfig interface{}
	RespondUser string
	Connection interface{}
	TaskHolder interface{}
	ActivateTaskStatus chan bool
	Uid string
}