package global

import (
	"phoenixbuilder/mirror"
)

type ChunkWriteFn func(chunk *mirror.ChunkData)