package world_provider

import (
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/minecraft/protocol/packet"
)


type ChunkDefine *packet.LevelChunk
type ChunkPosDefine world.ChunkPos
type ChunkWriteFn func(pos ChunkPosDefine,chunk ChunkDefine)
var GlobalLRUMemoryChunkCacher *LRUMemoryChunkCacher
var GlobalChunkFeeder *ChunkFeeder