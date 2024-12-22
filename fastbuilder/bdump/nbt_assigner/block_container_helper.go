package NBTAssigner

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft/protocol"
	"strings"
)

// 获取当前容器方块到快捷栏 5 ，
// containerName 指代该容器的名称
func (c *Container) getContainerItem(containerName string) error {
	var err error
	api := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 初始化
	c.BlockEntity.Block.Name, _, err = get_new_block_states_from_older(
		fmt.Sprintf("minecraft:%s", c.BlockEntity.Block.Name),
		c.BlockEntity.Block.States,
	)
	if err != nil {
		return fmt.Errorf("getContainerItem: %v", err)
	}
	// 取得当前容器方块在新版下的名称
	request := DefaultItem{
		ItemPackage: &ItemPackage{
			Interface: api,
			Item: GeneralItem{
				Basic: ItemBasicData{
					Name:     c.BlockEntity.Block.Name,
					Count:    1,
					MetaData: 0,
				},
				Enhancement: &ItemEnhancementData{
					DisplayName: containerName,
				},
				Custom: nil,
			},
			AdditionalData: ItemAdditionalData{
				Decoded:    false,
				HotBarSlot: 5,
				Position:   c.BlockEntity.AdditionalData.Position,
				Settings:   c.BlockEntity.AdditionalData.Settings,
				FastMode:   c.BlockEntity.AdditionalData.FastMode,
				Others:     c.BlockEntity.AdditionalData.Others,
			},
		},
	}
	err = request.WriteData()
	if err != nil {
		return fmt.Errorf("getContainerItem: %v", err)
	}
	// 从快捷栏 5 获得当前容器对应的物品形式，
	// 但不包括容器内的物品
	return nil
	// 返回值
}

// 从 c.BlockEntity.Block.NBT 获取潜影盒的朝向。
// 此函数仅应当在放置潜影盒时被使用。
// 如果朝向不存在，则重定向为 1(朝上)
func (c *Container) getFacingOfShulkerBox() (uint8, error) {
	if facing_origin, ok := c.BlockEntity.Block.NBT["facing"]; ok {
		facing_got, success := facing_origin.(byte)
		if !success {
			return 0, fmt.Errorf(`getFacingOfShulkerBox: Can not convert facing_origin into byte(uint8); c.BlockEntity.Block.NBT = %#v`, c.BlockEntity.Block.NBT)
		}
		return facing_got, nil
	}
	return 1, nil
}

// 从 c.BlockEntity.Block.NBT 获取容器的名字
func (c *Container) getNameOfContainer() (containerName string) {
	if c.BlockEntity.Block.NBT == nil {
		return
	}
	if customName, ok := c.BlockEntity.Block.NBT["CustomName"].(string); ok {
		return customName
	}
	return
}

// 放置 c.BlockEntity 所代表的容器。
// 此函数侧重于对潜影盒的专门化处理，
// 以保证放置出的潜影盒能拥有正确的朝向
func (c *Container) PlaceContainer() error {
	api := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	containerFacing := uint8(2)
	isShulkerBox := strings.Contains(c.BlockEntity.Block.Name, "shulker_box")
	containerName := c.getNameOfContainer()
	// 初始化
	if isShulkerBox || len(containerName) > 0 {
		err := api.SendSettingsCommand(
			fmt.Sprintf(
				"tp %d %d %d",
				c.BlockEntity.AdditionalData.Position[0],
				c.BlockEntity.AdditionalData.Position[1],
				c.BlockEntity.AdditionalData.Position[2],
			),
			true,
		)
		if err != nil {
			return fmt.Errorf("PlaceContainer: %v", err)
		}
		// 将机器人传送到当前容器所在位置
		err = api.ChangeSelectedHotbarSlot(5)
		if err != nil {
			return fmt.Errorf("PlaceContainer: %v", err)
		}
		// 切换物品栏
		err = c.getContainerItem(containerName)
		if err != nil {
			return fmt.Errorf("PlaceContainer: %v", err)
		}
		// 获得容器对应的物品形式，
		// 但不包括容器中的物品
		if isShulkerBox {
			containerFacing, err = c.getFacingOfShulkerBox()
			if err != nil {
				return fmt.Errorf("PlaceContainer: %v", err)
			}
		}
		// 如果当前容器是一个潜影盒，
		// 则尝试获得该潜影盒的放置朝向
		err = api.PlaceBlockWithFacing(c.BlockEntity.AdditionalData.Position, 5, containerFacing)
		if err != nil {
			return fmt.Errorf("PlaceContainer: %v", err)
		}
		// 生成容器
		err = api.AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf("PlaceContainer: %v", err)
		}
		// 等待更改
		if isShulkerBox {
			return nil
		}
		// 对于潜影盒，可以直接返回值。
		// 对于其他容器，需要修正它们的朝向
	}
	// 放置潜影盒或带有自定义名称的容器
	err := api.SetBlock(
		c.BlockEntity.AdditionalData.Position,
		c.BlockEntity.Block.Name,
		c.BlockEntity.AdditionalData.BlockStates,
	)
	if err != nil {
		return fmt.Errorf("PlaceContainer: %v", err)
	}
	// 放置不带自定义名称的容器，
	// 或修正带自定义名称的容器的朝向
	return nil
	// 返回值
}

