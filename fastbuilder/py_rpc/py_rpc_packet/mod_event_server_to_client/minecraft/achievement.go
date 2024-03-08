package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/interface/mod_event"
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/mod_event_server_to_client/minecraft/achievement"
)

// 成就
type Achievement struct{ mod_event.Module }

// Return the module name of a
func (a *Achievement) ModuleName() string {
	return "achievement"
}

// Return a pool/map that contains all the event of a
func (a *Achievement) EventPool() map[string]mod_event.Event {
	return map[string]mod_event.Event{
		"InitInformation": &achievement.InitInformation{},
	}
}
