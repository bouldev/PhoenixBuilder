package NBTAssigner

import (
	"fmt"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/go-gl/mathgl/mgl32"
)

// 从 s.BlockEntity.Block.NBT 提取物品展示框数据，
// 并保存在 s.StructureBlockData 中
func (s *StructureBlock) Decode() error {
	// 初始化
	var normal bool = false
	var animationMode byte
	var animationSeconds float32
	var data int32
	var dataField string
	var ignoreEntities bool
	var includePlayers bool
	var integrity float32
	var mirror byte
	var redstoneSaveMode int32
	var removeBlocks bool
	var rotation byte
	var seed int64
	var showBoundingBox bool
	var structureName string
	var xStructureOffset int32
	var xStructureSize int32
	var yStructureOffset int32
	var yStructureSize int32
	var zStructureOffset int32
	var zStructureSize int32
	// animationMode
	_, ok := s.BlockEntity.Block.NBT["animationMode"]
	if ok {
		animationMode, normal = s.BlockEntity.Block.NBT["animationMode"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"animationMode\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// animationSeconds
	_, ok = s.BlockEntity.Block.NBT["animationSeconds"]
	if ok {
		animationSeconds, normal = s.BlockEntity.Block.NBT["animationSeconds"].(float32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"animationSeconds\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// data
	_, ok = s.BlockEntity.Block.NBT["data"]
	if ok {
		data, normal = s.BlockEntity.Block.NBT["data"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"data\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// dataField
	_, ok = s.BlockEntity.Block.NBT["dataField"]
	if ok {
		dataField, normal = s.BlockEntity.Block.NBT["dataField"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"dataField\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// ignoreEntities
	_, ok = s.BlockEntity.Block.NBT["ignoreEntities"]
	if ok {
		got, normal := s.BlockEntity.Block.NBT["ignoreEntities"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"ignoreEntities\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			ignoreEntities = false
		} else {
			ignoreEntities = true
		}
	}
	// includePlayers
	_, ok = s.BlockEntity.Block.NBT["includePlayers"]
	if ok {
		got, normal := s.BlockEntity.Block.NBT["includePlayers"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"includePlayers\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			includePlayers = false
		} else {
			includePlayers = true
		}
	}
	// integrity
	_, ok = s.BlockEntity.Block.NBT["integrity"]
	if ok {
		integrity, normal = s.BlockEntity.Block.NBT["integrity"].(float32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"integrity\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// mirror
	_, ok = s.BlockEntity.Block.NBT["mirror"]
	if ok {
		mirror, normal = s.BlockEntity.Block.NBT["mirror"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"mirror\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// redstoneSaveMode
	_, ok = s.BlockEntity.Block.NBT["redstoneSaveMode"]
	if ok {
		redstoneSaveMode, normal = s.BlockEntity.Block.NBT["redstoneSaveMode"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"redstoneSaveMode\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// removeBlocks
	_, ok = s.BlockEntity.Block.NBT["removeBlocks"]
	if ok {
		got, normal := s.BlockEntity.Block.NBT["removeBlocks"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"removeBlocks\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			removeBlocks = false
		} else {
			removeBlocks = true
		}
	}
	// rotation
	_, ok = s.BlockEntity.Block.NBT["rotation"]
	if ok {
		rotation, normal = s.BlockEntity.Block.NBT["rotation"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"rotation\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// seed
	_, ok = s.BlockEntity.Block.NBT["seed"]
	if ok {
		seed, normal = s.BlockEntity.Block.NBT["seed"].(int64)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"seed\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// showBoundingBox
	_, ok = s.BlockEntity.Block.NBT["showBoundingBox"]
	if ok {
		got, normal := s.BlockEntity.Block.NBT["showBoundingBox"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"showBoundingBox\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			showBoundingBox = false
		} else {
			showBoundingBox = true
		}
	}
	// structureName
	_, ok = s.BlockEntity.Block.NBT["structureName"]
	if ok {
		structureName, normal = s.BlockEntity.Block.NBT["structureName"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"structureName\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// xStructureOffset
	_, ok = s.BlockEntity.Block.NBT["xStructureOffset"]
	if ok {
		xStructureOffset, normal = s.BlockEntity.Block.NBT["xStructureOffset"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"xStructureOffset\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// xStructureSize
	_, ok = s.BlockEntity.Block.NBT["xStructureSize"]
	if ok {
		xStructureSize, normal = s.BlockEntity.Block.NBT["xStructureSize"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"xStructureSize\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// yStructureOffset
	_, ok = s.BlockEntity.Block.NBT["yStructureOffset"]
	if ok {
		yStructureOffset, normal = s.BlockEntity.Block.NBT["yStructureOffset"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"yStructureOffset\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// yStructureSize
	_, ok = s.BlockEntity.Block.NBT["yStructureSize"]
	if ok {
		yStructureSize, normal = s.BlockEntity.Block.NBT["yStructureSize"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"yStructureSize\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// zStructureOffset
	_, ok = s.BlockEntity.Block.NBT["zStructureOffset"]
	if ok {
		zStructureOffset, normal = s.BlockEntity.Block.NBT["zStructureOffset"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"zStructureOffset\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// zStructureSize
	_, ok = s.BlockEntity.Block.NBT["zStructureSize"]
	if ok {
		zStructureSize, normal = s.BlockEntity.Block.NBT["zStructureSize"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at s.BlockEntity.Block.NBT[\"zStructureSize\"]; s.BlockEntity.Block.NBT = %#v", s.BlockEntity.Block.NBT)
		}
	}
	// return
	s.StructureBlockData = StructureBlockData{
		AnimationMode:    animationMode,
		AnimationSeconds: animationSeconds,
		Data:             data,
		DataField:        dataField,
		IgnoreEntities:   ignoreEntities,
		IncludePlayers:   includePlayers,
		Integrity:        integrity,
		Mirror:           mirror,
		RedstoneSaveMode: redstoneSaveMode,
		RemoveBlocks:     removeBlocks,
		Rotation:         rotation,
		Seed:             seed,
		ShowBoundingBox:  showBoundingBox,
		StructureName:    structureName,
		XStructureOffset: xStructureOffset,
		XStructureSize:   xStructureSize,
		YStructureOffset: yStructureOffset,
		YStructureSize:   yStructureSize,
		ZStructureOffset: zStructureOffset,
		ZStructureSize:   zStructureSize,
	}
	return nil
}

// 放置一个结构方块并写入结构方块数据
func (s *StructureBlock) WriteData() error {
	// 初始化
	api := s.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 放置结构方块
	if s.BlockEntity.AdditionalData.FastMode {
		// 以快速模式放置方块
		err := api.SetBlockAsync(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	} else {
		// 正常放置方块
		err := s.BlockEntity.Interface.SetBlock(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 向结构方块写入数据
	api.WritePacket(&packet.StructureBlockUpdate{
		Position:           s.BlockEntity.AdditionalData.Position,
		StructureName:      s.StructureBlockData.StructureName,
		DataField:          s.StructureBlockData.DataField,
		IncludePlayers:     s.StructureBlockData.IncludePlayers,
		ShowBoundingBox:    s.StructureBlockData.ShowBoundingBox,
		StructureBlockType: s.StructureBlockData.Data,
		Settings: protocol.StructureSettings{
			PaletteName:           "default",
			IgnoreEntities:        s.StructureBlockData.IgnoreEntities,
			IgnoreBlocks:          s.StructureBlockData.RemoveBlocks,
			AllowNonTickingChunks: true,
			Size: [3]int32{
				s.StructureBlockData.XStructureSize,
				s.StructureBlockData.YStructureSize,
				s.StructureBlockData.ZStructureSize,
			},
			Offset: [3]int32{
				s.StructureBlockData.XStructureOffset,
				s.StructureBlockData.YStructureOffset,
				s.StructureBlockData.ZStructureOffset,
			},
			LastEditingPlayerUniqueID: api.ClientInfo.EntityUniqueID,
			Rotation:                  s.StructureBlockData.Rotation,
			Mirror:                    s.StructureBlockData.Mirror,
			AnimationMode:             s.StructureBlockData.AnimationMode,
			AnimationDuration:         s.StructureBlockData.AnimationSeconds,
			Integrity:                 s.StructureBlockData.Integrity,
			Seed:                      uint32(s.StructureBlockData.Seed),
			Pivot: mgl32.Vec3{
				(float32(s.StructureBlockData.XStructureSize) - 1) / 2,
				(float32(s.StructureBlockData.YStructureSize) - 1) / 2,
				(float32(s.StructureBlockData.ZStructureSize) - 1) / 2,
			},
		},
		RedstoneSaveMode: s.StructureBlockData.RedstoneSaveMode,
		ShouldTrigger:    false,
		Waterlogged:      false,
	})
	// 返回值
	return nil
}
