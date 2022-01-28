package world_provider

import (
	"fmt"
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/dragonfly/server/world/chunk"
	"phoenixbuilder/dragonfly/server/block/cube"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/minecraft"
	"github.com/google/uuid"
)

var ChunkInput chan *packet.LevelChunk = nil
var ChunkCache map[world.ChunkPos]*packet.LevelChunk = nil


type OnlineWorldProvider struct {
	connection *minecraft.Conn
	//nbtmap map[world.ChunkPos][]map[string]interface{}
}

func NewOnlineWorldProvider(conn *minecraft.Conn) *OnlineWorldProvider {
	return &OnlineWorldProvider {
		connection: conn,
		//nbtmap: make(map[world.ChunkPos][]map[string]interface{}),
	}
}

func (p *OnlineWorldProvider) Settings() world.Settings {
	return world.Settings {
		Name: "World",
	}
}

func (p *OnlineWorldProvider) SaveSettings(_ world.Settings) {
	
}

func quickCache(pkt *packet.LevelChunk) {
	ChunkCache[world.ChunkPos[pkt.ChunkX,pkt.ChunkZ]]=pkt
}

func (p *OnlineWorldProvider) LoadChunk(position world.ChunkPos) (c *chunk.Chunk, exists bool, err error) {
	if(ChunkCache==nil) {
		panic("LoadChunk() before creating a world")
	}
	cacheitem,hascacheitem:=ChunkCache[position]
	if hascacheitem {
		chunk, err:=chunk.NetworkDecode(134, inp.RawPayload, int(inp.SubChunkCount))
		if(err!=nil) {
			fmt.Printf("Failed to decode chunk: %v\n",err)
			return nil, true, err
		}
		return chunk, true, nil
	}
	if(ChunkInput!=nil) {
		panic("Multithreading on OnlineWorldProvider's LoadChunk function isn't allowed")
	}
	ChunkInput=make(chan *packet.LevelChunk,32)
	u_d, _ := uuid.NewUUID()
	tr:=make(chan bool)
	command.PlayerTeleportSubscribers.Put(tr)
	err=command.SendWSCommand("tp 0 127 0",u_d,p.connection)
	<-tr
	close(tr)
	u_d, _ = uuid.NewUUID()
	err=command.SendWSCommand(fmt.Sprintf("tp %d 127 %d",100000-position[0],100000-position[1]),u_d,p.connection)
	if(err!=nil) {
		panic(fmt.Errorf("Connection closed: %+v",err))
	}
	//<-tr
	quickCache(<-ChunkInput)
	u_d, _ = uuid.NewUUID()
	//command.PlayerTeleportSubscribers.Put(tr)
	err=command.SendWSCommand(fmt.Sprintf("tp %d 127 %d",position[0]*16,position[1]*16),u_d,p.connection)
	if(err!=nil) {
		panic(fmt.Errorf("[2]Connection closed: %+v",err))
	}
	//<-tr
	//close(tr)
	//p.connection.WritePacket(&packet.RequestChunkRadius {
	//	ChunkRadius: 16,
	//})
	for {
		inp,hasqi:=ChunkCache[position]
		if !hasqi {
			inp=<-ChunkInput
			quickCache(inp)
			fmt.Printf("%d, %d | %v\n",inp.ChunkX,inp.ChunkZ,position)
			if(inp.ChunkX!=position[0]||inp.ChunkZ!=position[1]) {
				continue
			}
		}
		// Hit
		close(ChunkInput)
		ChunkInput=nil
		chunk, err:=chunk.NetworkDecode(134, inp.RawPayload, int(inp.SubChunkCount))
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
		//counter:=0
		return chunk, true, nil
		/*
		nbtout:=make([]map[string]interface{},len(chunk.BlockNBT()))
		for _, item := range chunk.BlockNBT() {
			nbtout[counter]=item
			counter++
		}
		p.nbtmap[position]=nbtout
		return chunk, true, nil*/
	}
}

func (p *OnlineWorldProvider) SaveChunk(position world.ChunkPos, c *chunk.Chunk) error {
	return nil
}

func (p *OnlineWorldProvider) LoadEntities(position world.ChunkPos) ([]world.SaveableEntity, error) {
	// Not implemented
	return []world.SaveableEntity{}, nil
}

func (p *OnlineWorldProvider) SaveEntities(position world.ChunkPos, entities []world.SaveableEntity) error {
	return nil
}

func (p *OnlineWorldProvider) LoadBlockNBT(position world.ChunkPos) ([]map[string]interface{}, error) {
	return nil, nil
	r, h:=p.nbtmap[position]
	if(!h) {
		fmt.Printf("No NBT for position %v.\n",position)
		return nil, fmt.Errorf("NO NBT")
	}
	return r, nil
}

func (p *OnlineWorldProvider) SaveBlockNBT(position world.ChunkPos, data []map[string]interface{}) error {
	return nil
}

func (p *OnlineWorldProvider) Close() error {
	return nil
}