// 打开已放置的容器，因此该函数应当后于 PlaceContainer 执行。
//
// 返回的布尔值代表该容器是否成功打开，
// 如果打开失败，则返回假，否则返回真。
//
// 请确保在使用此函数前占用了容器资源，否则会造成程序 panic
func (c *Container) OpenContainer() (bool, error) {
	api := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	backupBlockPos := c.BlockEntity.AdditionalData.Position
	// 初始化
	if strings.Contains(c.BlockEntity.Block.Name, "shulker_box") || strings.Contains(c.BlockEntity.Block.Name, "chest") {
		if strings.Contains(c.BlockEntity.Block.Name, "shulker_box") {
			facing, err := c.getFacingOfShulkerBox()
			if err != nil {
				return false, fmt.Errorf("OpenContainer: %v", err)
			}
			switch facing {
			case 0:
				backupBlockPos[1] = backupBlockPos[1] - 1
			case 1:
				backupBlockPos[1] = backupBlockPos[1] + 1
			case 2:
				backupBlockPos[2] = backupBlockPos[2] - 1
			case 3:
				backupBlockPos[2] = backupBlockPos[2] + 1
			case 4:
				backupBlockPos[0] = backupBlockPos[0] - 1
			case 5:
				backupBlockPos[0] = backupBlockPos[0] + 1
			}
		} else {
			backupBlockPos[1] = backupBlockPos[1] + 1
		}
		// 确定容器开启方向上前方一格方块的位置
		uniqueId, err := api.BackupStructure(GameInterface.MCStructure{
			BeginX: backupBlockPos[0],
			BeginY: backupBlockPos[1],
			BeginZ: backupBlockPos[2],
			SizeX:  1,
			SizeY:  1,
			SizeZ:  1,
		})
		if err != nil {
			return false, fmt.Errorf("OpenContainer: %v", err)
		}
		defer api.RevertStructure(uniqueId, backupBlockPos)
		err = api.SetBlockAsync(backupBlockPos, "air", "[]")
		if err != nil {
			return false, fmt.Errorf("OpenContainer: %v", err)
		}
		/*
			我们需要保证潜影盒开启方向上的方块为空气，
			否则潜影盒将无法正常开启。
			然而，对这个方块进行操作不是预期的行为，
			所以需要确定其坐标并发起一次备份，
			然后强行将其变更为空气
		*/
	}
	// 对潜影盒或者箱子的特殊化处理
	err := api.ChangeSelectedHotbarSlot(5)
	if err != nil {
		return false, fmt.Errorf("OpenContainer: %v", err)
	}
	success, err := api.OpenContainer(
		c.BlockEntity.AdditionalData.Position,
		fmt.Sprintf("minecraft:%s", c.BlockEntity.Block.Name),
		c.BlockEntity.Block.States,
		5,
	)
	if err != nil {
		return false, fmt.Errorf("OpenContainer: %v", err)
	}
	// 将快捷栏切换至 5 号槽位，
	// 然后使用该槽位的物品点击容器，
	// 以达到开启容器的目的
	return success, nil
	// 返回值
}

