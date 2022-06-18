package commands

import "phoenixbuilder/fastbuilder/environment"
import "sync"

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
