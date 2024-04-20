package minecraft

import (
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft/ai_command"
)

// 魔法指令
type AICommand struct{ mei.Module }

// Return the module name of a
func (a *AICommand) ModuleName() string {
	return "aiCommand"
}

// Return a pool/map that contains all the event of a
func (a *AICommand) EventPool() map[string]mei.Event {
	return map[string]mei.Event{
		"ExecuteCommandOutputEvent": &ai_command.ExecuteCommandOutputEvent{},
		"AfterExecuteCommandEvent":  &ai_command.AfterExecuteCommandEvent{},
		"AvailableCheckFailed":      &ai_command.AvailableCheckFailed{},
	}
}
