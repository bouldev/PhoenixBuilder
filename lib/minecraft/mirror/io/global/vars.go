package global

import (
	"phoenixbuilder/lib/minecraft/mirror"
)

type ChunkWriteFn func(chunk *mirror.ChunkData)
