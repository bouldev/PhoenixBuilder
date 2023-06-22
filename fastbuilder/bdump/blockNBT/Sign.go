package blockNBT

import (
	"fmt"
	"phoenixbuilder/GameControl/GlobalAPI"
	"phoenixbuilder/minecraft/protocol/packet"
)

// SignDatas 结构体用于描述告示牌的 NBT 在被解析后的数据
type SignDatas struct {
	TextOwner                   string // TextOwner(TAG_String) = ""
	IgnoreLighting              byte   // IgnoreLighting(TAG_Byte) = 0
	SignTextColor               int32  // SignTextColor(TAG_Int) = 0
	TextIgnoreLegacyBugResolved byte   // TextIgnoreLegacyBugResolved(TAG_Byte) = 0
	Text                        string // Text(TAG_String) = ""
}

// Sign 结构体用于描述一个完整的告示牌
type Sign struct {
	Package   *Package  // 该方块实体的详细数据
	SignDatas SignDatas // 告示牌数据
}

// 从 s.Package.Block.NBT 提取告示牌数据并保存在 s.SignDatas 中
func (s *Sign) Decode() error {
	var ok bool = false
	var normal bool = false
	var textOwner string = ""
	var ignoreLighting byte = byte(0)
	var signTextColor int32 = 0
	var textIgnoreLegacyBugResolved byte = byte(0)
	var text string = ""
	// 初始化
	_, ok = s.Package.Block.NBT["TextOwner"]
	if ok {
		textOwner, normal = s.Package.Block.NBT["TextOwner"].(string)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.Package.Block.NBT[\"TextOwner\"]; s.Package.Block.NBT = %#v", s.Package.Block.NBT)
		}
	}
	// TextOwner
	_, ok = s.Package.Block.NBT["IgnoreLighting"]
	if ok {
		ignoreLighting, normal = s.Package.Block.NBT["IgnoreLighting"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.Package.Block.NBT[\"IgnoreLighting\"]; s.Package.Block.NBT = %#v", s.Package.Block.NBT)
		}
	}
	// IgnoreLighting
	_, ok = s.Package.Block.NBT["SignTextColor"]
	if ok {
		signTextColor, normal = s.Package.Block.NBT["SignTextColor"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.Package.Block.NBT[\"SignTextColor\"]; s.Package.Block.NBT = %#v", s.Package.Block.NBT)
		}
	}
	// SignTextColor
	_, ok = s.Package.Block.NBT["TextIgnoreLegacyBugResolved"]
	if ok {
		textIgnoreLegacyBugResolved, normal = s.Package.Block.NBT["TextIgnoreLegacyBugResolved"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.Package.Block.NBT[\"TextIgnoreLegacyBugResolved\"]; s.Package.Block.NBT = %#v", s.Package.Block.NBT)
		}
	}
	// TextIgnoreLegacyBugResolved
	_, ok = s.Package.Block.NBT["Text"]
	if ok {
		text, normal = s.Package.Block.NBT["Text"].(string)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.Package.Block.NBT[\"Text\"]; s.Package.Block.NBT = %#v", s.Package.Block.NBT)
		}
	}
	// Text
	s.SignDatas = SignDatas{
		TextOwner:                   textOwner,
		IgnoreLighting:              ignoreLighting,
		SignTextColor:               signTextColor,
		TextIgnoreLegacyBugResolved: textIgnoreLegacyBugResolved,
		Text:                        text,
	}
	// 写入数据
	return nil
	// 返回值
}

// 放置一个告示牌并写入告示牌数据
func (s *Sign) WriteDatas() error {
	if s.Package.Datas.FastMode {
		err := s.Package.API.SetBlockFastly(s.Package.Datas.Position, s.Package.Block.Name, s.Package.Datas.StatesString)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		return nil
	}
	// 放置告示牌(快速导入模式下)
	{
		err := s.Package.API.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", s.Package.Datas.Position[0], s.Package.Datas.Position[1], s.Package.Datas.Position[2]), true)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 传送机器人到告示牌所在的位置
		err = s.Package.API.SendSettingsCommand(
			"replaceitem entity @s slot.hotbar 0 oak_sign",
			true,
		)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 先获取一个告示牌到快捷栏 0
		err = s.Package.API.ChangeSelectedHotbarSlot(0)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 切换手持物品栏到快捷栏 0
		uniqueID, err := s.Package.API.BackupStructure(
			GlobalAPI.MCStructure{
				BeginX: s.Package.Datas.Position[0] + 1,
				BeginY: s.Package.Datas.Position[1],
				BeginZ: s.Package.Datas.Position[2],
				SizeX:  1,
				SizeY:  1,
				SizeZ:  1,
			},
		)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 备份结构
		_, err = s.Package.API.SendWSCommandWithResponce(
			fmt.Sprintf(
				"setblock %d %d %d glass",
				s.Package.Datas.Position[0]+1,
				s.Package.Datas.Position[1],
				s.Package.Datas.Position[2],
			),
		)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 修改备份过的区域以用于放置方块
		err = s.Package.API.PlaceBlock(
			GlobalAPI.UseItemOnBlocks{
				HotbarSlotID: 0,
				BlockPos: [3]int32{
					s.Package.Datas.Position[0] + 1,
					s.Package.Datas.Position[1],
					s.Package.Datas.Position[2],
				},
				BlockName:   "minecraft:glass",
				BlockStates: map[string]interface{}{},
			},
			4,
		)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 将手中的告示牌放出来
		err = s.Package.API.SetBlock(s.Package.Datas.Position, s.Package.Block.Name, s.Package.Datas.StatesString)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 将告示牌修正到正确的方块上
		err = s.Package.API.RevertStructure(
			uniqueID,
			GlobalAPI.BlockPos{
				s.Package.Datas.Position[0] + 1,
				s.Package.Datas.Position[1],
				s.Package.Datas.Position[2],
			},
		)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		// 恢复备份用结构
	}
	// 放置告示牌
	err := s.Package.API.WritePacket(&packet.BlockActorData{
		Position: s.Package.Datas.Position,
		NBTData: map[string]interface{}{
			"TextOwner":                   s.SignDatas.TextOwner,
			"IgnoreLighting":              s.SignDatas.IgnoreLighting,
			"SignTextColor":               s.SignDatas.SignTextColor,
			"TextIgnoreLegacyBugResolved": s.SignDatas.TextIgnoreLegacyBugResolved,
			"Text":                        s.SignDatas.Text,
		},
	})
	if err != nil {
		return fmt.Errorf("WriteDatas: %v", err)
	}
	// 写入告示牌数据
	return nil
	// 返回值
}
