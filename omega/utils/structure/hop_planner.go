package structure

import (
	"phoenixbuilder/mirror/define"
)

type ExportHopPos struct {
	Pos         define.CubePos
	CachedMark  bool
	LinkedChunk []*ExportedChunkPos
}

type ExportHopPosMap map[define.CubePos]*ExportHopPos

type ExportedChunkPos struct {
	Pos          define.ChunkPos
	MasterHop    *ExportHopPos
	MasterHopMap *ExportHopPosMap
	CachedMark   bool
}

type ExportedChunksMap map[define.ChunkPos]*ExportedChunkPos

func (o *ExportHopPosMap) Squeeze() *ExportHopPosMap {
	squeezed := ExportHopPosMap{}
	for pos, hop := range *o {
		if !hop.CachedMark {
			squeezed[pos] = hop
		}
	}
	*o = squeezed
	return o
}

func (o *ExportedChunksMap) Hit(pos define.ChunkPos) {
	if c, hasK := (*o)[pos]; hasK {
		c.CachedMark = true
		allCached := true
		for _, lc := range c.MasterHop.LinkedChunk {
			if !lc.CachedMark {
				allCached = false
				break
			}
		}
		if allCached {
			masterHopPos := c.MasterHop.Pos
			for _, ck := range c.MasterHop.LinkedChunk {
				delete(*o, ck.Pos)
			}
			delete(*c.MasterHopMap, masterHopPos)
		}
	}
}

func PlanHopSwapPath(sx, sz, ex, ez, receptRangeByChunk int) (hopPath *ExportHopPosMap, allRequiredChunks *ExportedChunksMap) {
	chunkSize := 16
	receptRange := chunkSize * receptRangeByChunk
	alignSX := ((sx - chunkSize + 1) / chunkSize) * chunkSize
	alignSZ := ((sz - chunkSize + 1) / chunkSize) * chunkSize
	alignEX := ((ex) / chunkSize) * chunkSize
	alignEZ := ((ez) / chunkSize) * chunkSize
	hopXPoints := ((alignEX - alignSX + chunkSize) + receptRange - 1) / receptRange
	hopZPoints := ((alignEZ - alignSZ + chunkSize) + receptRange - 1) / receptRange
	hopXArray := []int{}
	hopZArray := []int{}
	preferHalfHopXSpace := int(float32(alignEX-alignSX+chunkSize) / float32(hopXPoints*2))
	preferHalfHopZSpace := int(float32(alignEZ-alignSZ+chunkSize) / float32(hopZPoints*2))
	hopXStart := alignSX + preferHalfHopXSpace
	hopZStart := alignSZ + preferHalfHopZSpace
	{
		for i := 0; i < hopXPoints; i++ {
			hopPoint := hopXStart + i*2*preferHalfHopXSpace
			hopXArray = append(hopXArray, hopPoint)
		}
	}
	{
		for i := 0; i < hopZPoints; i++ {
			hopPoint := hopZStart + i*2*preferHalfHopZSpace
			hopZArray = append(hopZArray, hopPoint)
		}
	}
	hopPoints := &ExportHopPosMap{}
	for i, x := range hopXArray {
		if i%2 == 0 {
			for _, z := range hopZArray {
				p := define.CubePos{x, 320, z}
				hp := &ExportHopPos{Pos: p, LinkedChunk: make([]*ExportedChunkPos, 0)}
				(*hopPoints)[p] = hp
			}
		} else {
			for zi, _ := range hopZArray {
				z := hopZArray[len(hopZArray)-1-zi]
				p := define.CubePos{x, 320, z}
				hp := &ExportHopPos{Pos: p, LinkedChunk: make([]*ExportedChunkPos, 0)}
				(*hopPoints)[p] = hp
			}
		}
	}
	chunkPosMap := ExportedChunksMap{}
	for xi := alignSX / chunkSize; xi <= alignEX/chunkSize; xi++ {
		for zi := alignSZ / chunkSize; zi <= alignEZ/chunkSize; zi++ {
			x, z := xi*chunkSize, zi*chunkSize
			// fmt.Println(x,z)
			xHalfHops := ((x - alignSX) / preferHalfHopXSpace)
			hopXPoint := hopXStart + (xHalfHops/2)*2*preferHalfHopXSpace
			zHalfHops := ((z - alignSZ) / preferHalfHopZSpace)
			hopZPoint := hopZStart + (zHalfHops/2)*2*preferHalfHopZSpace
			pos := define.ChunkPos{int32(x >> 4), int32(z >> 4)}
			chunkPos := &ExportedChunkPos{
				Pos:          pos,
				MasterHop:    (*hopPoints)[define.CubePos{hopXPoint, 320, hopZPoint}],
				MasterHopMap: hopPoints,
				CachedMark:   false,
			}
			chunkPosMap[pos] = chunkPos
			(*hopPoints)[define.CubePos{hopXPoint, 320, hopZPoint}].LinkedChunk = append((*hopPoints)[define.CubePos{hopXPoint, 320, hopZPoint}].LinkedChunk, chunkPos)
		}
	}
	return hopPoints, &chunkPosMap
}
