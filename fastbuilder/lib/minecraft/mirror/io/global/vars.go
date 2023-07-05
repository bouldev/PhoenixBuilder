package global

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror"
)

type ChunkWriteFn func(chunk *mirror.ChunkData)
