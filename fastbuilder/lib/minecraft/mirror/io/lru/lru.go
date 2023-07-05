package lru

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"sort"
	"sync"
	"time"
)

type LRUMemoryChunkCacher struct {
	eagerWrite       bool
	cacheLevel       int
	cacheMap         map[define.ChunkPos]time.Time
	memoryChunks     map[define.ChunkPos]*mirror.ChunkData
	OverFlowHolder   mirror.ChunkWriter
	FallBackProvider mirror.ChunkReader
	mu               sync.Mutex
}

func (o *LRUMemoryChunkCacher) Iter(fn func(pos define.ChunkPos, chunk *mirror.ChunkData) (stop bool)) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for pos, chunk := range o.memoryChunks {
		if fn(pos, chunk) {
			return
		}
	}
}

func (o *LRUMemoryChunkCacher) Get(pos define.ChunkPos) (data *mirror.ChunkData) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if chunk, hasK := o.memoryChunks[pos]; hasK {
		return chunk
	} else if o.FallBackProvider == nil {
		return nil
	} else {
		cd := o.FallBackProvider.Get(pos)
		if cd != nil {
			o.memoryChunks[pos] = cd
			o.cacheMap[pos] = time.Now()
			o.checkCacheSizeAndHandleFallBackNoLock()
		}
		return cd
	}
}

func (o *LRUMemoryChunkCacher) GetWithNoFallBack(pos define.ChunkPos) (data *mirror.ChunkData) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if chunk, hasK := o.memoryChunks[pos]; hasK {
		return chunk
	} else {
		return nil
	}
}

type timePosPair struct {
	p define.ChunkPos
	t time.Time
}
type SortableTimes []*timePosPair

func (s SortableTimes) Len() int           { return len(s) }
func (s SortableTimes) Less(i, j int) bool { return s[i].t.Before(s[j].t) }
func (s SortableTimes) Swap(i, j int) {
	t := s[i]
	s[i] = s[j]
	s[j] = t
}

// var count int
func (o *LRUMemoryChunkCacher) AdjustCacheLevel(level int) {
	o.cacheLevel = level
}

func (o *LRUMemoryChunkCacher) checkCacheSizeAndHandleFallBackNoLock() {
	if len(o.memoryChunks) > (1 << (o.cacheLevel + 1)) {
		// fmt.Println("release overflowed cached chunks")
		cacheList := make(SortableTimes, 0)
		for pos, t := range o.cacheMap {
			cacheList = append(cacheList, &timePosPair{
				p: pos,
				t: t,
			})
		}
		sort.Sort(cacheList)
		for i := 0; i < 1<<o.cacheLevel; i++ {
			pair := cacheList[i]
			if !o.eagerWrite {
				if o.OverFlowHolder != nil {
					o.OverFlowHolder.Write(o.memoryChunks[pair.p])
				}
			}
			delete(o.memoryChunks, pair.p)
			delete(o.cacheMap, pair.p)
		}
	}
}

func (o *LRUMemoryChunkCacher) Write(data *mirror.ChunkData) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.cacheMap[data.ChunkPos] = time.Now()
	o.memoryChunks[data.ChunkPos] = data
	if o.eagerWrite {
		if o.OverFlowHolder != nil {
			o.OverFlowHolder.Write(data)
		}
	}
	o.checkCacheSizeAndHandleFallBackNoLock()
	// count++
	// fmt.Println(count," ",pos)
	return nil
}

func (o *LRUMemoryChunkCacher) Close() {
	if o.eagerWrite {
		return
	}
	// pterm.Info.Println(o.memoryChunks)
	for _, chunk := range o.memoryChunks {
		if o.OverFlowHolder != nil {
			o.OverFlowHolder.Write(chunk)
		}
	}
}

// suggest level 12 (4096)
func NewLRUMemoryChunkCacher(level int, eagerWrite bool) *LRUMemoryChunkCacher {
	cacher := &LRUMemoryChunkCacher{}
	cacher.eagerWrite = eagerWrite
	cacher.cacheLevel = level
	cacher.cacheMap = make(map[define.ChunkPos]time.Time)
	cacher.memoryChunks = make(map[define.ChunkPos]*mirror.ChunkData)
	cacher.OverFlowHolder = nil
	cacher.mu = sync.Mutex{}
	return cacher
}
