package environment

// This package imports only external packages to avoid import cycle.
import "phoenixbuilder/fastbuilder/environment/interfaces"

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
	FunctionHolder interfaces.FunctionHolder
	FBUCUsername string
	WorldChatChannel chan []string
	FBAuthClient interface{}
	GlobalFullConfig interface{}
	RespondUser string
	CommandSender interfaces.CommandSender
	Connection interface{}
	TaskHolder interface{}
	ActivateTaskStatus chan bool
	Uid string
	ExternalConnectionHandler interface{}
	Destructors []func()
}