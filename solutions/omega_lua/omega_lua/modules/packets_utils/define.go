package packets_utils

import (
	"phoenixbuilder/solutions/omega_lua/omega_lua/mux_pumper"
)

type LuaGoPackets interface {
	GetMCPacketNameIDMapping() mux_pumper.MCPacketNameIDMapping
}
