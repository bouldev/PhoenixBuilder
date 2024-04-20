package mod_event_client_to_server

import (
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
)

// Return a pool/map that contains
// all the package of ModEventCS2
func PackagePool() map[string]mei.Package {
	return map[string]mei.Package{
		"Minecraft": &Minecraft{Default: mei.Default{}},
	}
}
