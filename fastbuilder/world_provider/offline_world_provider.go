package world_provider

import (
	"fmt"
	
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/dragonfly/server/world/chunk"
	"phoenixbuilder/dragonfly/server/block/cube"
	"github.com/google/uuid"

	"github.com/pterm/pterm"
)

type OfflineWorldProvider struct {
	chunksMap map[ChunkPosDefine]ChunkDefine
}

func NewOfflineWorldProvider(chunksMap map[ChunkPosDefine]ChunkDefine) *OfflineWorldProvider {
	return &OfflineWorldProvider {
		chunksMap: chunksMap,
	}
}

func (p *OfflineWorldProvider) LoadChunk(position world.ChunkPos, dim world.Dimension) (c *chunk.Chunk, exists bool, err error) {
	cacheitem,hascacheitem:=p.chunksMap[ChunkPosDefine(position)]
	if hascacheitem {
		// delete(ChunkCache,position)
		chunk, err:=chunk.NetworkDecode(AirRuntimeId, cacheitem.RawPayload, int(cacheitem.SubChunkCount), cube.Range{-64, 319})
		if(err!=nil) {
			panic(fmt.Errorf("Failed to decode chunk: %v", err))
			/*fileName:=fmt.Sprintf("ErrorLevelChunkSample[%v].gob",position)
			fp,err:=os.OpenFile(fileName,os.O_WRONLY|os.O_CREATE|os.O_TRUNC,0755)
			if err!=nil{
				panic(pterm.Error.Sprintf("Failed to decode chunk: %v, and even fail to save error chunk\n",err))
			}
			encoder:=gob.NewEncoder(fp)
			encoder.Encode(chunk)
			fp.Close()
			panic(pterm.Error.Sprintf("Failed to decode chunk: %v, sample saved to %v please contact developer\n",err,fileName))
			*/// return nil, true, err
		}
		return chunk, true, nil
	}else{
		pterm.Error.Printfln("chunk in position %v missing",position)
		return nil, true, err
	}
}

func (p *OfflineWorldProvider) Settings() *world.Settings {
	return &world.Settings {
		Name: "World",
	}
}

func (p *OfflineWorldProvider) SaveSettings(_ *world.Settings) {
	
}

func (p *OfflineWorldProvider) SaveChunk(position world.ChunkPos, c *chunk.Chunk, dim world.Dimension) error {
	return nil
}

func (p *OfflineWorldProvider) LoadEntities(position world.ChunkPos, dim world.Dimension) ([]world.SaveableEntity, error) {
	// Not implemented
	return []world.SaveableEntity{}, nil
}

func (p *OfflineWorldProvider) SaveEntities(position world.ChunkPos, entities []world.SaveableEntity, dim world.Dimension) error {
	return nil
}

func (p *OfflineWorldProvider) LoadBlockNBT(position world.ChunkPos, dim world.Dimension) ([]map[string]any, error) {
	return nil, nil
}

func (p *OfflineWorldProvider) SaveBlockNBT(position world.ChunkPos, data []map[string]interface{}, dim world.Dimension) error {
	return nil
}

func (p *OfflineWorldProvider) Close() error {
	return nil
}

func (p *OfflineWorldProvider) LoadPlayerSpawnPosition(uuid uuid.UUID) (pos cube.Pos, exists bool, err error) {
	return cube.Pos{}, false, nil
}

func (p *OfflineWorldProvider) SavePlayerSpawnPosition(uuid uuid.UUID, pos cube.Pos) error {
	return nil
}