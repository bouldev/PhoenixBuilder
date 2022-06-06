package fetcher

import "phoenixbuilder/minecraft/protocol/packet"


type ChunkPosDefine [2]int

type ExportHopPos struct {
	Pos           ChunkPosDefine
	LinkedChunk []*ExportedChunkPos
}

type ExportedChunkPos struct {
	Pos          ChunkPosDefine
	MasterHop  *ExportHopPos
	CachedMark bool
}

type ExportedChunksMap map[ChunkPosDefine]*ExportedChunkPos

type ChunkDefine *packet.LevelChunk

type ChunkDefineWithPos struct{
	Chunk ChunkDefine
	Pos ChunkPosDefine
}

type TeleportFn func (x,z int)

type ChunkFeedChan chan *ChunkDefineWithPos