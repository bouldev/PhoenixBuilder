package io

import (
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"time"
)

func NEMCPacketToChunkData(p *packet.LevelChunk) (cd *mirror.ChunkData) {
	c, nbt, err := chunk.NEMCNetworkDecode(p.RawPayload[:], int(p.SubChunkCount))
	if err != nil {
		return nil
	}
	cd = &mirror.ChunkData{
		Chunk: c, BlockNbts: nbt,
		ChunkPos:  define.ChunkPos{p.Position[0], p.Position[0]},
		TimeStamp: time.Now().Unix(),
	}
	return cd
}
