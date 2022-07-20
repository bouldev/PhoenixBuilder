package fetcher

import (
	"fmt"
)

func PlanHopSwapPath(sx, sz, ex, ez, receptRangeByChunk int) (hopPath []*ExportHopPos, allRequiredChunks ExportedChunksMap) {
	// receptRangeByChunk is how many chunks in x or z a robot can fetch in a specific point
	chunkSize := 16
	receptRange := chunkSize * receptRangeByChunk
	alignSX := ((sx - chunkSize + 1) / chunkSize) * chunkSize
	alignSZ := ((sz - chunkSize + 1) / chunkSize) * chunkSize
	alignCEX := ((ex ) / chunkSize) * chunkSize
	alignCEZ := ((ez ) / chunkSize) * chunkSize
	alignMEX := ((ex + chunkSize) / chunkSize) * chunkSize
	alignMEZ := ((ez + chunkSize) / chunkSize) * chunkSize
	hopXPoints := ((alignMEX - alignSX) + receptRange - 1) / receptRange
	hopZPoints := ((alignMEZ - alignSZ) + receptRange - 1) / receptRange
	// fmt.Println(alignSX, alignSZ, alignCEX, alignCEZ)
	// fmt.Println(alignSX, alignSZ, alignMEX, alignMEZ)
	// fmt.Println(hopXPoints, hopZPoints)
	hopXArray := []int{}
	hopZArray := []int{}
	preferHalfHopXSpace := int(float32(alignMEX-alignSX) / float32(hopXPoints*2))
	preferHalfHopZSpace := int(float32(alignMEZ-alignSZ) / float32(hopZPoints*2))
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
	hopPoints := []*ExportHopPos{}
	hopPointsLookUp := map[ChunkPosDefine]*ExportHopPos{}
	// snake folding
	for i, x := range hopXArray {
		if i%2 == 0 {
			for _, z := range hopZArray {
				p := ChunkPosDefine{x, z}
				hp := &ExportHopPos{Pos: p, LinkedChunk: make([]*ExportedChunkPos, 0)}
				hopPoints = append(hopPoints, hp)
				hopPointsLookUp[p] = hp
			}
		} else {
			for zi, _ := range hopZArray {
				z := hopZArray[len(hopZArray)-1-zi]
				p := ChunkPosDefine{x, z}
				hp := &ExportHopPos{Pos: p, LinkedChunk: make([]*ExportedChunkPos, 0)}
				hopPoints = append(hopPoints, hp)
				hopPointsLookUp[p] = hp
			}
		}
	}
	chunkPosMap := ExportedChunksMap{}
	for xi := alignSX / chunkSize; xi <= alignCEX/chunkSize; xi++ {
		for zi := alignSZ / chunkSize; zi <= alignCEZ/chunkSize; zi++ {
			x, z := xi*chunkSize, zi*chunkSize
			// fmt.Println(x,z)
			xHalfHops := ((x - alignSX) / preferHalfHopXSpace)
			hopXPoint := hopXStart + (xHalfHops/2)*2*preferHalfHopXSpace
			zHalfHops := ((z - alignSZ) / preferHalfHopZSpace)
			hopZPoint := hopZStart + (zHalfHops/2)*2*preferHalfHopZSpace
			chunkPos := &ExportedChunkPos{
				Pos:          ChunkPosDefine{x, z},
				MasterHop:  hopPointsLookUp[ChunkPosDefine{hopXPoint, hopZPoint}],
				CachedMark: false,
			}
			chunkPosMap[ChunkPosDefine{x, z}] = chunkPos
			hopPointsLookUp[ChunkPosDefine{hopXPoint, hopZPoint}].LinkedChunk = append(hopPointsLookUp[ChunkPosDefine{hopXPoint, hopZPoint}].LinkedChunk, chunkPos)
		}
	}
	return hopPoints, chunkPosMap
}


func CreateCacheHitFetcher(requiredChunks ExportedChunksMap,chunkPool map[ChunkPosDefine]ChunkDefine) (fetcher func(ChunkPosDefine,ChunkDefine)){
	return func(ecp ChunkPosDefine, cd ChunkDefine) {
		if c,hasK:=requiredChunks[ecp];hasK{
			fmt.Println("Hit Memory ",ecp)
			chunkPool[ecp]=cd
			c.CachedMark=true
		}
	}
}

func SimplifyHopPos(hopPath []*ExportHopPos) (simplifiedHopPos []*ExportHopPos){
	simplifiedHopPos=make([]*ExportHopPos, 0)
	for _,hp:=range hopPath{
		allCached:=true
		for _,lc:=range hp.LinkedChunk{
			if !lc.CachedMark{
				allCached=false
				break
			}
		}
		if !allCached{
			simplifiedHopPos = append(simplifiedHopPos, hp)
		}else{
			fmt.Printf("Master Node %v all fetched\n",hp.Pos)
		}
	}
	return simplifiedHopPos
}