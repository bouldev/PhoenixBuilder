package minecraft

import (
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft/pet"
)

// 宠物(狐狸, etc.)
type Pet struct{ mei.Module }

// Return the module name of p
func (p *Pet) ModuleName() string {
	return "pet"
}

// Return a pool/map that contains all the event of p
func (p *Pet) EventPool() map[string]mei.Event {
	return map[string]mei.Event{
		"check_server_contain_pet": &pet.CheckServerContainPet{},
		"pet_skill_remove_beacon":  &pet.PetSkillRemoveBeacon{},
	}
}
