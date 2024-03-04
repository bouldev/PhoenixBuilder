package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/general_interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/minecraft/ai_command"
)

// 魔法指令
type AICommand struct{ general_interface.Event }

// Return the module name of a
func (a *AICommand) ModuleName() string {
	return "aiCommand"
}

// Return a pool/map that contains all the event of a
func (a *AICommand) EventPool() map[string]general_interface.Event {
	return map[string]general_interface.Event{
		"ExecuteCommandOutputEvent": &ai_command.ExecuteCommandOutputEvent{},
		"AfterExecuteCommandEvent":  &ai_command.AfterExecuteCommandEvent{},
		"AvailableCheckFailed":      &ai_command.AvailableCheckFailed{},
	}
}
