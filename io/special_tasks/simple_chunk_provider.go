// +build !is_tweak

package special_tasks

import (
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
)

type SimpleChunkProvider struct {
	ChunkMap map[define.ChunkPos]*mirror.ChunkData
}

func (_ SimpleChunkProvider) Write(_ *mirror.ChunkData) error {
	return nil
}

func (p SimpleChunkProvider) Get(pos define.ChunkPos) *mirror.ChunkData {
	return p.ChunkMap[pos]
}