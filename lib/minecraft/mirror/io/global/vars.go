package global

import (
	"fastbuilder-core/lib/minecraft/mirror"
)

type ChunkWriteFn func(chunk *mirror.ChunkData)
