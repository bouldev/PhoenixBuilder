package command

/*
void *allocateRequestString();
void freeRequestString(void*);
*/
import "C"
import "unsafe"
import "phoenixbuilder/fastbuilder/environment"
import "sync"

var AdditionalChatCb func(string) = func(_ string) {}
var AdditionalTitleCb func(string) = func(_ string) {}

func AllocateRequestString() *string {
	return (*string)(C.allocateRequestString())
}

func FreeRequestString(str string) {
	C.freeRequestString(unsafe.Pointer(&str))
}

func FreeRequestStringPtr(str *string) {
	C.freeRequestString(unsafe.Pointer(str))
}

type CommandSender struct {
	env *environment.PBEnvironment
	UUIDMap sync.Map
	BlockUpdateSubscribeMap sync.Map
}

func InitCommandSender(env *environment.PBEnvironment) *CommandSender {
	env.CommandSender=&CommandSender {
		env: env,
	}
	return env.CommandSender.(*CommandSender)
}