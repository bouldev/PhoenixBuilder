package hop_planner

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
)

type HopToChunksPos struct {
	CenterPos        define.CubePos
	InRangeChunksPos []*ChunkPosToHop
}

type ChunkPosToHop struct {
	Pos       define.ChunkPos
	MasterHop *HopToChunksPos
}

type ExportedHopsMap map[define.CubePos]*HopToChunksPos
type ExportedChunksMap map[define.ChunkPos]*ChunkPosToHop
