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

	"github.com/pterm/pterm"
)

type Assembler struct {
	airRID           uint32
	pendingTasks     map[define.ChunkPos]*mirror.ChunkData
	mu               sync.RWMutex
	chunkRequestChan chan []*packet.SubChunkRequest
	visitTime        map[define.ChunkPos]time.Time
}

func NewAssembler() *Assembler {
	airRID, ok := chunk.StateToRuntimeID("minecraft:air", nil)
	if !ok {
		panic("cannot find air runtime ID")
	}
	a := &Assembler{
		airRID:           airRID,
		pendingTasks:     make(map[define.ChunkPos]*mirror.ChunkData),
		chunkRequestChan: make(chan []*packet.SubChunkRequest, 10240),
		visitTime:        make(map[define.ChunkPos]time.Time),
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
			fmt.Println("on handle sub chunk ", r, pk)
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
		if pk.RequestResult != packet.SubChunkRequestResultSuccess {
			// cancel pending task
			fmt.Println("Cancel Pending Task")
			o.mu.Lock()
			delete(o.pendingTasks, cp)
			o.mu.Unlock()
			return nil
		}
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
		o.visitTime[cp] = time.Now()
		o.mu.Unlock()
		return chunkData
	}

}

func (o *Assembler) ScheduleRequest(requests []*packet.SubChunkRequest) {
	o.chunkRequestChan <- requests

}

func (o *Assembler) CreateRequestScheduler(writeFn func(pk *packet.SubChunkRequest), sendPeriod time.Duration, validCacheTime time.Duration) {
	go func() {
		// t := time.NewTicker(time.Second / 40)
		t := time.NewTicker(sendPeriod)
		// visitTime := make(map[protocol.SubChunkPos]time.Time)
		for requests := range o.chunkRequestChan {
			// fmt.Println("request")
			if len(o.chunkRequestChan) > 512 {
				pterm.Warning.Println("chunk request too busy")
			}
			first_subchunk_request := requests[0]
			if visitTime, hasK := o.visitTime[define.ChunkPos{first_subchunk_request.Position.X(), first_subchunk_request.Position.Z()}]; hasK {
				o.mu.RLock()
				if time.Since(visitTime) < validCacheTime {
					o.mu.RUnlock()
					continue
				}
				o.mu.RUnlock()
			}
			// visitTime[request0.Position] = time.Now()
			for _, request := range requests {
				writeFn(request)
				// o.o.adaptor.Write(&packet.SubChunkRequest{
				// 	0,
				// 	protocol.SubChunkPos{1249, 4, -1249}, nil,
				// })
			}
			<-t.C
		}
	}()
}
