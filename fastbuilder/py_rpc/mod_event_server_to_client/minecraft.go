package mod_event_server_to_client

import (
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/general_interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/minecraft"
)

// Minecraft Package
type Minecraft struct{ general_interface.Module }

// Return the package name of m
func (m *Minecraft) PackageName() string {
	return "Minecraft"
}

// Return a pool/map that contains all the module of m
func (m *Minecraft) ModulePool() map[string]general_interface.Module {
	return map[string]general_interface.Module{
		"aiCommand": &minecraft.AICommand{},
	}
}