// 将背包中 itemLocation 处的物品移动到
// 容器的 destination 处，
//
// 此函数将会自动占用、释放容器资源，
// 并且打开容器
func (c *Container) MoveItemIntoContainer(
	itemLocation uint8,
	destination uint8,
) error {
	var containerID uint8
	api := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 初始化
	holder := api.Resources.Container.Occupy()
	defer api.Resources.Container.Release(holder)
	// 占用容器资源
	success, err := c.OpenContainer()
	if err != nil {
		return fmt.Errorf("MoveItemIntoContainer: %v", err)
	}
	if !success {
		return nil
	}
	defer api.CloseContainer()
	// 打开已放置的容器
	itemData, err := api.Resources.Inventory.GetItemStackInfo(0, itemLocation)
	if err != nil {
		return fmt.Errorf("MoveItemIntoContainer: %v", err)
	}
	container_opening_data := api.Resources.Container.GetContainerOpeningData()
	got := SupportContainerPool[c.BlockEntity.Block.Name]
	// 获取 itemLocation 处的物品数据，
	// 以及已打开容器的数据
	if len(got.ContainerIDMapping) > 0 {
		containerID = got.ContainerIDMapping[destination]
	} else {
		containerID = got.UniversalContainerID
	}
	// 确定容器 ID
	_, err = api.MoveItem(
		GameInterface.ItemLocation{
			WindowID:    0,
			ContainerID: GameInterface.ContainerIDInventory,
			Slot:        itemLocation,
		},
		GameInterface.ItemLocation{
			WindowID:    container_opening_data.WindowID,
			ContainerID: containerID,
			Slot:        destination,
		},
		uint8(itemData.Stack.Count),
		GameInterface.AirItem,
		itemData,
	)
	if err != nil && err != GameInterface.ErrMoveItemCheckFailure {
		return fmt.Errorf("MoveItemIntoContainer: %v", err)
	}
	// 将物品移动到容器中
	return nil
	// 返回值
}

// 将 item 所指代的子方块获取到物品栏。
// 如果 item 有自定义的物品显示名称，
// 则还会使用铁砧进行改名。
// 返回的布尔值代表以上操作是否成功，
// 返回的 uint8 代表子方块在快捷栏的生成位置
func (c *Container) GetSubBlock(
	item GeneralItem,
) (bool, uint8, error) {
	api := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 初始化
	err := api.SendSettingsCommand("clear", true)
	if err != nil {
		return false, 0, fmt.Errorf("GetSubBlock: %v", err)
	}
	// 清除物品栏
	uniqueId, err := api.BackupStructure(
		GameInterface.MCStructure{
			BeginX: c.BlockEntity.AdditionalData.Position[0],
			BeginY: c.BlockEntity.AdditionalData.Position[1],
			BeginZ: c.BlockEntity.AdditionalData.Position[2],
			SizeX:  1,
			SizeY:  1,
			SizeZ:  1,
		},
	)
	if err != nil {
		return false, 0, fmt.Errorf("GetSubBlock: %v", err)
	}
	defer api.RevertStructure(uniqueId, c.BlockEntity.AdditionalData.Position)
	// 备份容器
	err = item.Custom.SubBlockData.Decode()
	if err != nil {
		return false, 0, fmt.Errorf("GetSubBlock: %v", err)
	}
	err = item.Custom.SubBlockData.WriteData()
	if err != nil {
		return false, 0, fmt.Errorf("GetSubBlock: %v", err)
	}
	// 解码并放置子方块
	err = api.AwaitChangesGeneral()
	if err != nil {
		return false, 0, fmt.Errorf("GetSubBlock: %v", err)
	}
	// 等待更改
	success, spawnLocation, err := api.PickBlock(
		c.BlockEntity.AdditionalData.Position,
		true,
	)
	if err != nil {
		return false, 0, fmt.Errorf("GetSubBlock: %v", err)
	}
	if !success {
		return false, 0, nil
	}
	// 获取方块到物品栏
	if item.Enhancement != nil && len(item.Enhancement.DisplayName) != 0 {
		resp, err := api.RenameItemByAnvil(
			c.BlockEntity.AdditionalData.Position,
			`["direction"=0,"damage"="undamaged"]`,
			5,
			[]GameInterface.ItemRenamingRequest{
				{
					Slot: spawnLocation,
					Name: item.Enhancement.DisplayName,
				},
			},
		)
		if err != nil {
			return false, 0, fmt.Errorf("GetSubBlock: %v", err)
		}
		if resp[0].Destination == nil {
			return false, 0, fmt.Errorf("GetSubBlock: Inventory was full")
		}
		spawnLocation = resp[0].Destination.Slot
	}
	// 如果这个子方块有自定义的物品显示名称
	if item.Basic.Count > 1 {
		err = api.CopyItem(spawnLocation, c.BlockEntity.AdditionalData.Position, item.Basic.Count)
		if err != nil {
			return false, 0, fmt.Errorf("GetSubBlock: %v", err)
		}
	}
	// 如果这个子方块具有堆叠属性，例如告示牌
	return true, spawnLocation, nil
	// 返回值
}

