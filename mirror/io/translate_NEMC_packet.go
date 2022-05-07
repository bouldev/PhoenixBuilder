package io

import (
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"time"
)

func NEMCPacketToChunkData(p *packet.LevelChunk) (cd *mirror.ChunkData) {
	c, nbts, err := chunk.NEMCNetworkDecode(p.RawPayload, int(p.SubChunkCount))
	if err != nil {
		return nil
	}
	cd = &mirror.ChunkData{
		Chunk: c, BlockNbts: nbts,
		ChunkPos:  define.ChunkPos{p.ChunkX, p.ChunkZ},
		TimeStamp: time.Now().Unix(),
	}
	return cd
}
