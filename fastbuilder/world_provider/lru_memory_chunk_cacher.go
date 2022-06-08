package world_provider

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type LRUMemoryChunkCacher struct {
	cacheLevel     int
	cacheMap       map[ChunkPosDefine]time.Time
	memoryChunks   map[ChunkPosDefine]ChunkDefine
	OverFlowHolder ChunkWriteFn
	mu             sync.Mutex
}

func (o *LRUMemoryChunkCacher) Iter(fn func(pos ChunkPosDefine, chunk ChunkDefine) (stop bool)) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for pos, chunk := range o.memoryChunks {
		if fn(pos, chunk) {
			return
		}
	}

}

func (o *LRUMemoryChunkCacher) Get(pos ChunkPosDefine) ChunkDefine {
	o.mu.Lock()
	defer o.mu.Unlock()
	if chunk, hasK := o.memoryChunks[pos]; hasK {
		return chunk
	} else {
		return nil
	}
}

type timePosPair struct {
	p ChunkPosDefine
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

func (o *LRUMemoryChunkCacher) OnNewChunk(pos ChunkPosDefine, chunk ChunkDefine) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.cacheMap[pos] = time.Now()
	o.memoryChunks[pos] = chunk
	// count++
	// fmt.Println(count," ",pos)
	if len(o.memoryChunks) > (1 << (o.cacheLevel + 1)) {
		fmt.Println("release overflowed cached chunks")
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
			o.OverFlowHolder(pair.p, o.memoryChunks[pair.p])
			delete(o.memoryChunks, pair.p)
			delete(o.cacheMap, pair.p)
		}
	}
}

// suggest level 12 (4096)
func NewLRUMemoryChunkCacher(level int) *LRUMemoryChunkCacher {
	cacher := &LRUMemoryChunkCacher{}
	cacher.cacheLevel = level
	cacher.cacheMap = make(map[ChunkPosDefine]time.Time)
	cacher.memoryChunks = make(map[ChunkPosDefine]ChunkDefine)
	cacher.OverFlowHolder = func(pos ChunkPosDefine, chunk ChunkDefine) {}
	cacher.mu = sync.Mutex{}
	return cacher
}

func init() {
	GlobalLRUMemoryChunkCacher = NewLRUMemoryChunkCacher(12)
}
