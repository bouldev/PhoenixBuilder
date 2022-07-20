package world_provider

import (
	"fmt"
	"time"
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/dragonfly/server/world/chunk"
	"phoenixbuilder/dragonfly/server/block/cube"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/minecraft"
	"github.com/google/uuid"
	"runtime"
)

var ChunkInput chan *packet.LevelChunk = nil
var ChunkCache map[world.ChunkPos]*packet.LevelChunk = nil
var firstLoaded bool = false


type OnlineWorldProvider struct {
	env *environment.PBEnvironment
	connection *minecraft.Conn
	//nbtmap map[world.ChunkPos][]map[string]interface{}
}

func NewOnlineWorldProvider(env *environment.PBEnvironment) *OnlineWorldProvider {
	conn:=env.Connection.(*minecraft.Conn)
	return &OnlineWorldProvider {
		connection: conn,
		env: env,
		//nbtmap: make(map[world.ChunkPos][]map[string]interface{}),
	}
}

func (p *OnlineWorldProvider) Settings() *world.Settings {
	return &world.Settings {
		Name: "World",
	}
}

func (p *OnlineWorldProvider) SaveSettings(_ *world.Settings) {
	
}

func DoCache(pkt *packet.LevelChunk) {
	if ChunkCache != nil {
		quickCache(pkt)
	}
}

func quickCache(pkt *packet.LevelChunk) {
	ChunkCache[world.ChunkPos{pkt.Position[0],pkt.Position[1]}]=pkt
}

func wander(env *environment.PBEnvironment, position world.ChunkPos) {
	u_d, _ := uuid.NewUUID()
	cmdsender:=env.CommandSender
	err:=cmdsender.SendWSCommand(fmt.Sprintf("tp %d 127 %d",position[0]*16+100000,1000000-position[1]*16+100000),u_d)
	if(err!=nil) {
		panic(fmt.Errorf("Connection closed: %+v",err))
	}
	select {
	case <-ChunkInput :
		//quickCache(inp)
	case <-time.After(2*time.Second):
	
	}
	u_d, _ = uuid.NewUUID()
	err=cmdsender.SendWSCommand(fmt.Sprintf("tp %d 127 %d",position[0]*16,position[1]*16),u_d)
	if(err!=nil) {
		panic(fmt.Errorf("[2]Connection closed: %+v",err))
	}
}

func (p *OnlineWorldProvider) LoadChunk(position world.ChunkPos, dim world.Dimension) (c *chunk.Chunk, exists bool, err error) {
	if(ChunkCache==nil) {
		panic("LoadChunk() before creating a world")
	}
	cacheitem,hascacheitem:=ChunkCache[position]
	if hascacheitem {
		delete(ChunkCache,position)
		chunk, err:=chunk.NetworkDecode(AirRuntimeId, cacheitem.RawPayload, int(cacheitem.SubChunkCount), cube.Range{-64, 319})
		if(err!=nil) {
			fmt.Printf("Failed to decode chunk: %v\n",err)
			return nil, true, err
		}
		return chunk, true, nil
	}
	if(ChunkInput!=nil) {
		panic("Multithreading on OnlineWorldProvider's LoadChunk function isn't allowed")
	}
	u_d, _ := uuid.NewUUID()
	ChunkInput=make(chan *packet.LevelChunk,32)
	err=p.env.CommandSender.SendWSCommand(fmt.Sprintf("tp %d 127 %d",position[0]*16,position[1]*16),u_d)
	if(err!=nil) {
		panic(fmt.Errorf("[2]Connection closed: %+v",err))
	}
	for {
		inp,hasqi:=ChunkCache[position]
		if !hasqi {
			select {
			case inp=<-ChunkInput:
				quickCache(inp)
				fmt.Printf("Waiting for chunk: current: %d, %d | expected: %v\n",inp.Position[0],inp.Position[1],position)
				if(inp.Position[0]!=position[0]||inp.Position[1]!=position[1]) {
					continue
				}
			case <-time.After(2*time.Second):
				runtime.GC()
				fmt.Printf("Expected chunk %v didn't arrive, wandering around\n", position)
				wander(p.env, position)
				continue
			}
		}else{
			delete(ChunkCache,position)
		}
		// Hit
		close(ChunkInput)
		ChunkInput=nil
		chunk, err:=chunk.NetworkDecode(AirRuntimeId, inp.RawPayload, int(inp.SubChunkCount), cube.Range{-64, 319})
		if(err!=nil) {
			fmt.Printf("Failed to decode chunk: %v\n",err)
			return nil, true, err
		}
		/*blockentities:=bytes.NewReader(inp.RawPayload[len(inp.RawPayload)-ato:])
		blockentities.ReadByte()
		dec:=nbt.NewDecoderWithEncoding(blockentities, nbt.NetworkLittleEndian)
		nbtout:=make([]map[string]interface{},0)
		for {
			out:=make(map[string]interface{})
			err=dec.Decode(&out)
			if(err!=nil) {
				break
			}
			nbtout=append(nbtout,out)
		}
		p.nbtmap[position]=nbtout*/
		return chunk, true, nil
	}
}

func (p *OnlineWorldProvider) SaveChunk(position world.ChunkPos, c *chunk.Chunk, dim world.Dimension) error {
	return nil
}

func (p *OnlineWorldProvider) LoadEntities(position world.ChunkPos, dim world.Dimension) ([]world.SaveableEntity, error) {
	// Not implemented
	return []world.SaveableEntity{}, nil
}

func (p *OnlineWorldProvider) SaveEntities(position world.ChunkPos, entities []world.SaveableEntity, dim world.Dimension) error {
	return nil
}

func (p *OnlineWorldProvider) LoadBlockNBT(position world.ChunkPos, dim world.Dimension) ([]map[string]any, error) {
	return []map[string]any{}, nil
}

func (p *OnlineWorldProvider) SaveBlockNBT(position world.ChunkPos, data []map[string]interface{}, dim world.Dimension) error {
	return nil
}

func (p *OnlineWorldProvider) Close() error {
	return nil
}

func (p *OnlineWorldProvider) LoadPlayerSpawnPosition(uuid uuid.UUID) (pos cube.Pos, exists bool, err error) {
	return cube.Pos{}, false, nil
}

func (p *OnlineWorldProvider) SavePlayerSpawnPosition(uuid uuid.UUID, pos cube.Pos) error {
	return nil
}