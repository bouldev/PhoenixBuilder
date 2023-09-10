package environment

// This package imports only external packages to avoid import cycle.
import (
	"phoenixbuilder/fastbuilder/environment/interfaces"
	fbauth "phoenixbuilder/fastbuilder/pv4"
)

type LoginInfo struct {
	Token          string
	Username       string
	Password       string
	ServerCode     string
	ServerPasscode string
}

type PBEnvironment struct {
	LoginInfo
	IsDebug               bool
	FunctionHolder        interfaces.FunctionHolder
	FBAuthClient          interface{}
	GlobalFullConfig      interface{}
	RespondTo             string
	Connection            interface{}
	GetCheckNumEverPassed bool
	Resources             interface{}
	ResourcesUpdater      interface{}
	GameInterface         interfaces.GameInterface
	TaskHolder            interface{}
	LRUMemoryChunkCacher  interface{}
	ChunkFeeder           interface{}
	ClientOptions         *fbauth.ClientOptions
}
