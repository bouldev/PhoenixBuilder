package game_packet

import (
	"phoenixbuilder/solutions/luaMega/omega_lua/mux_pumper"
)

type LuaGoPackets interface {
	GetMCPacketNameIDMapping() mux_pumper.MCPacketNameIDMapping
}
