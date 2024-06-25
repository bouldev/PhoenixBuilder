package NBTAssigner

import (
	"fmt"
	"phoenixbuilder/fastbuilder/generics"
	GameInterface "phoenixbuilder/game_control/game_interface"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/go-gl/mathgl/mgl32"
)

// ...
const PlaceSignPath = "phoenixbuilder/bdump/nbt_assigner/place_sign.go"

/*
从 s.BlockEntity.Block.NBT 提取告示牌的
一部分数据并保存在 s.SignData 或
s.LegacySignData 中。

如果 s.IsNotLegacySignBlock 为真，
则 s.SignData 将存放这些数据，
否则由 s.LegacySignData 存放这些数据。

对于未被解码的部分，
我们不再检查这部分 NBT 是否正确，
我们信任并且永远认为它们是正确且完整的
*/
func (s *Sign) Decode() error {
	nbt := s.BlockEntity.Block.NBT
	_, s.IsNotLegacySignBlock = nbt["IsWaxed"]
	// 初始化
	id, err := generics.To[string](nbt["id"], `nbt["id"]`, PlaceSignPath)
	if err != nil {
		return fmt.Errorf("Decode: %v", err)
	}
	if id == "HangingSign" {
		s.IsHangingSignBlock = true
	}
	// 确定告示牌类型
	if !s.IsNotLegacySignBlock {
		s.LegacySignData = &LegacySignData{}
		// 初始化
		ignoreLighting, err := generics.To[byte](nbt["IgnoreLighting"], `nbt["IgnoreLighting"]`, PlaceSignPath)
		if err != nil {
			return fmt.Errorf("Decode: %v", err)
		}
		if ignoreLighting == byte(1) {
			s.LegacySignData.IgnoreLighting = true
		}
		// IgnoreLighting
		s.LegacySignData.SignTextColor, err = generics.To[int32](nbt["SignTextColor"], `nbt["SignTextColor"]`, PlaceSignPath)
		if err != nil {
			return fmt.Errorf("Decode: %v", err)
		}
		// SignTextColor
	} else {
		s.SignData = &SignData{}
		// 初始化
		isWaxed, err := generics.To[byte](nbt["IsWaxed"], `nbt["IsWaxed"]`, PlaceSignPath)
		if err != nil {
			return fmt.Errorf("Decode: %v", err)
		}
		if isWaxed == byte(1) {
			s.SignData.IsWaxed = true
		}
		// IsWaxed
		{
			text, err := generics.To[map[string]any](nbt["FrontText"], `nbt["FrontText"]`, PlaceSignPath)
			if err != nil {
				return fmt.Errorf("Decode: %v", err)
			}
			// FrontText
			ignoreLighting, err := generics.To[byte](text["IgnoreLighting"], `nbt["FrontText"]["IgnoreLighting"]`, PlaceSignPath)
			if err != nil {
				return fmt.Errorf("Decode: %v", err)
			}
			if ignoreLighting == byte(1) {
				s.SignData.FrontText.IgnoreLighting = true
			}
			// FrontText["IgnoreLighting"]
			s.SignData.FrontText.SignTextColor, err = generics.To[int32](text["SignTextColor"], `nbt["FrontText"]["SignTextColor"]`, PlaceSignPath)
			if err != nil {
				return fmt.Errorf("Decode: %v", err)
			}
			// FrontText["SignTextColor"]
		}
		// FrontText
		{
			text, err := generics.To[map[string]any](nbt["BackText"], `nbt["BackText"]`, PlaceSignPath)
			if err != nil {
				return fmt.Errorf("Decode: %v", err)
			}
			// BackText
			ignoreLighting, err := generics.To[byte](text["IgnoreLighting"], `nbt["BackText"]["IgnoreLighting"]`, PlaceSignPath)
			if err != nil {
				return fmt.Errorf("Decode: %v", err)
			}
			if ignoreLighting == byte(1) {
				s.SignData.BackText.IgnoreLighting = true
			}
			// BackText["IgnoreLighting"]
			s.SignData.BackText.SignTextColor, err = generics.To[int32](text["SignTextColor"], `nbt["BackText"]["SignTextColor"]`, PlaceSignPath)
			if err != nil {
				return fmt.Errorf("Decode: %v", err)
			}
			// BackText["SignTextColor"]
		}
		// BackText
	}
	// decode data
	return nil
	// return
}

