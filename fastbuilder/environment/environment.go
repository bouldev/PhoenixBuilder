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
	IsDebug                   bool
	ScriptBridge              interface{}
	ScriptHolder              interface{}
	FunctionHolder            interfaces.FunctionHolder
	FBAuthClient              interface{}
	GlobalFullConfig          interface{}
	RespondUser               string
	Connection                interface{}
	UQHolder                  interface{}
	Resources                 interface{}
	ResourcesUpdater          interface{}
	GameInterface             interfaces.GameInterface
	TaskHolder                interface{}
	OmegaHolder               interface{}
	OmegaAdaptorHolder        interface{}
	ActivateTaskStatus        chan bool
	ExternalConnectionHandler interface{}
	Destructors               []func()
	isStopping                bool
	stoppedWaiter             chan struct{}
	LRUMemoryChunkCacher      interface{}
	ChunkFeeder               interface{}
	ClientOptions             *fbauth.ClientOptions
}

func (env *PBEnvironment) Stop() {
	if env.isStopping {
		return
	}
	//fmt.Println("stopping")
	env.stoppedWaiter = make(chan struct{})
	env.isStopping = true
	for _, fn := range env.Destructors {
		fn()
	}
	//fmt.Println("stopped")
	close(env.stoppedWaiter)
}

func (env *PBEnvironment) WaitStopped() {
	//fmt.Println("waitting stopped")
	<-env.stoppedWaiter
}
