package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 蜂巢 或 蜂箱 内的实体
type BeehiveOccupants struct {
	ActorIdentifier string         `nbt:"ActorIdentifier"` // TAG_String(8) = "minecraft:bee<>"
	SaveData        map[string]any `nbt:"SaveData"`        // TAG_Compound(10)
	TicksLeftToStay int32          `nbt:"TicksLeftToStay"` // TAG_Int(4) = 0
}

func (b *BeehiveOccupants) Marshal(r protocol.IO) {
	r.String(&b.ActorIdentifier)
	r.Varint32(&b.TicksLeftToStay)
	r.NBTWithLength(&b.SaveData)
}

func (b *BeehiveOccupants) ToNBT() map[string]any {
	return map[string]any{
		"ActorIdentifier": b.ActorIdentifier,
		"SaveData":        b.SaveData,
		"TicksLeftToStay": b.TicksLeftToStay,
	}
}

func (b *BeehiveOccupants) FromNBT(x map[string]any) {
	b.ActorIdentifier = x["ActorIdentifier"].(string)
	b.SaveData = x["SaveData"].(map[string]any)
	b.TicksLeftToStay = x["TicksLeftToStay"].(int32)
}
