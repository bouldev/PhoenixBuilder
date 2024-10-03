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

// 描述旗帜的种类
const (
	BannerTypeNormal  int32 = iota // 普通旗帜
	BannerTypeOminous              // 不祥旗帜
)

// 检查当前物品是否可以仅使用命令生成
func (b *Banner) SpecialCheck() (bool, error) {
	// 解码
	err := b.Decode()
	if err != nil {
		return false, fmt.Errorf("SpecialCheck: %v", err)
	}
	b.ItemPackage.AdditionalData.Decoded = true
	// 判断并返回值
	if len(b.BannerData.Patterns) == 0 && b.BannerData.Type == BannerTypeNormal {
		return false, nil
	}
	return true, nil
}

// 从 b.ItemPackage.Item.Basic.ItemTag 提取成书数据，
// 然后保存在 b.BannerData 中
func (b *Banner) Decode() error {
	// 初始化
	var bannerType int32
	var success bool
	patterns := make([]BannerPattern, 0)
	tag := b.ItemPackage.Item.Basic.ItemTag
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
	// Type
	if type_origin, ok := tag["Type"]; ok {
		bannerType, success = type_origin.(int32)
		if !success {
			return fmt.Errorf("Decode: Failed to convert type_origin into int32; tag = %#v", tag)
		}
	}
	// return
	b.BannerData = BannerData{
		Patterns: patterns,
		Type:     bannerType,
	}
	return nil
}

