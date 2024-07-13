package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 网易特有方块，可能被用于储存模组的自定义数据
type ModBlock struct {
	general.BlockActor `mapstructure:",squash"`
	Tick               byte   `mapstructure:"_tick"`       // TAG_Byte(1) = 0
	Movable            byte   `mapstructure:"_movable"`    // TAG_Byte(1) = 1
	ExData             int32  `mapstructure:"exData"`      // TAG_Int(4) = 0
	BlockName          string `mapstructure:"_blockName"`  // TAG_String(8) = ""
	UniqueId           int64  `mapstructure:"_uniqueId"`   // TAG_Long(5) = 0
	TickClient         byte   `mapstructure:"_tickClient"` // TAG_Byte(1) = 0
}

// ID ...
func (*ModBlock) ID() string {
	return IDModBlock
}

func (m *ModBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &m.BlockActor)
	io.Uint8(&m.Tick)
	io.Uint8(&m.Movable)
	protocol.NBTInt(&m.ExData, io.Varuint32)
	io.String(&m.BlockName)
	io.Varint64(&m.UniqueId)
	io.Uint8(&m.TickClient)
}
