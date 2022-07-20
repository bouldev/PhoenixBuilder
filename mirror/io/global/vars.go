package global

import (
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/io/lru"
)

type ChunkWriteFn func(chunk *mirror.ChunkData)

var GlobalLRUMemoryChunkCacher *lru.LRUMemoryChunkCacher
var GlobalChunkFeeder *ChunkFeeder

func init() {
	GlobalLRUMemoryChunkCacher = lru.NewLRUMemoryChunkCacher(12)
	GlobalChunkFeeder = NewChunkFeeder()
}
