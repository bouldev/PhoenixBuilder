package io

import (
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"time"
)

func NEMCPacketToChunkData(p *packet.LevelChunk) (cd *mirror.ChunkData) {
	c, nbts, err := chunk.NEMCNetworkDecode(p.RawPayload[:], int(p.SubChunkCount))
	if err != nil {
		return nil
	}
	posedNbt := make(map[define.CubePos]map[string]interface{})
	for _, nbt := range nbts {
		if pos, success := define.GetCubePosFromNBT(nbt); success {
			posedNbt[pos] = nbt
		}
	}
	// define.GetCubePosFromNBT()
	cd = &mirror.ChunkData{
		Chunk: c, BlockNbts: posedNbt,
		ChunkPos:  define.ChunkPos{p.ChunkX, p.ChunkZ},
		TimeStamp: time.Now().Unix(),
	}
	return cd
}
