package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 蜂巢 或 蜂箱 内的实体
type BeehiveOccupants struct {
	ActorIdentifier string         `mapstructure:"ActorIdentifier"` // TAG_String(8) = "minecraft:bee<>"
	SaveData        map[string]any `mapstructure:"SaveData"`        // TAG_Compound(10)
	TicksLeftToStay int32          `mapstructure:"TicksLeftToStay"` // TAG_Int(4) = 0
}

func (b *BeehiveOccupants) Marshal(r protocol.IO) {
	r.String(&b.ActorIdentifier)
	r.Varint32(&b.TicksLeftToStay)
	r.NBTWithLength(&b.SaveData)
}
