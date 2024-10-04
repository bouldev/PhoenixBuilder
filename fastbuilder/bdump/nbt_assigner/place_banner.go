package NBTAssigner

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"strings"
)

// 从 b.BlockEntity.Block.NBT 提取告示牌的
// 一部分数据并保存在 b.BannerBlockData 中
func (b *BannerBlock) Decode() error {
	// 初始化
	var bannerBase int32
	var bannerType int32
	var success bool
	nbt := b.BlockEntity.Block.NBT
	patterns := make([]BannerPattern, 0)
	// Patterns
	if patterns_origin, ok := nbt["Patterns"]; ok {
		patterns_got, success := patterns_origin.([]any)
		if !success {
			return fmt.Errorf("Decode: Failed to convert patterns_origin into []any; nbt = %#v", nbt)
		}
		for key, value := range patterns_got {
			current, success := value.(map[string]any)
			if !success {
				return fmt.Errorf("Decode: Failed to convert patterns_got[%d] into map[string]any; nbt = %#v", key, nbt)
			}
			color, ok := current["Color"].(int32)
			if !ok {
				return fmt.Errorf(`Decode: Failed to convert patterns_got[%d]["Color"] into int32; nbt = %#v`, key, nbt)
			}
			pattern_string, ok := current["Pattern"].(string)
			if !ok {
				return fmt.Errorf(`Decode: Failed to convert patterns_got[%d]["Pattern"] into string; nbt = %#v`, key, nbt)
			}
			patterns = append(patterns, BannerPattern{
				Color:   color,
				Pattern: pattern_string,
			})
		}
	}
	// Base
	if base_origin, found := nbt["Base"]; found {
		bannerBase, success = base_origin.(int32)
		if !success {
			return fmt.Errorf("Decode: Failed to convert base_origin into int32; nbt = %#v", nbt)
		}
	}
	// Type
	if type_origin, ok := nbt["Type"]; ok {
		bannerType, success = type_origin.(int32)
		if !success {
			return fmt.Errorf("Decode: Failed to convert type_origin into int32; nbt = %#v", nbt)
		}
	}
	// 返回值
	b.BannerBlockData = BannerBlockData{
		Base:     bannerBase,
		Patterns: patterns,
		Type:     bannerType,
	}
	return nil
}

// 制作复杂旗帜并放置
func (b *BannerBlock) WriteData() error {
	// 初始化
	bannerFacing := uint8(1)
	api := b.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 如果当前需要快速导入，
	// 或当前旗帜为纯黑
	if b.BlockEntity.AdditionalData.FastMode || (b.BannerBlockData.Base == 0 && len(b.BannerBlockData.Patterns) == 0 && b.BannerBlockData.Type != BannerTypeOminous) {
		err := b.BlockEntity.Interface.SetBlockAsync(b.BlockEntity.AdditionalData.Position, b.BlockEntity.Block.Name, b.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		return nil
	}
	// 传送机器人到旗帜处
	err := api.SendSettingsCommand(
		fmt.Sprintf(
			"tp %d %d %d",
			b.BlockEntity.AdditionalData.Position[0],
			b.BlockEntity.AdditionalData.Position[1],
			b.BlockEntity.AdditionalData.Position[2],
		),
		true,
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 清理旗帜处的方块
	err = b.BlockEntity.Interface.SetBlock(b.BlockEntity.AdditionalData.Position, "minecraft:air", "[]")
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 构造用于获取旗帜的子请求
	newRequest := GetGenerateItemMethod(
		&ItemPackage{
			Interface: api,
			Item: GeneralItem{
				Basic: ItemBasicData{
					Name:     "banner",
					Count:    1,
					MetaData: uint16(b.BannerBlockData.Base),
					ItemTag:  b.BlockEntity.Block.NBT,
				},
			},
			AdditionalData: ItemAdditionalData{
				Decoded:    false,
				HotBarSlot: 5,
				Position:   b.BlockEntity.AdditionalData.Position,
				Type:       "Banner",
				Settings:   b.BlockEntity.AdditionalData.Settings,
				FastMode:   b.BlockEntity.AdditionalData.FastMode,
				Others:     b.BlockEntity.AdditionalData.Others,
			},
		},
	)
	// 获取旗帜到快捷栏 5
	needSpecialTreatment, err := newRequest.SpecialCheck()
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	if needSpecialTreatment {
		err = newRequest.WriteData()
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	} else {
		err = api.ReplaceItemInInventory(
			GameInterface.TargetMySelf,
			GameInterface.ItemGenerateLocation{
				Path: "slot.hotbar",
				Slot: 5,
			},
			types.ChestSlot{
				Name:   "banner",
				Count:  1,
				Damage: uint16(b.BannerBlockData.Base),
			},
			"", true,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 放置旗帜
	if strings.Contains(b.BlockEntity.Block.Name, "standing_banner") {
		bannerFacing = 1
	} else {
		if facing_direction, found := b.BlockEntity.Block.States["facing_direction"]; found {
			if facing, ok := facing_direction.(int32); ok {
				bannerFacing = uint8(facing)
			}
		}
	}
	err = api.PlaceBlockWithFacing(b.BlockEntity.AdditionalData.Position, 5, bannerFacing)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 等待更改
	err = api.AwaitChangesGeneral()
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 修正旗帜的朝向
	err = b.BlockEntity.Interface.SetBlockAsync(
		b.BlockEntity.AdditionalData.Position,
		b.BlockEntity.Block.Name,
		b.BlockEntity.AdditionalData.BlockStates,
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 返回值
	return nil
}
