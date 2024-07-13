package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 唱片机
type Jukebox struct {
	general.BlockActor `mapstructure:",squash"`
	RecordItem         *protocol.Item `mapstructure:"RecordItem,omitempty"` // TAG_Compound(10)
}

// ID ...
func (j *Jukebox) ID() string {
	return IDJukebox
}

func (j *Jukebox) Marshal(io protocol.IO) {
	f := func() *protocol.Item {
		if j.RecordItem == nil {
			j.RecordItem = new(protocol.Item)
		}
		return j.RecordItem
	}

	protocol.Single(io, &j.BlockActor)
	protocol.NBTOptionalFunc(io, j.RecordItem, f, true, io.NBTItem)
}