// 放置一个告示牌并写入告示牌数据
func (s *Sign) WriteData() error {
	var preBlockName string = "oak_hanging_sign"
	api := s.BlockEntity.Interface.(*GameInterface.GameInterface)
	useItemOnBlocks := GameInterface.UseItemOnBlocks{
		HotbarSlotID: 0,
		BlockPos:     s.BlockEntity.AdditionalData.Position,
		BlockStates:  map[string]interface{}{"facing_direction": int32(4)},
	}
	// 初始化
	if s.BlockEntity.AdditionalData.FastMode {
		err := s.BlockEntity.Interface.SetBlockAsync(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		return nil
	}
	// 放置告示牌(快速导入模式下)
	{
		err := api.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", s.BlockEntity.AdditionalData.Position[0], s.BlockEntity.AdditionalData.Position[1], s.BlockEntity.AdditionalData.Position[2]), true)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 传送机器人到告示牌所在的位置
		err = api.SetBlock(s.BlockEntity.AdditionalData.Position, "air", `[]`)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 清除告示牌处的方块
		if !s.IsHangingSignBlock {
			preBlockName = "wall_sign"
		}
		useItemOnBlocks.BlockName = preBlockName
		// 确定 预设告示牌 的方块名称
		err = api.SetBlock(s.BlockEntity.AdditionalData.Position, preBlockName, `["facing_direction"=4]`)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 放置预设告示牌方块
	}
	// 放置告示牌
	err := api.ChangeSelectedHotbarSlot(0)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 切换手持物品栏到快捷栏 0
	resp := api.SendWSCommandWithResponse(
		"replaceitem entity @s slot.hotbar 0 air",
		ResourcesControl.CommandRequestOptions{
			TimeOut: ResourcesControl.CommandRequestNoDeadLine,
		},
	)
	if resp.Error != nil {
		return fmt.Errorf("WriteData: %v", resp.Error)
	}
	// 清空快捷栏 0 以防止稍后在手持蜜脾的情况下点击告示牌，
	// 因为用 蜜脾 点击告示牌会导致告示牌被封装
	err = api.ClickBlock(GameInterface.UseItemOnBlocks{
		HotbarSlotID: 0,
		BlockPos:     s.BlockEntity.AdditionalData.Position,
		BlockName:    preBlockName,
		BlockStates:  map[string]interface{}{"facing_direction": int32(4)},
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 打开告示牌
	signBlockNBTData := s.BlockEntity.Block.NBT
	if !s.IsNotLegacySignBlock {
		signBlockNBTData = map[string]any{"FrontText": s.BlockEntity.Block.NBT}
	}
	err = api.WritePacket(&packet.BlockActorData{
		Position: s.BlockEntity.AdditionalData.Position,
		NBTData:  signBlockNBTData,
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 写入告示牌数据
	{
		var bestFrontColor [3]uint8
		var bestBackColor *[3]uint8
		playerPosition := mgl32.Vec3{
			float32(s.BlockEntity.AdditionalData.Position[0]),
			float32(s.BlockEntity.AdditionalData.Position[1]),
			float32(s.BlockEntity.AdditionalData.Position[2]),
		}
		// 初始化
		if s.IsNotLegacySignBlock {
			frontRGB, _ := DecodeVarRGBA(s.SignData.FrontText.SignTextColor)
			backRGB, _ := DecodeVarRGBA(s.SignData.BackText.SignTextColor)
			bestFrontColor = SearchForBestColor(frontRGB, DefaultDyeColor)
			bestBackColorTemp := SearchForBestColor(backRGB, DefaultDyeColor)
			bestBackColor = &bestBackColorTemp
		} else {
			rgb, _ := DecodeVarRGBA(s.LegacySignData.SignTextColor)
			bestFrontColor = SearchForBestColor(rgb, DefaultDyeColor)
		}
		// 确定告示牌各面的颜色
		if bestFrontColor != [3]uint8{0, 0, 0} {
			dyeItemName := RGBToDyeItemName[bestFrontColor]
			// 确定染料的物品名
			resp = api.SendWSCommandWithResponse(
				fmt.Sprintf("replaceitem entity @s slot.hotbar 0 %s", dyeItemName),
				ResourcesControl.CommandRequestOptions{
					TimeOut: ResourcesControl.CommandRequestNoDeadLine,
				},
			)
			if resp.Error != nil {
				return fmt.Errorf("WriteData: %v", resp.Error)
			}
			// 获取对应的染料到快捷栏 0
			err = api.ClickBlockWitchPlayerPosition(useItemOnBlocks, playerPosition)
			if err != nil {
				return fmt.Errorf("WriteData: %v", err)
			}
			// 告示牌正面染色
		}
		if bestBackColor != nil && *bestBackColor != [3]uint8{0, 0, 0} {
			dyeItemName := RGBToDyeItemName[*bestBackColor]
			// 确定染料的物品名
			resp = api.SendWSCommandWithResponse(
				fmt.Sprintf("replaceitem entity @s slot.hotbar 0 %s", dyeItemName),
				ResourcesControl.CommandRequestOptions{
					TimeOut: ResourcesControl.CommandRequestNoDeadLine,
				},
			)
			if resp.Error != nil {
				return fmt.Errorf("WriteData: %v", resp.Error)
			}
			// 获取对应的染料到快捷栏 0
			playerPosition[0] = playerPosition[0] + 1
			err = api.ClickBlockWitchPlayerPosition(useItemOnBlocks, playerPosition)
			if err != nil {
				return fmt.Errorf("WriteData: %v", err)
			}
			// 告示牌背面染色
		}
	}
	// 告示牌染色
	{
		playerPosition := mgl32.Vec3{
			float32(s.BlockEntity.AdditionalData.Position[0]),
			float32(s.BlockEntity.AdditionalData.Position[1]),
			float32(s.BlockEntity.AdditionalData.Position[2]),
		}
		// 初始化
		{
			matchA := s.IsNotLegacySignBlock && (s.SignData.FrontText.IgnoreLighting || s.SignData.BackText.IgnoreLighting)
			matchB := !s.IsNotLegacySignBlock && s.LegacySignData.IgnoreLighting
			if matchA || matchB {
				resp = api.SendWSCommandWithResponse(
					"replaceitem entity @s slot.hotbar 0 glow_ink_sac",
					ResourcesControl.CommandRequestOptions{
						TimeOut: ResourcesControl.CommandRequestNoDeadLine,
					},
				)
				if resp.Error != nil {
					return fmt.Errorf("WriteData: %v", resp.Error)
				}
				// 获取一个 发光墨囊 到快捷栏 0
			}
			// 取得 发光墨囊
			if (s.IsNotLegacySignBlock && s.SignData.FrontText.IgnoreLighting) || matchB {
				err = api.ClickBlockWitchPlayerPosition(useItemOnBlocks, playerPosition)
				if err != nil {
					return fmt.Errorf("WriteData: %v", err)
				}
			}
			if s.IsNotLegacySignBlock && s.SignData.BackText.IgnoreLighting {
				playerPosition[0] = playerPosition[0] + 1
				err = api.ClickBlockWitchPlayerPosition(useItemOnBlocks, playerPosition)
				if err != nil {
					return fmt.Errorf("WriteData: %v", err)
				}
			}
			// 使用 发光墨囊 点击告示牌的对应面以让该面发光
		}
		// 附加发光效果
	}
	// 告示牌发光效果
	if s.IsNotLegacySignBlock && s.SignData.IsWaxed {
		resp = api.SendWSCommandWithResponse(
			"replaceitem entity @s slot.hotbar 0 honeycomb",
			ResourcesControl.CommandRequestOptions{
				TimeOut: ResourcesControl.CommandRequestNoDeadLine,
			},
		)
		if resp.Error != nil {
			return fmt.Errorf("WriteData: %v", resp.Error)
		}
		// 获取一个 蜜脾 到快捷栏 0
		err = api.ClickBlock(useItemOnBlocks)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 封装告示牌
	}
	// 告示牌涂蜡
	err = api.SetBlockAsync(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 覆写告示牌的种类和朝向为正确值
	return nil
	// 返回值
}
