package world_provider

import (
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/environment"
)

var CurrentWorld *world.World = nil

func Create(env *environment.PBEnvironment) *world.World {
	intw:=world.New(&StubLogger{},32)
	intw.Provider(NewOnlineWorldProvider(env))
	return intw
}

func NewWorld(env *environment.PBEnvironment) {
	ChunkCache=make(map[world.ChunkPos]*packet.LevelChunk)
	CurrentWorld=Create(env)
	firstLoaded=false
}

func DestroyWorld() {
	firstLoaded=false
	CurrentWorld=nil
	ChunkCache=nil
}

func init() {
	InitRuntimeIdsWithoutMinecraftPrefix()
}