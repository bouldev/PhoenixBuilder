package session

import (
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/minecraft/protocol/packet"
)

// SubChunkRequestHandler handles sub-chunk requests from the client. The server will respond with a packet containing
// the requested sub-chunks.
type SubChunkRequestHandler struct{}

// Handle ...
func (*SubChunkRequestHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.SubChunkRequest)
	s.ViewSubChunks(world.SubChunkPos(pk.Position), pk.Offsets)
	return nil
}
