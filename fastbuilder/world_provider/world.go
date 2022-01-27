package world_provider

import (
	"phoenixbuilder/minecraft"
	"phoenixbuilder/dragonfly/server/world"
)

var CurrentWorld *world.World

func Create(conn *minecraft.Conn) *world.World {
	intw:=world.New(&StubLogger{},32)
	intw.Provider(NewOnlineWorldProvider(conn))
	return intw
}

func Init(conn *minecraft.Conn) {
	CurrentWorld=Create(conn)
}