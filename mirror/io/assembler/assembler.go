package assembler

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"sync"
	"time"
)

type Assembler struct {
	airRID       uint32
	pendingTasks map[define.ChunkPos]*mirror.ChunkData
	mu           sync.RWMutex
}

func NewAssembler() *Assembler {
	airRID, ok := chunk.StateToRuntimeID("minecraft:air", nil)
	if !ok {
		panic("cannot find air runtime ID")
	}
	a := &Assembler{
		airRID:       airRID,
		pendingTasks: make(map[define.ChunkPos]*mirror.ChunkData),
	}
	return a

}

func (o *Assembler) GenRequestFromLevelChunk(pk *packet.LevelChunk) (requests []*packet.SubChunkRequest) {
	requests = make([]*packet.SubChunkRequest, 0, 24)
	for i := -4; i <= 19; i++ {
		requests = append(requests, &packet.SubChunkRequest{
			Dimension: 0,
			Position:  protocol.SubChunkPos{pk.Position.X(), int32(i), pk.Position.Z()},
		})
	}
	return requests
}

func (o *Assembler) AddPendingTask(pk *packet.LevelChunk) (exist bool) {
	cp := define.ChunkPos{pk.Position.X(), pk.Position.Z()}
	if _, hasK := o.pendingTasks[cp]; hasK {
		return true
	}
	chunk := chunk.New(o.airRID, define.Range{-64, 319})
	o.mu.Lock()
	defer o.mu.Unlock()
	o.pendingTasks[cp] = &mirror.ChunkData{
		Chunk:     chunk,
		BlockNbts: make(map[define.CubePos]map[string]interface{}),
		TimeStamp: time.Now().Unix(),
		ChunkPos:  cp,
	}
	return false
}

func (o *Assembler) OnNewSubChunk(pk *packet.SubChunk) *mirror.ChunkData {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("on handle sub chunk ", r)
			return
		}
	}()
	cp := define.ChunkPos{pk.SubChunkX, pk.SubChunkZ}
	// subChunkIndex := pk.SubChunkY
	o.mu.RLock()
	if chunkData, hasK := o.pendingTasks[cp]; !hasK {
		o.mu.RUnlock()
		//fmt.Printf("Unexpected chunk\n")
		return nil
	} else {
		o.mu.RUnlock()
		subIndex, subChunk, nbts, err := chunk.NEMCSubChunkDecode(pk.Data)
		if err != nil {
			panic(err)
		}
		if subIndex != int8(pk.SubChunkY) || subIndex > 20 {
			panic(fmt.Sprintf("sub Index conflict %v %v", pk.SubChunkY, subIndex))
		}
		subs := chunkData.Chunk.Sub()
		//if subChunk.Empty() {
		//	fmt.Printf("REAL EMPTY\n")
		//}
		chunkData.Chunk.AssignSub(int(subIndex+4), subChunk)
		for _, nbt := range nbts {
			if pos, success := define.GetCubePosFromNBT(nbt); success {
				chunkData.BlockNbts[pos] = nbt
			}
		}
		// fmt.Printf("pending %v\n", len(o.pendingTasks))
		chunkData.TimeStamp = time.Now().Unix()
		//emptySubChunkCounter:=0
		for _, subChunk := range subs {
			if subChunk.Invalid() {
				//emptySubChunkCounter++
				return nil
			}
		}
		/*if emptySubChunkCounter!=0 {
			fmt.Printf("eta %d for %v\n", emptySubChunkCounter, cp)
			return nil
		}
		fmt.Printf("Finished %v\n", cp)
		*/
		o.mu.Lock()
		delete(o.pendingTasks, cp)
		o.mu.Unlock()
		return chunkData
	}

}
