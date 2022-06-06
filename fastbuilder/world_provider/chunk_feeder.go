package world_provider

import (
	"sync"

	"github.com/google/uuid"
)


type ChunkFeeder struct{
	readers map[uuid.UUID]ChunkWriteFn
	mu sync.Mutex
}

func NewChunkFeeder() *ChunkFeeder{
	return &ChunkFeeder{
		readers: make(map[uuid.UUID]ChunkWriteFn),
		mu: sync.Mutex{},
	}
}

func (o *ChunkFeeder) OnNewChunk(pos ChunkPosDefine,chunk ChunkDefine){
	go func ()  {
		o.mu.Lock()
		defer o.mu.Unlock()
		for _,reader:=range o.readers{
			reader(pos,chunk)
		}
	}()
}

func (o *ChunkFeeder)RegNewReader(fn ChunkWriteFn) (unRegFn func()){
	for {
		uid,_:=uuid.NewUUID()
		if _,hasK:=o.readers[uid];hasK{
			continue
		}else{
			o.readers[uid]=fn
			return func() {
				o.mu.Lock()
				defer o.mu.Unlock()
				delete(o.readers, uid)
			}
		}
	}
}

func init(){
	GlobalChunkFeeder=NewChunkFeeder()
}