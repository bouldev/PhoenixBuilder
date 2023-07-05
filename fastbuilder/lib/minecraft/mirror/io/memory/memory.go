package memory

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"sync"
)

type MemoryChunkHolder struct {
	memoryChunks map[define.ChunkPos]*mirror.ChunkData
	mu           sync.Mutex
}

func (o *MemoryChunkHolder) Iter(fn func(pos define.ChunkPos, chunk *mirror.ChunkData) (stop bool)) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for pos, chunk := range o.memoryChunks {
		if fn(pos, chunk) {
			return
		}
	}
}

func (o *MemoryChunkHolder) Get(pos define.ChunkPos) (data *mirror.ChunkData) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if chunk, hasK := o.memoryChunks[pos]; hasK {
		return chunk
	} else {
		return nil
	}
}

func (o *MemoryChunkHolder) GetWithNoFallBack(pos define.ChunkPos) (data *mirror.ChunkData) {
	return o.Get(pos)
}

func (o *MemoryChunkHolder) Write(data *mirror.ChunkData) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.memoryChunks[data.ChunkPos] = data
	return nil
}

func NewMemoryChunkCacher(chunks map[define.ChunkPos]*mirror.ChunkData) *MemoryChunkHolder {
	cacher := &MemoryChunkHolder{}
	cacher.memoryChunks = chunks
	if cacher.memoryChunks == nil {
		cacher.memoryChunks = make(map[define.ChunkPos]*mirror.ChunkData)
	}
	cacher.mu = sync.Mutex{}
	return cacher
}
