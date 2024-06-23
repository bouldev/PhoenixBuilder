package packet

import (
	"phoenixbuilder/minecraft/nbt"
	"phoenixbuilder/minecraft/protocol"
)

// AddVolumeEntity sends a volume entity's definition and metadata from server to client.
type AddVolumeEntity struct {
	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		EntityRuntimeID is the runtime ID of the volume. The runtime ID is unique for each world session, and
		entities are generally identified in packets using this runtime ID.

		For netease, the data type of this field is uint32,
		but on standard minecraft, this is uint64.
	*/
	EntityRuntimeID uint32
	// EntityRuntimeID uint64

	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the volume functions or looks.
	EntityMetadata map[string]any
	// EncodingIdentifier is the unique identifier for the volume. It must be of the form 'namespace:name', where
	// namespace cannot be 'minecraft'.
	EncodingIdentifier string
	// InstanceIdentifier is the identifier of a fog definition.
	InstanceIdentifier string
	// Bounds represent the volume's bounds. The first value is the minimum bounds, and the second value is the
	// maximum bounds.
	Bounds [2]protocol.BlockPos
	// Dimension is the dimension in which the volume exists.
	Dimension int32
	// EngineVersion is the engine version the entity is using, for example, '1.17.0'.
	EngineVersion string
}

// ID ...
func (*AddVolumeEntity) ID() uint32 {
	return IDAddVolumeEntity
}

func (pk *AddVolumeEntity) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Varuint32(&pk.EntityRuntimeID)
		// io.Uint64(&pk.EntityRuntimeID)
	}
	io.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	io.String(&pk.EncodingIdentifier)
	io.String(&pk.InstanceIdentifier)
	io.UBlockPos(&pk.Bounds[0])
	io.UBlockPos(&pk.Bounds[1])
	io.Varint32(&pk.Dimension)
	io.String(&pk.EngineVersion)
}
