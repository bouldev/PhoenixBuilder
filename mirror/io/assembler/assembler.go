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
	taskMu           sync.RWMutex
	pendingTasks     map[define.ChunkPos]*mirror.ChunkData
	visitTime        map[define.ChunkPos]time.Time
	chunkRequestChan chan []*packet.SubChunkRequest
	queueMu          sync.RWMutex
	requestQueue     map[define.ChunkPos][]*packet.SubChunkRequest
	centerChunk      *define.ChunkPos
	radius           int32
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
		requestQueue:     make(map[define.ChunkPos][]*packet.SubChunkRequest),
		taskMu:           sync.RWMutex{},
		queueMu:          sync.RWMutex{},
		centerChunk:      nil,
		radius:           10,
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
	o.taskMu.RLock()
	if _, hasK := o.pendingTasks[cp]; hasK {
		o.taskMu.RUnlock()
		return true
	}
	o.taskMu.RUnlock()
	chunk := chunk.New(o.airRID, define.Range{-64, 319})
	o.taskMu.Lock()
	o.pendingTasks[cp] = &mirror.ChunkData{
		Chunk:     chunk,
		BlockNbts: make(map[define.CubePos]map[string]interface{}),
		SyncTime:  time.Now().Unix(),
		ChunkPos:  cp,
	}
	o.taskMu.Unlock()
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
	o.taskMu.RLock()
	if chunkData, hasK := o.pendingTasks[cp]; !hasK {
		o.taskMu.RUnlock()
		//fmt.Printf("Unexpected chunk\n")
		return nil
	} else {
		o.taskMu.RUnlock()
		if pk.RequestResult != packet.SubChunkRequestResultSuccess {
			// cancel pending task
			// fmt.Println("Cancel Pending Task")
			o.taskMu.Lock()
			delete(o.pendingTasks, cp)
			o.taskMu.Unlock()
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
		chunkData.SyncTime = time.Now().Unix()
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
		o.taskMu.Lock()
		delete(o.pendingTasks, cp)
		o.visitTime[cp] = time.Now()
		o.taskMu.Unlock()
		return chunkData
	}

}

func (o *Assembler) ScheduleRequest(requests []*packet.SubChunkRequest) {
	o.chunkRequestChan <- requests
}

func (o *Assembler) CancelQueueByPublishUpdate(p *packet.NetworkChunkPublisherUpdate) {
	chunkCenterX := p.Position.X() >> 4
	chunkCenterZ := p.Position.Z() >> 4
	o.centerChunk = &define.ChunkPos{chunkCenterX, chunkCenterZ}
	o.queueMu.Lock()
	// cancelCounter := 0
	for cp, _ := range o.requestQueue {
		if (cp.X() < chunkCenterX-int32(o.radius)) || (cp.X() > chunkCenterX+int32(o.radius)) ||
			(cp.Z() < chunkCenterZ-int32(o.radius)) || (cp.Z() > chunkCenterZ+int32(o.radius)) {
			delete(o.requestQueue, cp)
			// cancelCounter += 1
		}
	}
	// if cancelCounter > 0 {
	// 	pterm.Warning.Printfln("cancel %v pending request task, %v left", cancelCounter, len(o.requestQueue))
	// }
	o.queueMu.Unlock()
}

func (o *Assembler) CreateRequestScheduler(writeFn func(pk *packet.SubChunkRequest), sendPeriod time.Duration, validCacheTime time.Duration) {
	tickerAwaked := false
	// 16 * 16 256 ~ 420 chunks
	requestSender := func() {
		// pterm.Info.Println("ticker awaked")
		t := time.NewTicker(sendPeriod / 24)
		o.queueMu.RLock()
		for cp, requests := range o.requestQueue {
			pendingTasksNum := len(o.requestQueue)
			o.queueMu.RUnlock()
			if pendingTasksNum > 512 {
				pterm.Warning.Printf("chunk request queue too long, pending %v tasks\n", pendingTasksNum*16)
			}
			o.taskMu.RLock()
			if visitTime, hasK := o.visitTime[cp]; hasK {
				if time.Since(visitTime) < validCacheTime {
					o.taskMu.RUnlock()
					o.queueMu.RLock()
					continue
				}
			}
			o.taskMu.RUnlock()
			for _, request := range requests {
				writeFn(request)
				<-t.C
			}
			o.queueMu.Lock()
			delete(o.requestQueue, cp)
			o.queueMu.Unlock()
			o.queueMu.RLock()
		}
		tickerAwaked = false
		o.queueMu.RUnlock()
	}

	go func() {
		for requests := range o.chunkRequestChan {
			firstSubChunkRequest := requests[0]
			cp := define.ChunkPos{firstSubChunkRequest.Position.X(), firstSubChunkRequest.Position.Z()}
			if o.centerChunk != nil && ((cp.X() < o.centerChunk.X()-o.radius) || (cp.X() > o.centerChunk.X()+o.radius) ||
				(cp.Z() < o.centerChunk.Z()-o.radius) || (cp.Z() > o.centerChunk.Z()+o.radius)) {
				// pterm.Warning.Printfln("Discard %v", cp)
				continue
			}
			o.queueMu.Lock()
			o.requestQueue[cp] = requests
			if !tickerAwaked {
				tickerAwaked = true
				o.queueMu.Unlock()
				requestSender()
			} else {
				o.queueMu.Unlock()
			}
		}
	}()
}
