package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 结构方块
type StructureBlock struct {
	general.BlockActor
	AnimationMode    uint32  `nbt:"animationMode"`    // * TAG_Byte(1) = 0
	AnimationSeconds float32 `nbt:"animationSeconds"` // TAG_Float(6) = 0
	Data             int32   `nbt:"data"`             // TAG_Int(4) = 1
	DataField        string  `nbt:"dataField"`        // TAG_String(8) = ""
	IgnoreEntities   byte    `nbt:"ignoreEntities"`   // TAG_Byte(1) = 0
	IncludePlayers   byte    `nbt:"includePlayers"`   // TAG_Byte(1) = 0
	Integrity        float32 `nbt:"integrity"`        // TAG_Float(6) = 100
	IsPowered        byte    `nbt:"isPowered"`        // TAG_Byte(1) = 0
	Mirror           uint32  `nbt:"mirror"`           // * TAG_Byte(1) = 0
	RedstoneSaveMode int32   `nbt:"redstoneSaveMode"` // TAG_Int(4) = 0
	RemoveBlocks     byte    `nbt:"removeBlocks"`     // TAG_Byte(1) = 0
	Rotation         uint32  `nbt:"rotation"`         // * TAG_Byte(1) = 0
	Seed             int64   `nbt:"seed"`             // TAG_Long(5) = 0
	ShowBoundingBox  byte    `nbt:"showBoundingBox"`  // TAG_Byte(1) = 0
	StructureName    string  `nbt:"structureName"`    // TAG_String(8) = ""
	XStructureOffset int32   `nbt:"xStructureOffset"` // TAG_Int(4) = 0
	XStructureSize   int32   `nbt:"xStructureSize"`   // TAG_Int(4) = 5
	YStructureOffset int32   `nbt:"yStructureOffset"` // TAG_Int(4) = -1
	YStructureSize   int32   `nbt:"yStructureSize"`   // TAG_Int(4) = 5
	ZStructureOffset int32   `nbt:"zStructureOffset"` // TAG_Int(4) = 0
	ZStructureSize   int32   `nbt:"zStructureSize"`   // TAG_Int(4) = 5
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
	io.Varuint32(&s.Rotation)
	io.Varuint32(&s.Mirror)
	io.Varuint32(&s.AnimationMode)
	io.Float32(&s.AnimationSeconds)
	io.Float32(&s.Integrity)
	io.Varint64(&s.Seed)
}

func (s *StructureBlock) ToNBT() map[string]any {
	return utils.MergeMaps(
		s.BlockActor.ToNBT(),
		map[string]any{
			"animationMode":    byte(s.AnimationMode),
			"animationSeconds": s.AnimationSeconds,
			"data":             s.Data,
			"dataField":        s.DataField,
			"ignoreEntities":   s.IgnoreEntities,
			"includePlayers":   s.IncludePlayers,
			"integrity":        s.Integrity,
			"isPowered":        s.IsPowered,
			"mirror":           byte(s.Mirror),
			"redstoneSaveMode": s.RedstoneSaveMode,
			"removeBlocks":     s.RemoveBlocks,
			"rotation":         byte(s.Rotation),
			"seed":             s.Seed,
			"showBoundingBox":  s.ShowBoundingBox,
			"structureName":    s.StructureName,
			"xStructureOffset": s.XStructureOffset,
			"xStructureSize":   s.XStructureSize,
			"yStructureOffset": s.YStructureOffset,
			"yStructureSize":   s.YStructureSize,
			"zStructureOffset": s.ZStructureOffset,
			"zStructureSize":   s.ZStructureSize,
		},
	)
}

func (s *StructureBlock) FromNBT(x map[string]any) {
	s.BlockActor.FromNBT(x)
	s.AnimationMode = uint32(x["animationMode"].(byte))
	s.AnimationSeconds = x["animationSeconds"].(float32)
	s.Data = x["data"].(int32)
	s.DataField = x["dataField"].(string)
	s.IgnoreEntities = x["ignoreEntities"].(byte)
	s.IncludePlayers = x["includePlayers"].(byte)
	s.Integrity = x["integrity"].(float32)
	s.IsPowered = x["isPowered"].(byte)
	s.Mirror = uint32(x["mirror"].(byte))
	s.RedstoneSaveMode = x["redstoneSaveMode"].(int32)
	s.RemoveBlocks = x["removeBlocks"].(byte)
	s.Rotation = uint32(x["rotation"].(byte))
	s.Seed = x["seed"].(int64)
	s.ShowBoundingBox = x["showBoundingBox"].(byte)
	s.StructureName = x["structureName"].(string)
	s.XStructureOffset = x["xStructureOffset"].(int32)
	s.XStructureSize = x["xStructureSize"].(int32)
	s.YStructureOffset = x["yStructureOffset"].(int32)
	s.YStructureSize = x["yStructureSize"].(int32)
	s.ZStructureOffset = x["zStructureOffset"].(int32)
	s.ZStructureSize = x["zStructureSize"].(int32)
}
