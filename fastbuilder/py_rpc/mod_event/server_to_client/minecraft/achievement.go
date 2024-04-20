package minecraft

import (
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft/achievement"
)

// 成就
type Achievement struct{ mei.Module }

// Return the module name of a
func (a *Achievement) ModuleName() string {
	return "achievement"
}

// Return a pool/map that contains all the event of a
func (a *Achievement) EventPool() map[string]mei.Event {
	return map[string]mei.Event{
		"InitInformation": &achievement.InitInformation{},
	}
}
