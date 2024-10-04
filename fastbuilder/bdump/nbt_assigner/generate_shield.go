package NBTAssigner

import (
	"encoding/gob"
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	GameInterface "phoenixbuilder/game_control/game_interface"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 检查当前物品是否应该通过复杂的步骤制作
func (s *Shield) SpecialCheck() (bool, error) {
	// 解码
	err := s.Decode()
	if err != nil {
		return false, fmt.Errorf("SpecialCheck: %v", err)
	}
	s.ItemPackage.AdditionalData.Decoded = true
	// 判断并返回值
	if s.ShieldData == nil {
		return false, nil
	}
	return true, nil
}

// 从 s.ItemPackage.Item.Basic.ItemTag 提取成书数据，
// 然后保存在 s.ShieldData 中
func (s *Shield) Decode() error {
	// 初始化
	var base int32
	var hasBase bool
	var success bool
	patterns := make([]BannerPattern, 0)
	tag := s.ItemPackage.Item.Basic.ItemTag
	// Patterns
	if patterns_origin, ok := tag["Patterns"]; ok {
		patterns_got, success := patterns_origin.([]any)
		if !success {
			return fmt.Errorf("Decode: Failed to convert patterns_origin into []any; tag = %#v", tag)
		}
		for key, value := range patterns_got {
			current, success := value.(map[string]any)
			if !success {
				return fmt.Errorf("Decode: Failed to convert patterns_got[%d] into map[string]any; tag = %#v", key, tag)
			}
			color, ok := current["Color"].(int32)
			if !ok {
				return fmt.Errorf(`Decode: Failed to convert patterns_got[%d]["Color"] into int32; tag = %#v`, key, tag)
			}
			pattern_string, ok := current["Pattern"].(string)
			if !ok {
				return fmt.Errorf(`Decode: Failed to convert patterns_got[%d]["Pattern"] into string; tag = %#v`, key, tag)
			}
			patterns = append(patterns, BannerPattern{
				Color:   color,
				Pattern: pattern_string,
			})
		}
	}
	// Base
	if base_origin, found := tag["Base"]; found {
		base, success = base_origin.(int32)
		if !success {
			return fmt.Errorf("Decode: Failed to convert base_origin into int32; tag = %#v", tag)
		}
		hasBase = true
	}
	// return
	if hasBase || len(patterns) != 0 {
		s.ShieldData = &ShieldData{
			Patterns: patterns,
			Base:     base,
		}
	}
	return nil
}

// ...
func (s *Shield) WriteData() error {
	// 初始化
	var itemTag map[string]any
	var shieldBase uint16
	api := s.ItemPackage.Interface.(*GameInterface.GameInterface)
	// 如果当前是快速模式
	if s.ItemPackage.AdditionalData.FastMode {
		err := api.ReplaceItemInInventory(
			GameInterface.TargetMySelf,
			GameInterface.ItemGenerateLocation{
				Path: "slot.hotbar",
				Slot: s.ItemPackage.AdditionalData.HotBarSlot,
			},
			types.ChestSlot{
				Name:   s.ItemPackage.Item.Basic.Name,
				Count:  s.ItemPackage.Item.Basic.Count,
				Damage: s.ItemPackage.Item.Basic.MetaData,
			},
			"", false,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		return nil
	}
	// 深拷贝盾牌 NBT 数据
	err := ResourcesControl.DeepCopy(
		&s.ItemPackage.Item.Basic.ItemTag,
		&itemTag,
		func() {
			gob.Register(map[string]interface{}{})
			gob.Register([]interface{}{})
		},
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 构造用于获取旗帜的子请求
	if s.ShieldData != nil {
		for _, value := range s.ShieldData.Patterns {
			if value.Pattern == "ill" {
				itemTag = map[string]any{"Type": BannerTypeOminous}
				break
			}
		}
		shieldBase = uint16(s.ShieldData.Base)
	}
	newRequest := GetGenerateItemMethod(
		&ItemPackage{
			Interface: api,
			Item: GeneralItem{
				Basic: ItemBasicData{
					Name:     "banner",
					Count:    1,
					MetaData: shieldBase,
					ItemTag:  itemTag,
				},
			},
			AdditionalData: ItemAdditionalData{
				Decoded:    false,
				HotBarSlot: 5,
				Position:   s.ItemPackage.AdditionalData.Position,
				Type:       "Banner",
				Settings:   s.ItemPackage.AdditionalData.Settings,
				FastMode:   s.ItemPackage.AdditionalData.FastMode,
				Others:     s.ItemPackage.AdditionalData.Others,
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
				Damage: shieldBase,
			},
			"", true,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 获取盾牌到快捷栏 4
	err = api.ReplaceItemInInventory(
		GameInterface.TargetMySelf,
		GameInterface.ItemGenerateLocation{
			Path: "slot.hotbar",
			Slot: 4,
		},
		types.ChestSlot{
			Name:   "shield",
			Count:  1,
			Damage: s.ItemPackage.Item.Basic.MetaData,
		},
		"", true,
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 切换物品栏
	err = api.ChangeSelectedHotbarSlot(5)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 放置工作台
	err = api.SetBlock(s.ItemPackage.AdditionalData.Position, "minecraft:crafting_table", "[]")
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 占用容器资源
	holder := api.Resources.Container.Occupy()
	defer api.Resources.Container.Release(holder)
	// 打开工作台
	success, err := api.OpenContainer(
		s.ItemPackage.AdditionalData.Position,
		"minecraft:crafting_table",
		map[string]any{},
		5,
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	if !success {
		return fmt.Errorf("WriteData: Failed to open the crafting block when write NBT data to a shield item")
	}
	defer api.CloseContainer()
	containerOpeningData := api.Resources.Container.GetContainerOpeningData()
	// 移动旗帜到工作台
	bannerItem, err := api.Resources.Inventory.GetItemStackInfo(0, 5)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	bannerMoveResp, err := api.MoveItem(
		GameInterface.ItemLocation{
			WindowID:    0,
			ContainerID: GameInterface.ContainerIDInventory,
			Slot:        5,
		},
		GameInterface.ItemLocation{
			WindowID:    containerOpeningData.WindowID,
			ContainerID: GameInterface.ContainerIDCraftingInput,
			Slot:        0x20,
		},
		1,
		GameInterface.AirItem,
		bannerItem,
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	if bannerMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
		return fmt.Errorf("WriteData: The request was rejected by the remote server when try to move banner item to the crafting block; bannerMoveResp[0] = %#v", bannerMoveResp[0])
	}
	// 移动盾牌到工作台
	shieldItem, err := api.Resources.Inventory.GetItemStackInfo(0, 4)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	shieldMoveResp, err := api.MoveItem(
		GameInterface.ItemLocation{
			WindowID:    0,
			ContainerID: GameInterface.ContainerIDInventory,
			Slot:        4,
		},
		GameInterface.ItemLocation{
			WindowID:    containerOpeningData.WindowID,
			ContainerID: GameInterface.ContainerIDCraftingInput,
			Slot:        0x21,
		},
		1,
		GameInterface.AirItem,
		shieldItem,
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	if shieldMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
		return fmt.Errorf("WriteData: The request was rejected by the remote server when try to move shield item to the crafting block; shieldMoveResp[0] = %#v", shieldMoveResp[0])
	}
	// 解除可能的物品占用
	err = api.ReplaceItemInInventory(
		GameInterface.TargetMySelf,
		GameInterface.ItemGenerateLocation{
			Path: "slot.hotbar",
			Slot: 5,
		},
		types.ChestSlot{
			Name:   "minecraft:air",
			Count:  1,
			Damage: 0,
		},
		"", true,
	)
	if err != nil {
		return fmt.Errorf("MakeNormalBanner: %v", err)
	}
	// 注册物品堆栈请求至资源管理中心
	requestID := api.Resources.ItemStackOperation.GetNewRequestID()
	err = api.Resources.ItemStackOperation.WriteRequest(
		requestID,
		map[ResourcesControl.ContainerID]ResourcesControl.StackRequestContainerInfo{
			ResourcesControl.ContainerID(GameInterface.ContainerIDCraftingInput): {
				WindowID: uint32(containerOpeningData.WindowID),
				ChangeResult: map[uint8]protocol.ItemInstance{
					0x20: GameInterface.AirItem,
					0x21: GameInterface.AirItem,
				},
			},
			ResourcesControl.ContainerID(GameInterface.ContainerIDInventory): {
				WindowID: 0,
				ChangeResult: map[uint8]protocol.ItemInstance{
					5: {
						Stack: protocol.ItemStack{
							ItemType:       protocol.ItemType{NetworkID: 358, MetadataValue: 0x0},
							BlockRuntimeID: 0,
							Count:          0x1,
							NBTData:        s.ItemPackage.Item.Basic.ItemTag,
							CanBePlacedOn:  []string{},
							CanBreak:       []string{},
							HasNetworkID:   false,
						},
					},
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 构造 takeStackRequestAction
	takeStackRequestAction := protocol.TakeStackRequestAction{}
	takeStackRequestAction.Count = 1
	takeStackRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    protocol.ContainerCreatedOutput, // [NEMC 1.20.10] 60 -> 61 (Added by Happy2018new)
		Slot:           0x32,
		StackNetworkID: requestID,
	}
	takeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    GameInterface.ContainerIDInventory,
		Slot:           5,
		StackNetworkID: 0,
	}
	// 向服务器发送物品堆栈请求
	err = api.WritePacket(&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				RequestID: requestID,
				Actions: []protocol.StackRequestAction{
					&protocol.CraftRecipeStackRequestAction{RecipeNetworkID: 0x8f9},
					&protocol.ConsumeStackRequestAction{
						DestroyStackRequestAction: protocol.DestroyStackRequestAction{
							Count: 1,
							Source: protocol.StackRequestSlotInfo{
								ContainerID:    GameInterface.ContainerIDCraftingInput,
								Slot:           0x20,
								StackNetworkID: bannerMoveResp[0].ContainerInfo[1].SlotInfo[0].StackNetworkID,
							},
						},
					},
					&protocol.ConsumeStackRequestAction{
						DestroyStackRequestAction: protocol.DestroyStackRequestAction{
							Count: 1,
							Source: protocol.StackRequestSlotInfo{
								ContainerID:    GameInterface.ContainerIDCraftingInput,
								Slot:           0x21,
								StackNetworkID: shieldMoveResp[0].ContainerInfo[1].SlotInfo[0].StackNetworkID,
							},
						},
					},
					&takeStackRequestAction,
				},
				FilterStrings: []string{},
				FilterCause:   -1,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 获取物品堆栈请求结果
	resp, err := api.Resources.ItemStackOperation.LoadResponseAndDelete(requestID)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	if resp.Status != protocol.ItemStackResponseStatusOK {
		return fmt.Errorf("WriteData: The request was rejected by the remote server when get new shield from the crafting block; resp = %#v", resp)
	}
	// 将新盾牌修正到正确的位置
	if s.ItemPackage.AdditionalData.HotBarSlot != 5 {
		// 解除可能的物品占用
		err = api.ReplaceItemInInventory(
			GameInterface.TargetMySelf,
			GameInterface.ItemGenerateLocation{
				Path: "slot.hotbar",
				Slot: s.ItemPackage.AdditionalData.HotBarSlot,
			},
			types.ChestSlot{
				Name:   "minecraft:air",
				Count:  1,
				Damage: 0,
			},
			"", true,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 获取新盾牌的物品信息
		shieldItem, err := api.Resources.Inventory.GetItemStackInfo(0, 5)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 修正位置
		shieldMoveResp, err = api.MoveItem(
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: GameInterface.ContainerIDInventory,
				Slot:        5,
			},
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: GameInterface.ContainerIDInventory,
				Slot:        s.ItemPackage.AdditionalData.HotBarSlot,
			},
			1,
			GameInterface.AirItem,
			shieldItem,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		if shieldMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
			return fmt.Errorf("WriteData: The request was rejected by the remote server when try to correct the item place; shieldMoveResp[0] = %#v", shieldMoveResp[0])
		}
	}
	// 返回值
	return nil
}
