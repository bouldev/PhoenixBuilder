package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// AvailableActorIdentifiers is sent by the server at the start of the game to let the client know all
// entities that are available on the server.
type AvailableActorIdentifiers struct {
	// SerialisedEntityIdentifiers is a network NBT serialised compound of all entity identifiers that are
	// available in the server.
	SerialisedEntityIdentifiers []byte
}

// ID ...
func (*AvailableActorIdentifiers) ID() uint32 {
	return IDAvailableActorIdentifiers
}

func (pk *AvailableActorIdentifiers) Marshal(io protocol.IO) {
	// PhoenixBuilder specific comments.
	// Author: Liliya233
	//
	// Netease: io.NBT() can be used here
	io.Bytes(&pk.SerialisedEntityIdentifiers)
}