// 制作一个不祥旗帜
func (b *Banner) MakeOminousBanner() error {
	// 初始化
	api := b.ItemPackage.Interface.(*GameInterface.GameInterface)
	// 占用容器资源
	holder := api.Resources.Container.Occupy()
	defer api.Resources.Container.Release(holder)
	// 打开背包
	success, err := api.OpenInventory()
	if err != nil {
		return fmt.Errorf("MakeOminousBanner: %v", err)
	}
	if !success {
		return fmt.Errorf("MakeOminousBanner: Failed to open the inventory when write NBT data to a ominous banner")
	}
	defer api.CloseContainer()
	// 解除可能的物品占用
	err = api.ReplaceItemInInventory(
		GameInterface.TargetMySelf,
		GameInterface.ItemGenerateLocation{
			Path: "slot.hotbar",
			Slot: b.ItemPackage.AdditionalData.HotBarSlot,
		},
		types.ChestSlot{
			Name:   "minecraft:air",
			Count:  1,
			Damage: 0,
		},
		"", true,
	)
	if err != nil {
		return fmt.Errorf("MakeOminousBanner: %v", err)
	}
	// 注册物品堆栈请求至资源管理中心
	requestID := api.Resources.ItemStackOperation.GetNewRequestID()
	err = api.Resources.ItemStackOperation.WriteRequest(
		requestID,
		map[ResourcesControl.ContainerID]ResourcesControl.StackRequestContainerInfo{
			ResourcesControl.ContainerID(GameInterface.ContainerIDInventory): {
				WindowID: 0,
				ChangeResult: map[uint8]protocol.ItemInstance{
					b.ItemPackage.AdditionalData.HotBarSlot: {
						Stack: protocol.ItemStack{
							ItemType:       protocol.ItemType{NetworkID: 576, MetadataValue: 0xf},
							BlockRuntimeID: 0,
							Count:          0x1,
							NBTData:        map[string]any{"Type": int32(1)},
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
		return fmt.Errorf("MakeOminousBanner: %v", err)
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
		Slot:           b.ItemPackage.AdditionalData.HotBarSlot,
		StackNetworkID: 0,
	}
	// 向服务器发送物品堆栈请求
	err = api.WritePacket(&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				RequestID: requestID,
				Actions: []protocol.StackRequestAction{
					&protocol.CraftCreativeStackRequestAction{CreativeItemNetworkID: 0x5bc},
					&takeStackRequestAction,
				},
				FilterStrings: []string{},
				FilterCause:   -1,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("MakeOminousBanner: %v", err)
	}
	// 获取物品堆栈请求结果
	resp, err := api.Resources.ItemStackOperation.LoadResponseAndDelete(requestID)
	if err != nil {
		return fmt.Errorf("MakeOminousBanner: %v", err)
	}
	if resp.Status != protocol.ItemStackResponseStatusOK {
		return fmt.Errorf("MakeOminousBanner: The request was rejected by the remote server when get new banner from the creative inventory; resp = %#v", resp)
	}
	// 返回值
	return nil
}

// 根据 b.BannerData 制作普通旗帜
func (b *Banner) MakeNormalBanner() error {
	// 初始化
	api := b.ItemPackage.Interface.(*GameInterface.GameInterface)
	// 获取旗帜
	newPackage := *b.ItemPackage
	newRequest := DefaultItem{ItemPackage: &newPackage}
	newRequest.ItemPackage.Item.Basic.Count = 1
	newRequest.ItemPackage.AdditionalData.HotBarSlot = 5
	err := newRequest.WriteData()
	if err != nil {
		return fmt.Errorf("MakeNormalBanner: %v", err)
	}
	// 切换物品栏
	err = api.ChangeSelectedHotbarSlot(5)
	if err != nil {
		return fmt.Errorf("MakeNormalBanner: %v", err)
	}
	// 放置织布机
	err = api.SetBlock(b.ItemPackage.AdditionalData.Position, "minecraft:loom", `["direction"=0]`)
	if err != nil {
		return fmt.Errorf("MakeNormalBanner: %v", err)
	}
	// 占用容器资源
	holder := api.Resources.Container.Occupy()
	defer api.Resources.Container.Release(holder)
	// 打开织布机
	success, err := api.OpenContainer(
		b.ItemPackage.AdditionalData.Position,
		"minecraft:loom",
		map[string]any{"direction": int32(0)},
		5,
	)
	if err != nil {
		return fmt.Errorf("MakeNormalBanner: %v", err)
	}
	if !success {
		return fmt.Errorf("MakeNormalBanner: Failed to open the loom block when write NBT data to a banner item")
	}
	defer api.CloseContainer()
	containerOpeningData := api.Resources.Container.GetContainerOpeningData()
	// 修改旗帜
	for _, value := range b.BannerData.Patterns {
		// 初始化
		var bannerMoveResp []protocol.ItemStackResponse
		var dyeMoveResp []protocol.ItemStackResponse
		var patternMoveResp []protocol.ItemStackResponse
		var newerBanner protocol.ItemInstance
		var patternItem protocol.ItemInstance
		// 确定染料和旗帜图案对应的物品名
		dyeName, found := BannerColorToDyeName[uint8(value.Color)]
		if !found {
			continue
		}
		patternName := BannerPatternToItemName[value.Pattern]
		// 获取染料到快捷栏 4
		err = api.ReplaceItemInInventory(
			GameInterface.TargetMySelf,
			GameInterface.ItemGenerateLocation{
				Path: "slot.hotbar",
				Slot: 4,
			},
			types.ChestSlot{
				Name:   dyeName,
				Count:  1,
				Damage: 0,
			},
			"", true,
		)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		// 获取旗帜图案到快捷栏 3
		if len(patternName) != 0 {
			err = api.ReplaceItemInInventory(
				GameInterface.TargetMySelf,
				GameInterface.ItemGenerateLocation{
					Path: "slot.hotbar",
					Slot: 3,
				},
				types.ChestSlot{
					Name:   patternName,
					Count:  1,
					Damage: 0,
				},
				"", true,
			)
			if err != nil {
				return fmt.Errorf("MakeNormalBanner: %v", err)
			}
		}
		// 移动旗帜到织布机
		bannerItem, err := api.Resources.Inventory.GetItemStackInfo(0, 5)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		bannerMoveResp, err = api.MoveItem(
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: GameInterface.ContainerIDInventory,
				Slot:        5,
			},
			GameInterface.ItemLocation{
				WindowID:    containerOpeningData.WindowID,
				ContainerID: GameInterface.ContainerIDLoomInput,
				Slot:        0x9,
			},
			1,
			GameInterface.AirItem,
			bannerItem,
		)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		if bannerMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
			return fmt.Errorf("MakeNormalBanner: The request was rejected by the remote server when try to move banner item to the loom block; bannerMoveResp[0] = %#v", bannerMoveResp[0])
		}
		// 移动染料到织布机
		dyeItem, err := api.Resources.Inventory.GetItemStackInfo(0, 4)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		dyeMoveResp, err = api.MoveItem(
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: GameInterface.ContainerIDInventory,
				Slot:        4,
			},
			GameInterface.ItemLocation{
				WindowID:    containerOpeningData.WindowID,
				ContainerID: GameInterface.ContainerIDLoomDye,
				Slot:        0xa,
			},
			1,
			GameInterface.AirItem,
			dyeItem,
		)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		if dyeMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
			return fmt.Errorf("MakeNormalBanner: The request was rejected by the remote server when try to move dye item to the loom block; dyeMoveResp[0] = %#v", dyeMoveResp[0])
		}
		// 移动旗帜图案到织布机
		if len(patternName) != 0 {
			patternItem, err = api.Resources.Inventory.GetItemStackInfo(0, 3)
			if err != nil {
				return fmt.Errorf("MakeNormalBanner: %v", err)
			}
			patternMoveResp, err = api.MoveItem(
				GameInterface.ItemLocation{
					WindowID:    0,
					ContainerID: GameInterface.ContainerIDInventory,
					Slot:        3,
				},
				GameInterface.ItemLocation{
					WindowID:    containerOpeningData.WindowID,
					ContainerID: GameInterface.ContainerIDLoomMaterial,
					Slot:        0xb,
				},
				1,
				GameInterface.AirItem,
				patternItem,
			)
			if err != nil {
				return fmt.Errorf("MakeNormalBanner: %v", err)
			}
			if patternMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
				return fmt.Errorf("MakeNormalBanner: The request was rejected by the remote server when try to move banner pattern to the loom block; patternMoveResp[0] = %#v", patternMoveResp[0])
			}
		}
		// 设置旗帜的新 NBT 数据
		err = ResourcesControl.DeepCopy(&bannerItem, &newerBanner, func() {
			gob.Register(map[string]interface{}{})
			gob.Register([]interface{}{})
		})
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		newerBanner.Stack.NBTData = b.ItemPackage.Item.Basic.ItemTag
		// 注册物品堆栈请求至资源管理中心
		requestID := api.Resources.ItemStackOperation.GetNewRequestID()
		err = api.Resources.ItemStackOperation.WriteRequest(
			requestID,
			map[ResourcesControl.ContainerID]ResourcesControl.StackRequestContainerInfo{
				ResourcesControl.ContainerID(GameInterface.ContainerIDLoomInput): {
					WindowID: uint32(containerOpeningData.WindowID),
					ChangeResult: map[uint8]protocol.ItemInstance{
						0x9: GameInterface.AirItem,
					},
				},
				ResourcesControl.ContainerID(GameInterface.ContainerIDLoomDye): {
					WindowID: uint32(containerOpeningData.WindowID),
					ChangeResult: map[uint8]protocol.ItemInstance{
						0xa: GameInterface.AirItem,
					},
				},
				ResourcesControl.ContainerID(GameInterface.ContainerIDInventory): {
					WindowID: 0,
					ChangeResult: map[uint8]protocol.ItemInstance{
						5: newerBanner,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
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
						&protocol.CraftLoomRecipeStackRequestAction{Pattern: value.Pattern},
						&protocol.ConsumeStackRequestAction{
							DestroyStackRequestAction: protocol.DestroyStackRequestAction{
								Count: 1,
								Source: protocol.StackRequestSlotInfo{
									ContainerID:    GameInterface.ContainerIDLoomInput,
									Slot:           0x9,
									StackNetworkID: bannerMoveResp[0].ContainerInfo[1].SlotInfo[0].StackNetworkID,
								},
							},
						},
						&protocol.ConsumeStackRequestAction{
							DestroyStackRequestAction: protocol.DestroyStackRequestAction{
								Count: 1,
								Source: protocol.StackRequestSlotInfo{
									ContainerID:    GameInterface.ContainerIDLoomDye,
									Slot:           0xa,
									StackNetworkID: dyeMoveResp[0].ContainerInfo[1].SlotInfo[0].StackNetworkID,
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
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		// 获取物品堆栈请求结果
		resp, err := api.Resources.ItemStackOperation.LoadResponseAndDelete(requestID)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		if resp.Status != protocol.ItemStackResponseStatusOK {
			return fmt.Errorf("MakeNormalBanner: The request was rejected by the remote server when get new banner from the loom block; resp = %#v", resp)
		}
		// 将旗帜图案从织布机中拿回
		if len(patternName) != 0 {
			patternMoveResp, err = api.MoveItem(
				GameInterface.ItemLocation{
					WindowID:    containerOpeningData.WindowID,
					ContainerID: GameInterface.ContainerIDLoomMaterial,
					Slot:        0xb,
				},
				GameInterface.ItemLocation{
					WindowID:    0,
					ContainerID: GameInterface.ContainerIDInventory,
					Slot:        3,
				},
				1,
				GameInterface.AirItem,
				patternItem,
			)
			if err != nil {
				return fmt.Errorf("MakeNormalBanner: %v", err)
			}
			if patternMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
				return fmt.Errorf("MakeNormalBanner: The request was rejected by the remote server when try to move banner pattern back; patternMoveResp[0] = %#v", patternMoveResp[0])
			}
		}
	}
	// 将新旗帜修正到正确的位置
	if b.ItemPackage.AdditionalData.HotBarSlot != 5 {
		// 解除可能的物品占用
		err = api.ReplaceItemInInventory(
			GameInterface.TargetMySelf,
			GameInterface.ItemGenerateLocation{
				Path: "slot.hotbar",
				Slot: b.ItemPackage.AdditionalData.HotBarSlot,
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
		// 获取新旗帜的物品信息
		bannerItem, err := api.Resources.Inventory.GetItemStackInfo(0, 5)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		// 修正位置
		bannerMoveResp, err := api.MoveItem(
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: GameInterface.ContainerIDInventory,
				Slot:        5,
			},
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: GameInterface.ContainerIDInventory,
				Slot:        b.ItemPackage.AdditionalData.HotBarSlot,
			},
			1,
			GameInterface.AirItem,
			bannerItem,
		)
		if err != nil {
			return fmt.Errorf("MakeNormalBanner: %v", err)
		}
		if bannerMoveResp[0].Status != protocol.ItemStackResponseStatusOK {
			return fmt.Errorf("MakeNormalBanner: The request was rejected by the remote server when try to correct the item place; bannerMoveResp[0] = %#v", bannerMoveResp[0])
		}
	}
	// 返回值
	return nil
}

// ...
func (b *Banner) WriteData() error {
	// 初始化
	var err error
	api := b.ItemPackage.Interface.(*GameInterface.GameInterface)
	// 制作单个旗帜
	switch b.BannerData.Type {
	case BannerTypeNormal:
		err = b.MakeNormalBanner()
	case BannerTypeOminous:
		err = b.MakeOminousBanner()
	}
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 堆叠旗帜的处理
	if b.ItemPackage.Item.Basic.Count > 1 {
		err = api.CopyItem(
			b.ItemPackage.AdditionalData.HotBarSlot,
			b.ItemPackage.AdditionalData.Position,
			b.ItemPackage.Item.Basic.Count,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 返回值
	return nil
}