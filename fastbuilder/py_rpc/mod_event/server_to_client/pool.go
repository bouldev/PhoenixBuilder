package mod_event_server_to_client

import (
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
)

// Return a pool/map that contains
// all the package of ModEventS2C
func PackagePool() map[string]mei.Package {
	return map[string]mei.Package{
		"Minecraft": &Minecraft{Default: mei.Default{}},
	}
}
