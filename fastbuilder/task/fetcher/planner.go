package fetcher

import (
	"fmt"
)

func PlanHopSwapPath(sx, sz, ex, ez, reachableChunks int) (hopPath []*ExportHopPos, allRequiredChunks ExportedChunksMap) {
	// reachableChunks is how many chunks in x or z direction can be fetched in a specific point
	chunkSize := 16
	reachableBlocks := chunkSize * reachableChunks
	alignSX := ((sx - chunkSize + 1) / chunkSize) * chunkSize
	alignSZ := ((sz - chunkSize + 1) / chunkSize) * chunkSize
	alignCEX := ((ex ) / chunkSize) * chunkSize
	alignCEZ := ((ez ) / chunkSize) * chunkSize
	alignMEX := ((ex + chunkSize) / chunkSize) * chunkSize
	alignMEZ := ((ez + chunkSize) / chunkSize) * chunkSize
	hopXPoints := ((alignMEX - alignSX) + reachableBlocks - 1) / reachableBlocks
	hopZPoints := ((alignMEZ - alignSZ) + reachableBlocks - 1) / reachableBlocks
	hopXArray := []int{}
	hopZArray := []int{}
	preferredHalfHopXSpace := int(float32(alignMEX-alignSX) / float32(hopXPoints*2))
	preferredHalfHopZSpace := int(float32(alignMEZ-alignSZ) / float32(hopZPoints*2))
	hopXOrigin := alignSX + preferredHalfHopXSpace
	hopZOrigin := alignSZ + preferredHalfHopZSpace
	{
		for i := 0; i < hopXPoints; i++ {
			hopPoint := hopXOrigin + i*2*preferredHalfHopXSpace
			//fmt.Printf("NEW hopPoint (X+) %d\n", hopPoint)
			hopXArray = append(hopXArray, hopPoint)
		}
	}
	{
		for i := 0; i < hopZPoints; i++ {
			hopPoint := hopZOrigin + i*2*preferredHalfHopZSpace
			//fmt.Printf("NEW hopPoint (Z+) %d\n", hopPoint)
			hopZArray = append(hopZArray, hopPoint)
		}
	}
	hopPoints := []*ExportHopPos{}
	hopPointsMap := map[ChunkPosDefine]*ExportHopPos{}
	for i, x := range hopXArray {
		if i%2 == 0 {
			for _, z := range hopZArray {
				p := ChunkPosDefine{x, z}
				hp := &ExportHopPos{Pos: p, LinkedChunk: make([]*ExportedChunkPos, 0)}
				hopPoints = append(hopPoints, hp)
				hopPointsMap[p] = hp
			}
		} else {
			for zi, _ := range hopZArray {
				z := hopZArray[len(hopZArray)-1-zi]
				p := ChunkPosDefine{x, z}
				hp := &ExportHopPos{Pos: p, LinkedChunk: make([]*ExportedChunkPos, 0)}
				hopPoints = append(hopPoints, hp)
				hopPointsMap[p] = hp
			}
		}
	}
	chunkPosMap := ExportedChunksMap{}
	for xi := alignSX / chunkSize; xi < alignCEX/chunkSize; xi++ {
		for zi := alignSZ / chunkSize; zi < alignCEZ/chunkSize; zi++ {
			x, z := xi*chunkSize, zi*chunkSize
			xHalfHops := ((x - alignSX) / preferredHalfHopXSpace)
			hopXPoint := hopXOrigin + (xHalfHops/2)*2*preferredHalfHopXSpace
			zHalfHops := ((z - alignSZ) / preferredHalfHopZSpace)
			hopZPoint := hopZOrigin + (zHalfHops/2)*2*preferredHalfHopZSpace
			chunkPos := &ExportedChunkPos{
				Pos: ChunkPosDefine{x, z},
				MasterHop: hopPointsMap[ChunkPosDefine{hopXPoint, hopZPoint}],
				CachedMark: false,
			}
			chunkPosMap[ChunkPosDefine{x, z}] = chunkPos
			//fmt.Printf("x=%d, z=%d\n", x, z)
			//fmt.Printf("REACHING hopPoint (%d, %d)\n", hopXPoint, hopZPoint)
			hopPointsMap[ChunkPosDefine{hopXPoint, hopZPoint}].LinkedChunk = append(hopPointsMap[ChunkPosDefine{hopXPoint, hopZPoint}].LinkedChunk, chunkPos)
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
