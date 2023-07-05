package hop_planner

import (
	"math"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
)

func PlanHopSwapPath(sx, sz, ex, ez, receptiveRangeByChunk int) (hopPath []*HopToChunksPos, allRequiredChunks ExportedChunksMap) {
	// receptiveRangeByChunk is how many chunks in x or z a robot can fetch in a specific point
	if sx > ex {
		ex, sx = sx, ex
	}
	if ex == sx {
		ex++
	}
	if sz > ez {
		ez, sz = sz, ez
	}
	if ez == sz {
		ez++
	}
	chunkSize := 16
	getBL := func(v int) int {
		if v%16 != 0 {
			if v > 0 {
				v -= v % 16
			} else {
				v = v - 16 - v%16
			}

		}
		return v
	}
	getUR := func(v int) int {
		if v%16 != 16-1 {
			if v > 0 {
				v = v + 15 - v%16
			} else {
				v -= v%16 + 1
			}
		}
		return v
	}
	receptiveRange := chunkSize * receptiveRangeByChunk
	alignSX := getBL(sx)
	alignSZ := getBL(sz)
	alignEX := getBL(ex)
	alignEZ := getBL(ez)
	alignMaxEX := getUR(ex)
	alignMaxEZ := getUR(ez)
	hopXPoints := ((alignMaxEX - alignSX) + receptiveRange - 1) / receptiveRange
	hopZPoints := ((alignMaxEZ - alignSZ) + receptiveRange - 1) / receptiveRange
	hopXArray := []int{}
	hopZArray := []int{}
	preferHalfHopXSpace := int(math.Ceil(float64(alignMaxEX-alignSX) / float64(hopXPoints*2)))
	preferHalfHopZSpace := int(math.Ceil(float64(alignMaxEZ-alignSZ) / float64(hopZPoints*2)))
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
	hops := []define.CubePos{}
	hopPoints := ExportedHopsMap{}
	for i, x := range hopXArray {
		if i%2 == 0 {
			for _, z := range hopZArray {
				p := define.CubePos{x, 320, z}
				hp := &HopToChunksPos{CenterPos: p, InRangeChunksPos: make([]*ChunkPosToHop, 0)}
				hopPoints[p] = hp
				hops = append(hops, p)
			}
		} else {
			for zi, _ := range hopZArray {
				z := hopZArray[len(hopZArray)-1-zi]
				p := define.CubePos{x, 320, z}
				hp := &HopToChunksPos{CenterPos: p, InRangeChunksPos: make([]*ChunkPosToHop, 0)}
				hopPoints[p] = hp
				hops = append(hops, p)
			}
		}
	}
	chunkPosMap := ExportedChunksMap{}
	xLookUp := map[int]int{}
	for xi := alignSX / chunkSize; xi <= alignEX/chunkSize; xi++ {
		x := xi * chunkSize
		_s := ((x - alignSX) / (preferHalfHopXSpace * 2)) - 2
		if _s < 0 {
			_s = 0
		}
		_e := ((x - alignSX) / (preferHalfHopXSpace * 2)) + 2
		if _e > hopXPoints {
			_e = hopXPoints
		}
		m := -1
		l := 0
		for i := _s; i < _e; i++ {
			hopPoint := hopXStart + i*2*preferHalfHopXSpace
			d := 0
			if hopPoint > x {
				d = hopPoint - x
			} else {
				d = x - hopPoint
			}
			if m == -1 || d < m {
				m = d
				l = hopPoint
			}
		}
		if m == -1 {
			panic("should not happen")
		} else {
			//if m >= receptiveRange*chunkSize/2 {
			//	panic("should not happen")
			//}
			xLookUp[xi] = l
		}
	}
	zLookUp := map[int]int{}
	for zi := alignSZ / chunkSize; zi <= alignEZ/chunkSize; zi++ {
		z := zi * chunkSize
		_s := ((z - alignSZ) / (preferHalfHopZSpace * 2)) - 2
		if _s < 0 {
			_s = 0
		}
		_e := ((z - alignSZ) / (preferHalfHopZSpace * 2)) + 2
		if _e > hopZPoints {
			_e = hopZPoints
		}
		m := -1
		l := 0
		for i := _s; i < _e; i++ {
			hopPoint := hopZStart + i*2*preferHalfHopZSpace
			d := 0
			if hopPoint > z {
				d = hopPoint - z
			} else {
				d = z - hopPoint
			}
			if m == -1 || d < m {
				m = d
				l = hopPoint
			}
		}
		if m == -1 {
			panic("should not happen")
		} else {
			//if m >= receptiveRange*chunkSize/2 {
			//	panic("should not happen")
			//}
			zLookUp[zi] = l
		}
	}

	for zi := alignSZ / chunkSize; zi <= alignEZ/chunkSize; zi++ {
		for xi := alignSX / chunkSize; xi <= alignEX/chunkSize; xi++ {
			x, z := xi*chunkSize, zi*chunkSize
			xh, zh := xLookUp[xi], zLookUp[zi]
			pos := define.ChunkPos{int32(x >> 4), int32(z >> 4)}
			chunkPos := &ChunkPosToHop{
				Pos:       pos,
				MasterHop: hopPoints[define.CubePos{xh, 320, zh}],
			}
			//fmt.Println(x, z, hopXPoint, hopZPoint)
			chunkPosMap[pos] = chunkPos
			hopPoints[define.CubePos{xh, 320, zh}].InRangeChunksPos = append(hopPoints[define.CubePos{xh, 320, zh}].InRangeChunksPos, chunkPos)
		}
	}
	for _, p := range hops {
		if len(hopPoints[p].InRangeChunksPos) == 0 {
			continue
		}
		hopPath = append(hopPath, hopPoints[p])
	}

	return hopPath, chunkPosMap
}
