package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// RemoveEntity is sent by the server to the client. Its function is not entirely clear: It does not remove an
// entity in the sense of an in-game entity, but has to do with the ECS that Minecraft uses.
type RemoveEntity struct {
	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		EntityNetworkID is the network ID of the entity that should be removed.

		For netease, the data type of this field is uint32,
		but on standard minecraft, this is uint64.
	*/
	EntityNetworkID uint32
	// EntityNetworkID uint64
}

// ID ...
func (pk *RemoveEntity) ID() uint32 {
	return IDRemoveEntity
}

func (pk *RemoveEntity) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Varuint32(&pk.EntityNetworkID)
		// io.Varuint64(&pk.EntityNetworkID)
	}
}