// 获取 itemPackage.Item 所指代的 NBT 物品到快捷栏 5 。
// 如果 itemPackage.Item 有自定义的物品显示名称或附魔属性，
// 则还会使用铁砧进行改名并使用 enchant 命令附魔。
//
// 返回的布尔值代表以上操作是否成功
func (c *Container) GetNBTItem(
	itemPackage ItemPackage,
) (bool, error) {
	api := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 初始化
	err := api.SendSettingsCommand("clear", true)
	if err != nil {
		return false, fmt.Errorf("GetNBTItem: %v", err)
	}
	// 清除物品栏
	uniqueId, err := api.BackupStructure(
		GameInterface.MCStructure{
			BeginX: c.BlockEntity.AdditionalData.Position[0],
			BeginY: c.BlockEntity.AdditionalData.Position[1],
			BeginZ: c.BlockEntity.AdditionalData.Position[2],
			SizeX:  1,
			SizeY:  1,
			SizeZ:  1,
		},
	)
	if err != nil {
		return false, fmt.Errorf("GetNBTItem: %v", err)
	}
	defer api.RevertStructure(uniqueId, c.BlockEntity.AdditionalData.Position)
	// 备份容器
	err = itemPackage.Item.Custom.NBTItemData.Decode()
	if err != nil {
		return false, fmt.Errorf("GetNBTItem: %v", err)
	}
	err = itemPackage.Item.Custom.NBTItemData.WriteData()
	if err != nil {
		return false, fmt.Errorf("GetNBTItem: %v", err)
	}
	// 解码并取得该 NBT 物品
	err = api.AwaitChangesGeneral()
	if err != nil {
		return false, fmt.Errorf("GetNBTItem: %v", err)
	}
	// 等待更改
	return true, nil
	// 返回值
}

