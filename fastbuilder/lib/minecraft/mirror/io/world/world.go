package world

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/chunk"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
)

type World struct {
	provider  mirror.ChunkProvider
	lastPos   define.ChunkPos
	lastChunk *mirror.ChunkData
}

func (w *World) chunk(pos define.ChunkPos) *mirror.ChunkData {
	c := w.provider.Get(pos)
	return c
}

// TODO Check if WoodAxe is affected by 0 -> -64
func (w *World) Block(pos define.CubePos) (rtid uint32, found bool) {
	if w == nil || pos.OutOfYBounds() {
		// Fast way out.
		return chunk.AirRID, false
	}
	chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	var c *mirror.ChunkData
	if w.lastChunk != nil && w.lastPos == chunkPos {
		c = w.lastChunk
	} else {
		c = w.chunk(chunkPos)
		w.lastChunk = c
		w.lastPos = chunkPos
	}
	if c == nil {
		return chunk.AirRID, false
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	rtid = c.Chunk.Block(x, y, z, 0)
	return rtid, true
}

func (w *World) BlockWithNbt(pos define.CubePos) (rtid uint32, nbt map[string]interface{}, found bool) {
	if w == nil || pos.OutOfYBounds() {
		// Fast way out.
		return chunk.AirRID, nil, false
	}
	chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	var c *mirror.ChunkData
	if w.lastChunk != nil && w.lastPos == chunkPos {
		c = w.lastChunk
	} else {
		c = w.chunk(chunkPos)
		w.lastChunk = c
		w.lastPos = chunkPos
	}
	if c == nil {
		return chunk.AirRID, nil, false
	}
	rtid = c.Chunk.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
	if nbt, hasK := c.BlockNbts[pos]; hasK {
		return rtid, nbt, true
	} else {
		return rtid, nil, true
	}
}

func (w *World) SetBlock(pos define.CubePos, rtid uint32) (success bool) {
	if w == nil || pos.OutOfYBounds() {
		// Fast way out.
		return false
	}
	chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	c := w.chunk(chunkPos)
	if c == nil {
		return false
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	c.Chunk.SetBlock(x, y, z, 0, rtid)
	return true
}

func (w *World) UpdateBlock(pos define.CubePos, rtid uint32) (origBlockRTID uint32, success bool) {
	if w == nil || pos.OutOfYBounds() {
		// Fast way out.
		return chunk.AirRID, false
	}
	chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	var c *mirror.ChunkData
	if w.lastChunk != nil && w.lastPos == chunkPos {
		c = w.lastChunk
	} else {
		c = w.chunk(chunkPos)
		w.lastChunk = c
		w.lastPos = chunkPos
	}
	if c == nil {
		return chunk.AirRID, false
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	origBlockRTID = c.Chunk.Block(x, y, z, 0)
	c.Chunk.SetBlock(x, y, z, 0, rtid)
	return origBlockRTID, true
}

func (w *World) SetBlockNbt(pos define.CubePos, nbt map[string]interface{}) (success bool) {
	if w == nil || pos.OutOfYBounds() {
		// Fast way out.
		return false
	}
	chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	var c *mirror.ChunkData
	if w.lastChunk != nil && w.lastPos == chunkPos {
		c = w.lastChunk
	} else {
		c = w.chunk(chunkPos)
		w.lastChunk = c
		w.lastPos = chunkPos
	}
	if c == nil {
		return false
	}
	if nbtBlockPos, success := define.GetCubePosFromNBT(nbt); success {
		if c.BlockNbts == nil {
			c.BlockNbts = make(map[define.CubePos]map[string]interface{})
		}
		c.BlockNbts[nbtBlockPos] = nbt
	}
	// c.BlockNbts
	return true
}
func (w *World) SetBlockWithNbt(pos define.CubePos, rtid uint32, nbt map[string]interface{}) (success bool) {
	if w == nil || pos.OutOfYBounds() {
		// Fast way out.
		return false
	}
	chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	var c *mirror.ChunkData
	if w.lastChunk != nil && w.lastPos == chunkPos {
		c = w.lastChunk
	} else {
		c = w.chunk(chunkPos)
		w.lastChunk = c
		w.lastPos = chunkPos
	}
	if c == nil {
		return false
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	c.Chunk.SetBlock(x, y, z, 0, rtid)
	if nbtBlockPos, success := define.GetCubePosFromNBT(nbt); success {
		c.BlockNbts[nbtBlockPos] = nbt
	}
	// c.BlockNbts
	return true
}

func NewWorld(provider mirror.ChunkProvider) *World {
	return &World{provider: provider}
}
