package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 唱片机
type Jukebox struct {
	general.BlockActor
	RecordItem protocol.Optional[protocol.Item] `nbt:"RecordItem"` // TAG_Compound(10)
}

// ID ...
func (j *Jukebox) ID() string {
	return IDJukebox
}

func (j *Jukebox) Marshal(io protocol.IO) {
	protocol.Single(io, &j.BlockActor)
	protocol.OptionalMarshaler(io, &j.RecordItem)
}

func (j *Jukebox) ToNBT() map[string]any {
	var temp map[string]any
	if recordItem, has := j.RecordItem.Value(); has {
		temp = map[string]any{
			"RecordItem": recordItem.ToNBT(),
		}
	}
	return utils.MergeMaps(
		j.BlockActor.ToNBT(),
		temp,
	)
}

func (j *Jukebox) FromNBT(x map[string]any) {
	j.BlockActor.FromNBT(x)

	if recordItem, has := x["RecordItem"].(map[string]any); has {
		new := protocol.Item{}
		new.FromNBT(recordItem)
		j.RecordItem = protocol.Optional[protocol.Item]{Set: true, Val: new}
	}
}
