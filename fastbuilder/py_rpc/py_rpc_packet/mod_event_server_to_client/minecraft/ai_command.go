package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/interface/mod_event"
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/mod_event_server_to_client/minecraft/ai_command"
)

// 魔法指令
type AICommand struct{ mod_event.Module }

// Return the module name of a
func (a *AICommand) ModuleName() string {
	return "aiCommand"
}

// Return a pool/map that contains all the event of a
func (a *AICommand) EventPool() map[string]mod_event.Event {
	return map[string]mod_event.Event{
		"ExecuteCommandOutputEvent": &ai_command.ExecuteCommandOutputEvent{},
		"AfterExecuteCommandEvent":  &ai_command.AfterExecuteCommandEvent{},
		"AvailableCheckFailed":      &ai_command.AvailableCheckFailed{},
	}
}
