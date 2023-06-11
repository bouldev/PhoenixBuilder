package core

import (
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
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
