package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// AddEntity is sent by the server to the client. Its function is not entirely clear: It does not add an
// entity in the sense of an in-game entity, but has to do with the ECS that Minecraft uses.
type AddEntity struct {
	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		EntityNetworkID is the network ID of the entity that should be added.

		For netease, the data type of this field is uint32,
		but on standard minecraft, this is uint64.
	*/
	EntityNetworkID uint32
	// EntityNetworkID uint64
}

// ID ...
func (pk *AddEntity) ID() uint32 {
	return IDAddEntity
}

func (pk *AddEntity) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Varuint32(&pk.EntityNetworkID)
		// io.Varuint64(&pk.EntityNetworkID)
	}
}
