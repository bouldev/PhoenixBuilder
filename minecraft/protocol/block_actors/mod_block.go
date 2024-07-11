package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 网易特有方块，可能被用于储存模组的自定义数据
type ModBlock struct {
	general.BlockActor
	Tick       byte   `nbt:"_tick"`       // TAG_Byte(1) = 0
	Movable    byte   `nbt:"_movable"`    // TAG_Byte(1) = 1
	ExData     uint32 `nbt:"exData"`      // * TAG_Int(4) = 0
	BlockName  string `nbt:"_blockName"`  // TAG_String(8) = ""
	UniqueId   int64  `nbt:"_uniqueId"`   // TAG_Long(5) = 0
	TickClient byte   `nbt:"_tickClient"` // TAG_Byte(1) = 0
}

// ID ...
func (*ModBlock) ID() string {
	return IDModBlock
}

func (m *ModBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &m.BlockActor)
	io.Uint8(&m.Tick)
	io.Uint8(&m.Movable)
	io.Varuint32(&m.ExData)
	io.String(&m.BlockName)
	io.Varint64(&m.UniqueId)
	io.Uint8(&m.TickClient)
}

func (m *ModBlock) ToNBT() map[string]any {
	return utils.MergeMaps(
		m.BlockActor.ToNBT(),
		map[string]any{
			"_tick":       m.Tick,
			"_movable":    m.Movable,
			"exData":      int32(m.ExData),
			"_blockName":  m.BlockName,
			"_uniqueId":   m.UniqueId,
			"_tickClient": m.TickClient,
		},
	)
}

func (m *ModBlock) FromNBT(x map[string]any) {
	m.BlockActor.FromNBT(x)
	m.Tick = x["_tick"].(byte)
	m.Movable = x["_movable"].(byte)
	m.ExData = uint32(x["exData"].(int32))
	m.BlockName = x["_blockName"].(string)
	m.UniqueId = x["_uniqueId"].(int64)
	m.TickClient = x["_tickClient"].(byte)
}
