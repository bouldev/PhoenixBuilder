package assembler

import (
	"fmt"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/chunk"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
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
	allowCache       bool
	validCacheTime   time.Duration
	sendPeriod       time.Duration
}

func NewAssembler(sendPeriod time.Duration, validCacheTime time.Duration) *Assembler {
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
	a.AdjustSendPeriod(sendPeriod)
	a.AdjustValidCacheTime(validCacheTime)
	return a

}

const (
	REQUEST_AGGRESSIVE = time.Millisecond * 40
	REQUEST_NORMAL     = time.Millisecond * 500
	REQUEST_LAZY       = time.Millisecond * 10000
)

func (o *Assembler) AdjustValidCacheTime(d time.Duration) {
	o.validCacheTime = d
}

func (o *Assembler) AdjustSendPeriod(d time.Duration) {
	o.sendPeriod = d / 24
}

func (o *Assembler) GenRequestFromLevelChunk(pk *packet.LevelChunk) (requests []*packet.SubChunkRequest) {
	requests = make([]*packet.SubChunkRequest, 0, 1)
	offsets := make([][3]int8, 24)
	for i := -4; i <= 19; i++ {
		offsets[i+4] = [3]int8{0, int8(i), 0}
	}
	return []*packet.SubChunkRequest{
		&packet.SubChunkRequest{
			Dimension: 0,
			Position:  protocol.SubChunkPos{pk.Position[0], 0, pk.Position[1]},
			Offsets:   offsets,
		},
	}
}

func (o *Assembler) NoCache() {
	o.queueMu.Lock()
	o.taskMu.Lock()
	o.allowCache = false
	o.visitTime = make(map[define.ChunkPos]time.Time)
	o.requestQueue = make(map[define.ChunkPos][]*packet.SubChunkRequest)
	o.pendingTasks = make(map[define.ChunkPos]*mirror.ChunkData)
	for len(o.chunkRequestChan) > 0 {
		<-o.chunkRequestChan
	}
	o.taskMu.Unlock()
	o.queueMu.Unlock()
}

func (o *Assembler) AllowCache() {
	o.allowCache = true
}

func (o *Assembler) AddPendingTask(pk *packet.LevelChunk) (exist bool) {
	cp := define.ChunkPos{pk.Position.X(), pk.Position.Z()}
	o.taskMu.RLock()
	if _, hasK := o.pendingTasks[cp]; hasK {
		o.taskMu.RUnlock()
		return true
	}
	o.taskMu.RUnlock()
	chunk := chunk.New(o.airRID, define.WorldRange)
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
			fmt.Println("on handle sub chunk ", r)
			return
		}
	}()
	cp := define.ChunkPos{pk.Position[0], pk.Position[2]}
	// subChunkIndex := pk.Position[1]
	o.taskMu.RLock()
	if chunkData, hasK := o.pendingTasks[cp]; !hasK {
		o.taskMu.RUnlock()
		//fmt.Printf("Unexpected chunk %#v %#v\n", cp, o.pendingTasks)
		return nil
	} else {
		o.taskMu.RUnlock()
		for _, entry := range pk.SubChunkEntries {
			chunkPosX := int(pk.Position.X())
			chunkPosY := int(int8(pk.Position[1]) + entry.Offset[1] + 4)
			chunkPosZ := int(pk.Position.Z())
			worldPos := define.CubePos{chunkPosX * 16, chunkPosY * 16, chunkPosZ * 16}
			if entry.Result != protocol.SubChunkResultSuccess {
				if entry.Result == protocol.SubChunkResultSuccessAllAir {
					allAirSubChunk := chunk.NewSubChunk(o.airRID)
					allAirSubChunk.Validate()
					chunkData.Chunk.AssignSub(chunkPosY, allAirSubChunk)
					continue
				}
				fmt.Printf("SubChunkResult Err, pos: %v result: %v, bot in main world?\n", worldPos, entry.Result)
				o.taskMu.Lock()
				delete(o.pendingTasks, cp)
				o.taskMu.Unlock()
				return nil
			}
			subIndex, subChunk, nbts, err := chunk.NEMCSubChunkDecode(entry.RawPayload)
			if err != nil {
				fmt.Printf("%#v (world pos: %v)", entry, worldPos)
				panic(err)
			}
			if subIndex != int8(pk.Position[1])+entry.Offset[1] || subIndex > 20 {
				panic(fmt.Sprintf("sub Index conflict %v %v (world pos: %v)", pk.Position[1], subIndex, worldPos))
			}
			//subs := chunkData.Chunk.Sub()
			chunkData.Chunk.AssignSub(int(subIndex+4), subChunk)
			for _, nbt := range nbts {
				if pos, success := define.GetCubePosFromNBT(nbt); success {
					chunkData.BlockNbts[pos] = nbt
				}
			}
			chunkData.SyncTime = time.Now().Unix()
		}
		// fmt.Printf("pending %v\n", len(o.pendingTasks))
		chunkData.SyncTime = time.Now().Unix()
		emptySubChunkCounter := 0
		subs := chunkData.Chunk.Sub()
		for _, subChunk := range subs {
			if subChunk.Invalid() {
				emptySubChunkCounter++
				//return nil
			}
		}
		if emptySubChunkCounter != 0 {
			fmt.Printf("Error combining chunk: eta %d for %v\n", emptySubChunkCounter, cp)
			return nil
		}
		//fmt.Printf("Finished %v\n", cp)

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

func (o *Assembler) CreateRequestScheduler(writeFn func(pk *packet.SubChunkRequest)) {
	tickerAwaked := false
	// 16 * 16 256 ~ 420 chunks
	requestSender := func() {
		// pterm.Info.Println("ticker awaked")
		lastCheckPointTime := time.Now()
		o.queueMu.RLock()
		for cp, requests := range o.requestQueue {
			pendingTasksNum := len(o.requestQueue)
			o.queueMu.RUnlock()
			if pendingTasksNum > 512 {
				pterm.Warning.Printf("chunk request queue too long, pending %v tasks\n", pendingTasksNum*16)
			}
			if o.allowCache {
				o.taskMu.RLock()
				if visitTime, hasK := o.visitTime[cp]; hasK {
					if time.Since(visitTime) < o.validCacheTime {
						o.taskMu.RUnlock()
						o.queueMu.RLock()
						continue
					}
				}
				o.taskMu.RUnlock()
			}
			for _, request := range requests {
				writeFn(request)
				time.Sleep(time.Until(lastCheckPointTime.Add(o.sendPeriod)))
				lastCheckPointTime = time.Now()
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
