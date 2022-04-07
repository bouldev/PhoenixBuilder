package defines

import (
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol/packet"
)

type Adaptor interface {
	Read() packet.Packet
	Write(packet.Packet)
	GetInitUQHolderCopy() *uqHolder.UQHolder
}
