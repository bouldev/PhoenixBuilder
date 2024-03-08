package mod_event_server_to_client

import (
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/interface/mod_event"
)

// Return a pool/map that contains
// all the package of ModEventCS2
func PackagePool() map[string]mod_event.Package {
	return map[string]mod_event.Package{
		"Minecraft": &Minecraft{Default: mod_event.Default{}},
	}
}
