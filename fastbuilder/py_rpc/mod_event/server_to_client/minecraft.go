package mod_event_server_to_client

import (
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft"
)

// Minecraft Package
type Minecraft struct{ mei.Default }

// Return the package name of m
func (m *Minecraft) PackageName() string {
	return "Minecraft"
}

// Return a pool/map that contains all the module of m
func (m *Minecraft) ModulePool() map[string]mei.Module {
	return map[string]mei.Module{
		"aiCommand":     &minecraft.AICommand{Module: &mei.DefaultModule{}},
		"pet":           &minecraft.Pet{Module: &mei.DefaultModule{}},
		"chatPhrases":   &minecraft.ChatPhrases{Module: &mei.DefaultModule{}},
		"achievement":   &minecraft.Achievement{Module: &mei.DefaultModule{}},
		"chatExtension": &minecraft.ChatExtension{Module: &mei.DefaultModule{}},
	}
}
