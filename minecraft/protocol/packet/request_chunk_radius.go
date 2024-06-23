package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// RequestChunkRadius is sent by the client to the server to update the server on the chunk view radius that
// it has set in the settings. The server may respond with a ChunkRadiusUpdated packet with either the chunk
// radius requested, or a different chunk radius if the server chooses so.
type RequestChunkRadius struct {
	// ChunkRadius is the requested chunk radius. This value is always the value set in the settings of the
	// player.
	ChunkRadius int32

	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		MaxChunkRadius is the maximum chunk radius that the player wants to receive. The reason for the client sending this
		is currently unknown.

		For netease, the data type of this field is uint8,
		but on standard minecraft, this is int32.
	*/
	MaxChunkRadius uint8
	// MaxChunkRadius int32
}

// ID ...
func (*RequestChunkRadius) ID() uint32 {
	return IDRequestChunkRadius
}

func (pk *RequestChunkRadius) Marshal(io protocol.IO) {
	io.Varint32(&pk.ChunkRadius)

	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Uint8(&pk.MaxChunkRadius)
		// io.Varint32(&pk.MaxChunkRadius)
	}
}
