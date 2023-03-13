package environment

// This package imports only external packages to avoid import cycle.
import (
	"phoenixbuilder/fastbuilder/environment/interfaces"
)

type LoginInfo struct {
	Token          string
	ServerCode     string
	ServerPasscode string
}

type PBEnvironment struct {
	LoginInfo
	IsDebug                   bool
	ScriptBridge              interface{}
	ScriptHolder              interface{}
	FunctionHolder            interfaces.FunctionHolder
	FBUCUsername              string
	WorldChatChannel          chan []string
	FBAuthClient              interface{}
	GlobalFullConfig          interface{}
	RespondUser               string
	CommandSender             interfaces.CommandSender
	Connection                interface{}
	UQHolder                  interface{}
	NewUQHolder               interface{} // for blockNBT
	TaskHolder                interface{}
	OmegaHolder               interface{}
	OmegaAdaptorHolder        interface{}
	ActivateTaskStatus        chan bool
	Uid                       string
	ExternalConnectionHandler interface{}
	Destructors               []func()
	isStopping                bool
	stoppedWaiter             chan struct{}
	CertSigning               bool
	LocalKey                  string
	LocalCert                 string
	LRUMemoryChunkCacher      interface{}
	ChunkFeeder               interface{}
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
