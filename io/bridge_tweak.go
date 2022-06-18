// +build is_tweak

package io

import (
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/io/commands"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

import "C"

var refid uint64=1
var refs map[uint64]interface{}=map[uint64]interface{} {}

//export phoenixbuilder_create
func phoenixbuilder_create() uint64 {
	env:=&environment.PBEnvironment {}
	env.ActivateTaskStatus=make(chan bool)
	env.TaskHolder=task.NewTaskHolder()
	env.LoginInfo=environment.LoginInfo {
		Token: "",
		ServerCode: "LOCAL",
		ServerPasscode: "",
	}
	commands.InitCommandSender(env)
	functionHolder := function.NewFunctionHolder(env)
	env.FunctionHolder=functionHolder
	function.InitInternalFunctions(functionHolder)
	retval:=refid
	refid++
	refs[retval]=env
	return retval
}

// Caller free !
//export phoenixbuilder_execute
func phoenixbuilder_execute(ref uint64, command *C.char) bool {
	env:=refs[ref].(*environment.PBEnvironment)
	found:=env.FunctionHolder.Process(C.GoString(command))
	return found
}

//export phoenixbuilder_destroy
func phoenixbuilder_destroy(ref uint64) {
	env:=refs[ref].(*environment.PBEnvironment)
	env.Stop()
	env.WaitStopped()
	delete(refs, ref)
	return
}

// caller free
//export phoenixbuilder_command_output
func phoenixbuilder_command_output(ref uint64,uuid *C.char, succ bool, message *C.char, param *C.char) {
	env:=refs[ref].(*environment.PBEnvironment)
	item, found:=env.CommandSender.GetUUIDMap().LoadAndDelete(C.GoString(uuid))
	if(!found) {
		return
	}
	succCount:=0
	if(succ) {
		succCount++
	}
	rit:=item.(chan *packet.CommandOutput)
	rit<-&packet.CommandOutput {
		SuccessCount: uint32(succCount),
		OutputMessages: []protocol.CommandOutputMessage {
			protocol.CommandOutputMessage {
				Success: succ,
				Message: C.GoString(message),
				Parameters: []string{C.GoString(param)},
			},
		},
	}
}
	