// 将 contents 中仅包含附魔属性、
// 物品组件和自定义物品显示名称的物品
// 放入容器。
// 返回的物品列表代表应当直接在容器上
// 应用 replaceitem 命令的物品项目
func (c *Container) ItemPlanner(contents []GeneralItem) ([]GeneralItem, error) {
	var containerID uint8
	var needOpenInventory bool
	var needOpenContainer bool
	moveIndex := map[uint8]GeneralItem{}
	defaultSituation := []GeneralItem{}
	api := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 初始化
	{
		current := 0
		firstFiltration := []GeneralItem{}
		// 初始化
		if !ContainerCouldOpen(c.BlockEntity.Block.Name) {
			return contents, nil
		}
		// 如果该容器不可被打开，
		// 则所有的物品均应当使用默认放入方法，
		// 此时将直接返回值
		err := api.SendSettingsCommand("clear", true)
		if err != nil {
			return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
		}
		// 清空物品栏
		for _, value := range contents {
			if value.Enhancement != nil && value.Enhancement.ItemComponents != nil && len(value.Enhancement.ItemComponents.ItemLock) != 0 {
				defaultSituation = append(defaultSituation, value)
				continue
			}
			// 如果该物品存在 item_lock 物品组件，
			// 则将其忽略，因为存在该组件的物品不能跨容器移动
			if (value.Enhancement == nil && value.Custom == nil) || (value.Enhancement != nil && len(value.Enhancement.DisplayName) == 0 && value.Enhancement.Enchantments == nil) {
				defaultSituation = append(defaultSituation, value)
				continue
			}
			// 这是一个普通的物品，
			// 可以直接在容器上应用 replaceitem 命令
			if value.Custom != nil {
				continue
			}
			// 此类物品已在之前被处理过了
			firstFiltration = append(firstFiltration, value)
			needOpenContainer = true
			// 这些物品需要被特殊处理
		}
		// 首次过滤
		for _, value := range firstFiltration {
			if value.Enhancement != nil && value.Enhancement.Enchantments != nil {
				needOpenInventory = true
				switch {
				case current == 8:
					moveIndex[uint8(current+1)] = value
					current = current + 2
				default:
					moveIndex[uint8(current)] = value
					current++
				}
			}
		}
		// 过滤出包含附魔属性的物品
		if current <= 8 {
			current = 9
		}
		for _, value := range firstFiltration {
			if value.Enhancement != nil && value.Enhancement.Enchantments == nil && len(value.Enhancement.DisplayName) != 0 {
				moveIndex[uint8(current)] = value
				current++
			}
		}
		// 过滤出**仅**包含自定义物品显示名称的物品
	}
	// 确定物品的生成位置
	{
		for key, value := range moveIndex {
			if key <= 8 {
				err := api.ReplaceItemInInventory(
					GameInterface.TargetMySelf,
					GameInterface.ItemGenerateLocation{
						Path: "slot.hotbar",
						Slot: uint8(key),
					},
					types.ChestSlot{
						Name:   value.Basic.Name,
						Count:  value.Basic.Count,
						Damage: value.Basic.MetaData,
					},
					MarshalItemComponents(value.Enhancement.ItemComponents),
					false,
				)
				if err != nil {
					return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
				}
			} else if value.Enhancement.Enchantments == nil {
				err := api.ReplaceItemInInventory(
					GameInterface.TargetMySelf,
					GameInterface.ItemGenerateLocation{
						Path: "slot.inventory",
						Slot: uint8(key - 9),
					},
					types.ChestSlot{
						Name:   value.Basic.Name,
						Count:  value.Basic.Count,
						Damage: value.Basic.MetaData,
					},
					MarshalItemComponents(value.Enhancement.ItemComponents),
					false,
				)
				if err != nil {
					return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
				}
			}
		}
		err := api.AwaitChangesGeneral()
		if err != nil {
			return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
		}
	}
	/*
		将物品生成到背包中。

		附魔物品的优先级是最高的，
		它们会先被优先放置到快捷栏。

		剩余的物品(不含存在附魔属性的物品)只会被放置到背包，
		而不会是快捷栏，即便不存在附魔物品。

		注：
		此处为第 9 个快捷栏保留了一个槽位，
		这用于为剩下还未处理的附魔物品附魔
	*/
	{
		for key, value := range moveIndex {
			if key >= 8 || value.Enhancement.Enchantments == nil {
				continue
			}
			err := api.ChangeSelectedHotbarSlot(uint8(key))
			if err != nil {
				return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
			}
			for _, v := range *value.Enhancement.Enchantments {
				err = api.SendSettingsCommand(
					fmt.Sprintf(
						"enchant @s %d %d",
						v.ID,
						v.Level,
					),
					true,
				)
				if err != nil {
					return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
				}
			}
			err = api.AwaitChangesGeneral()
			if err != nil {
				return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
			}
		}
	}
	// 将快捷栏中的物品逐个附魔
	{
		subFunc := func() error {
			if needOpenInventory {
				err := api.ChangeSelectedHotbarSlot(8)
				if err != nil {
					return fmt.Errorf("subFunc: %v", err)
				}
				holder := api.Resources.Container.Occupy()
				defer api.Resources.Container.Release(holder)
				success, err := api.OpenInventory()
				if err != nil {
					return fmt.Errorf("subFunc: %v", err)
				}
				if !success {
					return nil
				}
				defer api.CloseContainer()
			}
			// 切换物品栏到保留的物品栏并占用容器资源，
			// 然后打开物品栏
			for key, value := range moveIndex {
				if key < 8 || value.Enhancement.Enchantments == nil {
					continue
				}
				err := api.ReplaceItemInInventory(
					GameInterface.TargetMySelf,
					GameInterface.ItemGenerateLocation{
						Path: "slot.hotbar",
						Slot: 8,
					},
					types.ChestSlot{
						Name:   value.Basic.Name,
						Count:  value.Basic.Count,
						Damage: value.Basic.MetaData,
					},
					MarshalItemComponents(value.Enhancement.ItemComponents),
					true,
				)
				if err != nil {
					return fmt.Errorf("subFunc: %v", err)
				}
				// 在第 9 个快捷栏生成 value 所指代的物品
				for _, v := range *value.Enhancement.Enchantments {
					err = api.SendSettingsCommand(
						fmt.Sprintf(
							"enchant @s %d %d",
							v.ID,
							v.Level,
						),
						true,
					)
					if err != nil {
						return fmt.Errorf("subFunc: %v", err)
					}
				}
				// 附加附魔属性
			ReTry:
				err = api.ReplaceItemInInventory(
					GameInterface.TargetMySelf,
					GameInterface.ItemGenerateLocation{
						Path: "slot.inventory",
						Slot: key - 9,
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
					return fmt.Errorf("subFunc: %v", err)
				}
				itemData, err := api.Resources.Inventory.GetItemStackInfo(0, 8)
				if err != nil {
					return fmt.Errorf("subFunc: %v", err)
				}
				resp, err := api.MoveItem(
					GameInterface.ItemLocation{
						WindowID:    0,
						ContainerID: GameInterface.ContainerIDInventory,
						Slot:        8,
					},
					GameInterface.ItemLocation{
						WindowID:    0,
						ContainerID: GameInterface.ContainerIDInventory,
						Slot:        uint8(key),
					},
					uint8(itemData.Stack.Count),
					GameInterface.AirItem,
					itemData,
				)
				if err != nil && err != GameInterface.ErrMoveItemCheckFailure {
					return fmt.Errorf("subFunc: %v", err)
				}
				if api.Resources.Container.GetContainerOpeningData() != nil && resp[0].Status != protocol.ItemStackResponseStatusOK {
					goto ReTry
				}
				// 将该物品移动到背包中
			}
			// 逐一遍历剩余的附魔物品(它们一定在背包而非快捷栏中)，
			// 然后每次遍历时将其生成在保留的第 9 个快捷栏，
			// 然后为其附魔并其放置到背包中安排好的位置
			return nil
			// 返回值
		}
		// 考虑到相关联的操作需要开启背包，
		// 因此这里构造了一个子函数以使用 defer 特性
		err := subFunc()
		if err != nil {
			return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
		}
		// 调用子函数
	}
	// 对于剩余物品的附魔处理
	{
		request := []GameInterface.ItemRenamingRequest{}
		requestIndex := map[int]uint8{} // location_in_request -> location_in_inventory
		// 初始化
		for key, value := range moveIndex {
			if len(value.Enhancement.DisplayName) != 0 {
				request = append(request, GameInterface.ItemRenamingRequest{
					Slot: key,
					Name: value.Enhancement.DisplayName,
				})
				requestIndex[len(request)-1] = key
			}
		}
		// 整理物品名称修改请求
		if len(request) > 0 {
			result, err := api.RenameItemByAnvil(
				c.BlockEntity.AdditionalData.Position,
				`["direction"=0,"damage"="undamaged"]`,
				5,
				request,
			)
			if err != nil {
				return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
			}
			// 发送物品修改请求
			copy := map[uint8]GeneralItem{}
			for key, value := range moveIndex {
				copy[key] = value
			}
			for key := range request {
				delete(moveIndex, requestIndex[key])
			}
			for key, value := range result {
				if value.Destination != nil {
					moveIndex[value.Destination.Slot] = copy[requestIndex[key]]
				}
			}
			// 将物品映射表修正到正确的位置
		}
		// 物品名称修改
	}
	// 对于所有物品的物品名称的处理
	{
		if !needOpenContainer {
			return defaultSituation, nil
		}
		// 确定容器是否需要打开
		holder := api.Resources.Container.Occupy()
		defer api.Resources.Container.Release(holder)
		success, err := c.OpenContainer()
		if err != nil {
			return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
		}
		if !success {
			return []GeneralItem{}, fmt.Errorf("ItemPlanner: Failed to open the container named %s", c.BlockEntity.Block.Name)
		}
		defer api.CloseContainer()
		// 占用容器资源并打开容器
		container_opening_data := api.Resources.Container.GetContainerOpeningData()
		got := SupportContainerPool[c.BlockEntity.Block.Name]
		// 获取已打开容器的数据
		for key, value := range moveIndex {
		ReTry:
			itemData, err := api.Resources.Inventory.GetItemStackInfo(0, key)
			if err != nil {
				return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
			}
			// 获取物品数据
			if itemData.Stack.NetworkID == 0 {
				continue
			}
			// 如果当前物品是空气，
			// 那么忽略当前物品并继续
			if len(got.ContainerIDMapping) > 0 {
				containerID = got.ContainerIDMapping[value.Basic.Slot]
			} else {
				containerID = got.UniversalContainerID
			}
			// 确定容器 ID
			resp, err := api.MoveItem(
				GameInterface.ItemLocation{
					WindowID:    0,
					ContainerID: GameInterface.ContainerIDInventory,
					Slot:        key,
				},
				GameInterface.ItemLocation{
					WindowID:    container_opening_data.WindowID,
					ContainerID: containerID,
					Slot:        value.Basic.Slot,
				},
				uint8(itemData.Stack.Count),
				GameInterface.AirItem,
				itemData,
			)
			if err != nil && err != GameInterface.ErrMoveItemCheckFailure {
				return []GeneralItem{}, fmt.Errorf("ItemPlanner: %v", err)
			}
			if api.Resources.Container.GetContainerOpeningData() != nil && resp[0].Status != protocol.ItemStackResponseStatusOK {
				goto ReTry
			}
			// 将当前物品移动到容器
		}
	}
	// 移动物品到容器
	return defaultSituation, nil
	// 返回值
}
