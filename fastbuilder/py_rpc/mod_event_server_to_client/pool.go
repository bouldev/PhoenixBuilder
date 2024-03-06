package mod_event_server_to_client

import (
	"phoenixbuilder/fastbuilder/py_rpc/general_interface/mod_event"
)

// Return a pool/map that contains
// all the package of ModEventS2C
func PackagePool() map[string]mod_event.Package {
	return map[string]mod_event.Package{
		"Minecraft": &Minecraft{Default: mod_event.Default{}},
	}
}
