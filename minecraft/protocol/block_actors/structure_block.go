package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 结构方块
type StructureBlock struct {
	general.BlockActor `mapstructure:",squash"`
	AnimationMode      byte    `mapstructure:"animationMode"`    // TAG_Byte(1) = 0
	AnimationSeconds   float32 `mapstructure:"animationSeconds"` // TAG_Float(6) = 0
	Data               int32   `mapstructure:"data"`             // TAG_Int(4) = 1
	DataField          string  `mapstructure:"dataField"`        // TAG_String(8) = ""
	IgnoreEntities     byte    `mapstructure:"ignoreEntities"`   // TAG_Byte(1) = 0
	IncludePlayers     byte    `mapstructure:"includePlayers"`   // TAG_Byte(1) = 0
	Integrity          float32 `mapstructure:"integrity"`        // TAG_Float(6) = 100
	IsPowered          byte    `mapstructure:"isPowered"`        // TAG_Byte(1) = 0
	Mirror             byte    `mapstructure:"mirror"`           // TAG_Byte(1) = 0
	RedstoneSaveMode   int32   `mapstructure:"redstoneSaveMode"` // TAG_Int(4) = 0
	RemoveBlocks       byte    `mapstructure:"removeBlocks"`     // TAG_Byte(1) = 0
	Rotation           byte    `mapstructure:"rotation"`         // TAG_Byte(1) = 0
	Seed               int64   `mapstructure:"seed"`             // TAG_Long(5) = 0
	ShowBoundingBox    byte    `mapstructure:"showBoundingBox"`  // TAG_Byte(1) = 0
	StructureName      string  `mapstructure:"structureName"`    // TAG_String(8) = ""
	XStructureOffset   int32   `mapstructure:"xStructureOffset"` // TAG_Int(4) = 0
	XStructureSize     int32   `mapstructure:"xStructureSize"`   // TAG_Int(4) = 5
	YStructureOffset   int32   `mapstructure:"yStructureOffset"` // TAG_Int(4) = -1
	YStructureSize     int32   `mapstructure:"yStructureSize"`   // TAG_Int(4) = 5
	ZStructureOffset   int32   `mapstructure:"zStructureOffset"` // TAG_Int(4) = 0
	ZStructureSize     int32   `mapstructure:"zStructureSize"`   // TAG_Int(4) = 5
}

// ID ...
func (*StructureBlock) ID() string {
	return IDStructureBlock
}

func (s *StructureBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &s.BlockActor)
	io.Uint8(&s.IsPowered)
	io.Varint32(&s.Data)
	io.Varint32(&s.RedstoneSaveMode)
	io.Varint32(&s.XStructureOffset)
	io.Varint32(&s.YStructureOffset)
	io.Varint32(&s.ZStructureOffset)
	io.Varint32(&s.XStructureSize)
	io.Varint32(&s.YStructureSize)
	io.Varint32(&s.ZStructureSize)
	io.String(&s.StructureName)
	io.String(&s.DataField)
	io.Uint8(&s.IgnoreEntities)
	io.Uint8(&s.IncludePlayers)
	io.Uint8(&s.RemoveBlocks)
	io.Uint8(&s.ShowBoundingBox)
	protocol.NBTInt(&s.Rotation, io.Varuint32)
	protocol.NBTInt(&s.Mirror, io.Varuint32)
	protocol.NBTInt(&s.AnimationMode, io.Varuint32)
	io.Float32(&s.AnimationSeconds)
	io.Float32(&s.Integrity)
	io.Varint64(&s.Seed)
}
