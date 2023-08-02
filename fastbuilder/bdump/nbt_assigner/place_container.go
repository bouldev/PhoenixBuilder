package NBTAssigner

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	GameInterface "phoenixbuilder/game_control/game_interface"
)

// 从容器的 NBT 数据中提取物品数据。
// 返回的列表代表提取到的每个物品
func (c *Container) getContainerContents() ([]ItemOrigin, error) {
	got, ok := SupportContainerPool[c.BlockEntity.Block.Name]
	if !ok {
		return nil, ErrNotASupportedContainer
	}
	// 确定目标容器是否已被支持
	itemContentsGot, ok := c.BlockEntity.Block.NBT[got.StorageItemValue]
	// 从 containerOriginNBT 获取物品的数据
	if !ok {
		return []ItemOrigin{}, nil
	}
	// 对于唱片机和讲台这种容器，如果它们没有被放物品的话，
	// 那么对应的 key 是找不到的，但是这并非是错误
	switch itemContents := itemContentsGot.(type) {
	case map[string]interface{}:
		return []ItemOrigin{itemContents}, nil
		// 如果这个物品是一个唱片机或者讲台，
		// 那么传入的 itemContents 是一个复合标签而非列表。
		// 因此，为了统一数据格式，
		// 我们将复合标签处理成通常情况下的列表
	case []interface{}:
		res := []ItemOrigin{}
		for key, value := range itemContents {
			singleItem, success := value.(map[string]interface{})
			if !success {
				return nil, fmt.Errorf(`getContainerContents: Crashed on itemContents[%d]; itemContents = %#v`, key, itemContents)
			}
			res = append(res, singleItem)
		}
		return res, nil
		// 常规型物品的(多个)物品数据存放在一张表中，
		// 而每个物品都用一个复合标签来描述
	default:
		return nil, fmt.Errorf(`getContainerContents: Unexpected data type of itemContentsGot; itemContentsGot = %#v`, itemContentsGot)
	}
	// 处理方块实体数据并返回值
}

// 从 c.Package.Block.NBT 提取物品数据并保存在 c.Contents 中
func (c *Container) Decode() error {
	itemContents, err := c.getContainerContents()
	if err != nil {
		return fmt.Errorf("Decode: %v", err)
	}
	// 获取容器内的物品数据
	for _, value := range itemContents {
		newPackage := ItemPackage{
			Interface: c.BlockEntity.Interface,
			Item:      GeneralItem{},
			AdditionalData: ItemAdditionalData{
				HotBarSlot: 5,
				Position:   c.BlockEntity.AdditionalData.Position,
				Type:       "",
				Settings:   c.BlockEntity.AdditionalData.Settings,
				FastMode:   c.BlockEntity.AdditionalData.FastMode,
				Others:     c.BlockEntity.AdditionalData.Others,
			},
		}
		err := newPackage.ParseItemFromNBT(value)
		if err != nil {
			return fmt.Errorf("Decode: %v", err)
		}
		c.Contents = append(c.Contents, newPackage)
	}
	// 解码
	return nil
	// 返回值
}

// 放置一个容器并填充物品。
// 这是针对 FastMode 模式的专门化处理
func (c *Container) FastWrite() error {
	err := c.BlockEntity.Interface.SetBlockAsync(c.BlockEntity.AdditionalData.Position, c.BlockEntity.Block.Name, c.BlockEntity.AdditionalData.BlockStates)
	if err != nil {
		return fmt.Errorf("FastWrite: %v", err)
	}
	// 放置容器
	for _, value := range c.Contents {
		err := c.BlockEntity.Interface.(*GameInterface.GameInterface).ReplaceItemInContainerAsync(
			c.BlockEntity.AdditionalData.Position,
			types.ChestSlot{
				Name:   value.Item.Basic.Name,
				Count:  value.Item.Basic.Count,
				Damage: value.Item.Basic.MetaData,
				Slot:   value.Item.Basic.Slot,
			},
			"",
		)
		if err != nil {
			return fmt.Errorf("FastWrite: %v", err)
		}
	}
	// 向容器内填充物品
	return nil
	// 返回值
}

// 放置一个容器并填充物品
func (c *Container) WriteData() error {
	if c.BlockEntity.AdditionalData.FastMode {
		err := c.FastWrite()
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 针对 FastMode 模式的专门化处理
	err := c.PlaceContainer()
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 放置容器
	for key, value := range c.Contents {
		if ContainerCouldOpen(c.BlockEntity.Block.Name) && value.Item.Custom != nil && value.Item.Custom.SubBlockData != nil {
			success, spawnLocation, err := c.GetSubBlock(value.Item)
			if err != nil {
				return fmt.Errorf("WriteData: Failed to process the sub block from c.Contents[%d]; c.Contents[%d].Item.Custom.SubBlockData = %#v; err = %v", key, key, value.Item.Custom.SubBlockData, err)
			}
			if !success {
				continue
			}
			// 获取子方块的物品形式
			err = c.MoveItemIntoContainer(spawnLocation, value.Item.Basic.Slot)
			if err != nil {
				return fmt.Errorf("WriteData: Failed to process the sub block from c.Contents[%d]; c.Contents[%d].Item.Custom.SubBlockData = %#v; err = %v", key, key, value.Item.Custom.SubBlockData, err)
			}
			// 将子方块移动到容器中
			continue
		}
		// 子方块
		if value.Item.Custom != nil && value.Item.Custom.ItemTag != nil {
			success, err := c.GetNBTItem(value)
			if err != nil {
				return fmt.Errorf("WriteData: Failed to process the nbt item from c.Contents[%d]; c.Contents[%d].Item.Custom.ItemTag = %#v; err = %v", key, key, value.Item.Custom.ItemTag, err)
			}
			if !success {
				continue
			}
			err = c.MoveItemIntoContainer(5, value.Item.Basic.Slot)
			if err != nil {
				return fmt.Errorf("WriteData: Failed to process the nbt item from c.Contents[%d]; c.Contents[%d].Item.Custom.ItemTag = %#v; err = %v", key, key, value.Item.Custom.ItemTag, err)
			}
		}
		// NBT 物品
	}
	// 针对子方块或 NBT 物品的特殊化处理
	newList := []GeneralItem{}
	for _, value := range c.Contents {
		newList = append(newList, value.Item)
	}
	defaultSituation, err := c.ItemPlanner(newList)
	if err != nil {
		return fmt.Errorf("GetSubBlock: %v", err)
	}
	// 仅包含附魔属性、 物品组件和自定义物品显示名称的物品
	for _, value := range defaultSituation {
		var itemComponents *ItemComponents
		if value.Enhancement != nil {
			itemComponents = value.Enhancement.ItemComponents
		}
		err := c.BlockEntity.Interface.(*GameInterface.GameInterface).ReplaceItemInContainerAsync(
			c.BlockEntity.AdditionalData.Position,
			types.ChestSlot{
				Name:   value.Basic.Name,
				Count:  value.Basic.Count,
				Damage: value.Basic.MetaData,
				Slot:   value.Basic.Slot,
			},
			MarshalItemComponents(itemComponents),
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 对于可以直接在容器上应用 replaceitem 命令的物品
	return nil
	// 返回值
}
