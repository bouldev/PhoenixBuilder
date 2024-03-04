package mod_event_server_to_client

import (
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/general_interface"
)

// Return a pool/map that contains
// all the package of ModEventS2C
func PackagePool() map[string]general_interface.Package {
	return map[string]general_interface.Package{
		"Minecraft": &Minecraft{},
	}
}
