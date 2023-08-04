package NBTAssigner

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft/protocol"
)

// 以 replaceitem 的形式获取 i.Item 所指代的物品到物品栏
func (i *ItemPackage) ReplaceItemInInventory() error {
	var itemComponents *ItemComponents
	item := i.Item
	api := i.Interface.(*GameInterface.GameInterface)
	// 初始化
	if i.Item.Enhancement != nil {
		itemComponents = i.Item.Enhancement.ItemComponents
	}
	err := api.ReplaceItemInInventory(
		GameInterface.TargetMySelf,
		GameInterface.ItemGenerateLocation{
			Path: "slot.hotbar",
			Slot: i.AdditionalData.HotBarSlot,
		},
		types.ChestSlot{
			Name:   item.Basic.Name,
			Count:  item.Basic.Count,
			Damage: item.Basic.MetaData,
		},
		MarshalItemComponents(itemComponents),
		true,
	)
	if err != nil {
		return fmt.Errorf("ReplaceItemInInventory: %v", err)
	}
	// 获取物品到物品栏，并附加物品组件数据
	return nil
	// 返回值
}

// 将 i.Item.Enhancement.Enchantments 所指代
// 的附魔属性以 enchant 的形式添加到物品上
func (i *ItemPackage) AddEnchantments() error {
	item := i.Item
	api := i.Interface.(*GameInterface.GameInterface)
	// 初始化
	for _, value := range *item.Enhancement.Enchantments {
		err := api.SendSettingsCommand(
			fmt.Sprintf(
				"enchant @s %d %d",
				value.ID,
				value.Level,
			),
			true,
		)
		if err != nil {
			return fmt.Errorf("AddEnchantments: %v", err)
		}
	}
	// 发送附魔命令
	err := api.AwaitChangesGeneral()
	if err != nil {
		return fmt.Errorf("AddEnchantments: %v", err)
	}
	// 等待更改
	return nil
	// 返回值
}

// 将 i.Item.Enhancement.DisplayName 所指代
// 的自定义物品名称以铁砧改名的形式添加到物品上
func (i *ItemPackage) AddDisplayName() error {
	item := i.Item
	api := i.Interface.(*GameInterface.GameInterface)
	// 初始化
	resp, err := api.RenameItemByAnvil(
		i.AdditionalData.Position,
		`["direction": 0, "damage": "undamaged"]`,
		5,
		[]GameInterface.ItemRenamingRequest{
			{
				Slot: i.AdditionalData.HotBarSlot,
				Name: item.Enhancement.DisplayName,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("AddDisplayName: %v", err)
	}
	if resp[0].Destination == nil {
		return fmt.Errorf("AddDisplayName: Inventory was full")
	}
	// 利用铁砧修改物品名称
	if resp[0].Destination.Slot != i.AdditionalData.HotBarSlot {
		itemData, err := api.Resources.Inventory.GetItemStackInfo(0, resp[0].Destination.Slot)
		if err != nil {
			return fmt.Errorf("AddDisplayName: %v", err)
		}
		// 获取已被铁砧操作后的物品数据
		err = api.ReplaceItemInInventory(
			GameInterface.TargetMySelf,
			GameInterface.ItemGenerateLocation{
				Path: "slot.hotbar",
				Slot: i.AdditionalData.HotBarSlot,
			},
			types.ChestSlot{
				Name:   "air",
				Count:  1,
				Damage: 0,
			},
			"",
			true,
		)
		if err != nil {
			return fmt.Errorf("AddDisplayName: %v", err)
		}
		// 将原有物品栏替换为空气以解除它的占用态
		res, err := api.MoveItem(
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: 0xc,
				Slot:        resp[0].Destination.Slot,
			},
			GameInterface.ItemLocation{
				WindowID:    0,
				ContainerID: 0xc,
				Slot:        i.AdditionalData.HotBarSlot,
			},
			uint8(itemData.Stack.Count),
			GameInterface.AirItem,
			itemData,
		)
		if err != nil {
			return fmt.Errorf("AddDisplayName: %v", err)
		}
		if res[0].Status != protocol.ItemStackResponseStatusOK {
			return fmt.Errorf("AddDisplayName: Failed to restore the item to its original position")
		}
		// 尝试将物品恢复到原始位置
	}
	// 对于改名后物品不在原始位置的处理(基本)
	return nil
	// 返回值
}
