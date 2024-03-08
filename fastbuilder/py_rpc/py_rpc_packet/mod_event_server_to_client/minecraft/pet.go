package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/interface/mod_event"
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/mod_event_server_to_client/minecraft/pet"
)

// 宠物(狐狸, etc.)
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
