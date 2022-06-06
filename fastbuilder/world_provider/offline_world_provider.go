package world_provider

import (
	"encoding/gob"
	"fmt"
	"os"
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/dragonfly/server/world/chunk"

	"github.com/pterm/pterm"
)

type OfflineWorldProvider struct {
	chunksMap map[ChunkPosDefine]ChunkDefine
}

func NewOfflineWorldProvider(chunksMap map[ChunkPosDefine]ChunkDefine) *OfflineWorldProvider {
	return &OfflineWorldProvider {
		chunksMap :chunksMap,
	}
}

func (p *OfflineWorldProvider) LoadChunk(position world.ChunkPos) (c *chunk.Chunk, exists bool, err error) {
	cacheitem,hascacheitem:=p.chunksMap[ChunkPosDefine(position)]
	if hascacheitem {
		// delete(ChunkCache,position)
		chunk, err:=chunk.NetworkDecode(AirRuntimeId, cacheitem.RawPayload, int(cacheitem.SubChunkCount))
		if(err!=nil) {
			fileName:=fmt.Sprintf("ErrorLevelChunkSample[%v].gob",position)
			fp,err:=os.OpenFile(fileName,os.O_WRONLY|os.O_CREATE|os.O_TRUNC,0755)
			if err!=nil{
				panic(pterm.Error.Sprintf("Failed to decode chunk: %v, and even fail to save error chunk\n",err))
			}
			encoder:=gob.NewEncoder(fp)
			encoder.Encode(chunk)
			fp.Close()
			panic(pterm.Error.Sprintf("Failed to decode chunk: %v, sample saved to %v please contact developer\n",err,fileName))
			// return nil, true, err
		}
		return chunk, true, nil
	}else{
		pterm.Error.Printfln("chunk in position %v missing",position)
		return nil, true, err
	}
}

func (p *OfflineWorldProvider) Settings() world.Settings {
	return world.Settings {
		Name: "World",
	}
}

func (p *OfflineWorldProvider) SaveSettings(_ world.Settings) {
	
}

func (p *OfflineWorldProvider) SaveChunk(position world.ChunkPos, c *chunk.Chunk) error {
	return nil
}

func (p *OfflineWorldProvider) LoadEntities(position world.ChunkPos) ([]world.SaveableEntity, error) {
	// Not implemented
	return []world.SaveableEntity{}, nil
}

func (p *OfflineWorldProvider) SaveEntities(position world.ChunkPos, entities []world.SaveableEntity) error {
	return nil
}

func (p *OfflineWorldProvider) LoadBlockNBT(position world.ChunkPos) ([]map[string]interface{}, error) {
	return nil, nil
	/*r, h:=p.nbtmap[position]
	if(!h) {
		fmt.Printf("No NBT for position %v.\n",position)
		return nil, fmt.Errorf("NO NBT")
	}
	return r, nil*/
}

func (p *OfflineWorldProvider) SaveBlockNBT(position world.ChunkPos, data []map[string]interface{}) error {
	return nil
}

func (p *OfflineWorldProvider) Close() error {
	return nil
}