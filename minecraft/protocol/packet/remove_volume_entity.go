package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// RemoveVolumeEntity indicates a volume entity to be removed from server to client.
type RemoveVolumeEntity struct {
	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		For netease, the data type of this field is uint32,
		but on standard minecraft, this is uint64.
	*/
	EntityRuntimeID uint32
	// EntityRuntimeID uint64

	// Dimension ...
	Dimension int32
}

// ID ...
func (*RemoveVolumeEntity) ID() uint32 {
	return IDRemoveVolumeEntity
}

func (pk *RemoveVolumeEntity) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Varuint32(&pk.EntityRuntimeID)
		// io.Uint64(&pk.EntityRuntimeID)
	}
	io.Varint32(&pk.Dimension)
}
