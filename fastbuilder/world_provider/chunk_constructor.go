package world_provider

import (
	"fmt"
	"sync"
	"bytes"
	"phoenixbuilder/dragonfly/server/world/chunk"
	"phoenixbuilder/dragonfly/server/block/cube"
	//"phoenixbuilder/minecraft/nbt"
)

type ChunkConstructionTask struct {
	UnarrivedSubChunks int16 // 319+64=383
	Result *chunk.Chunk
	Chunk ChunkDefine
}

type ChunkConstructor struct {
	tasks map[ChunkPosDefine]*ChunkConstructionTask
	lock sync.Mutex
}

func (c *ChunkConstructor) BeginConstruction(chunkd ChunkDefine) {
	c.lock.Lock()
	c.tasks[ChunkPosDefine{chunkd.Position[0],chunkd.Position[1]}]=&ChunkConstructionTask {
		UnarrivedSubChunks: 24,
		Chunk: chunkd,
		Result: chunk.New(AirRuntimeId, cube.Range{-64,319}),
	}
	c.lock.Unlock()
}

func (c *ChunkConstructor) SubChunkArrived(subchunk []byte, x int32, y int32, z int32) ChunkDefine {
	c.lock.Lock()
	task, found:=c.tasks[ChunkPosDefine{x,z}]
	if !found {
		c.lock.Unlock()
		// This shouldn't happen, I think.
		return nil
	}
	task.UnarrivedSubChunks--
	err:=task.Result.SubChunkArrived(int16(y), subchunk)
	if err != nil {
		fmt.Printf("ERROR: SubChunk [%d, %d, %d] - failed to decode: %v, assuming it as an air-only subchunk.\n",x,y,z,err)
	}
	if task.UnarrivedSubChunks==0 {
		r:=task.Result
		ret:=task.Chunk
		delete(c.tasks, ChunkPosDefine{x,z})
		c.lock.Unlock()
		res:=chunk.Encode(r, chunk.NetworkEncoding)
		cb:=bytes.NewBuffer(nil)
		cb.Write(res.Biomes)
		cb.WriteByte(0)
		/*enc := nbt.NewEncoderWithEncoding(cb, nbt.NetworkLittleEndian)
		for bp, b := range blockEntities {
			if n, ok := b.(world.NBTer); ok {
				d := n.EncodeNBT()
				d["x"], d["y"], d["z"] = int32(bp[0]), int32(bp[1]), int32(bp[2])
				_ = enc.Encode(d)
			}
		}
		// TODO: Add NBT
		*/
		ret.RawPayload=append([]byte(nil), cb.Bytes()...)
		return ret
	}
	c.lock.Unlock()
	return nil
}