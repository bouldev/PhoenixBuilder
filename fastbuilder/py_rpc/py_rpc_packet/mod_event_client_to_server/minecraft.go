package mod_event_server_to_client

import (
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/interface/mod_event"
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/mod_event_client_to_server/minecraft"
)

// Minecraft Package
type Minecraft struct{ mod_event.Default }

// Return the package name of m
func (m *Minecraft) PackageName() string {
	return "Minecraft"
}

// Return a pool/map that contains all the module of m
func (m *Minecraft) ModulePool() map[string]mod_event.Module {
	return map[string]mod_event.Module{
		"vipEventSystem": &minecraft.VIPEventSystem{Module: &mod_event.DefaultModule{}},
	}
}
