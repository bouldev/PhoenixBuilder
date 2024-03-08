package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/interface/mod_event"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/minecraft/pet"
)

// 网易宠物(狐狸, etc.)
type Pet struct{ mod_event.Module }

// Return the module name of p
func (p *Pet) ModuleName() string {
	return "pet"
}

// Return a pool/map that contains all the event of p
func (p *Pet) EventPool() map[string]mod_event.Event {
	return map[string]mod_event.Event{
		"check_server_contain_pet": &pet.CheckServerContainPet{},
	}
}
