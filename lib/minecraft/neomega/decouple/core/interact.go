package core

import (
	minecraft "fastbuilder-core/lib/minecraft/gophertunnel"
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol/packet"
	"fastbuilder-core/lib/minecraft/neomega/omega"
)

type InteractCore struct {
	*minecraft.Conn
}

func (i *InteractCore) SendPacket(packet packet.Packet) {
	i.WritePacket(packet)
}

func NewInteractCore(conn *minecraft.Conn) omega.InteractCore {
	return &InteractCore{Conn: conn}
}
