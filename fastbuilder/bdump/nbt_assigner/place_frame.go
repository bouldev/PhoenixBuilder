package NBTAssigner

import (
	"fmt"
	GameInterface "phoenixbuilder/game_control/game_interface"
)

// 从 f.BlockEntity.Block.NBT 提取物品展示框数据，
// 并保存在 f.FrameData 中
func (f *Frame) Decode() error {
	// 初始化
	var itemRotation float32 = 0
	var item map[string]any
	var itemPackage *ItemPackage
	var normal bool
	// ItemRotation
	_, ok := f.BlockEntity.Block.NBT["ItemRotation"]
	if ok {
		itemRotation, normal = f.BlockEntity.Block.NBT["ItemRotation"].(float32)
		if !normal {
			return fmt.Errorf("Decode: Crashed at f.BlockEntity.Block.NBT[\"ItemRotation\"]; f.BlockEntity.Block.NBT = %#v", f.BlockEntity.Block.NBT)
		}
	}
	// Item
	_, ok = f.BlockEntity.Block.NBT["Item"]
	if ok {
		item, normal = f.BlockEntity.Block.NBT["Item"].(map[string]any)
		if !normal {
			return fmt.Errorf("Decode: Crashed at f.BlockEntity.Block.NBT[\"Item\"]; f.BlockEntity.Block.NBT = %#v", f.BlockEntity.Block.NBT)
		}
		itemPackage = &ItemPackage{
			Interface: f.BlockEntity.Interface,
			Item:      GeneralItem{},
			AdditionalData: ItemAdditionalData{
				HotBarSlot: 5,
				Position:   f.BlockEntity.AdditionalData.Position,
				Type:       "",
				Settings:   f.BlockEntity.AdditionalData.Settings,
				FastMode:   f.BlockEntity.AdditionalData.FastMode,
				Others:     f.BlockEntity.AdditionalData.Others,
			},
		}
		err := itemPackage.ParseItemFromNBT(item)
		if err != nil {
			return fmt.Errorf("Decode: %v", err)
		}
	}
	// return
	f.FrameData = FrameData{
		ItemRotation: itemRotation,
		Item:         itemPackage,
	}
	return nil
}

// 放置一个物品展示框并写入物品展示框数据
func (f *Frame) WriteData() error {
	// 初始化
	itemSpawnSlot := uint8(5)
	api := f.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 放置物品展示框(快速模式下)
	if f.BlockEntity.AdditionalData.FastMode {
		err := api.SetBlockAsync(f.BlockEntity.AdditionalData.Position, f.BlockEntity.Block.Name, f.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		return nil
	}
	// 放置物品展示框(正常情况下)
	err := api.SetBlock(f.BlockEntity.AdditionalData.Position, "minecraft:air", "[]")
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	err = api.SetBlock(f.BlockEntity.AdditionalData.Position, f.BlockEntity.Block.Name, f.BlockEntity.AdditionalData.BlockStates)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 判断物品展示框中是否存在物品
	if f.FrameData.Item == nil {
		return nil
	}
	// 对于复杂的方块或物品
	if f.FrameData.Item.Item.Custom != nil {
		// 准备
		container := Container{
			BlockEntity: f.BlockEntity,
			Contents:    nil,
		}
		// 子方块
		if f.FrameData.Item.Item.Custom.SubBlockData != nil {
			success, spawnLocation, err := container.GetSubBlock(f.FrameData.Item.Item)
			if err != nil {
				return fmt.Errorf("WriteData: Failed to process the sub block in frame block; f.FrameData.Item.Item.Custom.SubBlockData = %#v; err = %v", f.FrameData.Item.Item.Custom.SubBlockData, err)
			}
			if !success {
				return fmt.Errorf("WriteData: Failed to get the sub block in frame block due to unknown reason; f.FrameData.Item.Item.Custom.SubBlockData = %#v", f.FrameData.Item.Item.Custom.SubBlockData)
			}
			itemSpawnSlot = spawnLocation
		}
		// NBT 物品
		if f.FrameData.Item.Item.Custom.ItemTag != nil {
			success, err := container.GetNBTItem(*f.FrameData.Item)
			if err != nil {
				return fmt.Errorf("WriteData: Failed to process the nbt item in frame block; f.FrameData.Item.Item.Custom.ItemTag = %#v; err = %v", f.FrameData.Item.Item.Custom.ItemTag, err)
			}
			if !success {
				return fmt.Errorf("WriteData: Failed to get the nbt item in frame block due to unknown reason; f.FrameData.Item.Item.Custom.ItemTag = %#v", f.FrameData.Item.Item.Custom.ItemTag)
			}
		}
	}
	// 对于无需通过复杂工序制作的物品
	if f.FrameData.Item.Item.Custom == nil {
		err = GetGenerateItemMethod(f.FrameData.Item).WriteData()
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 切换物品栏
	err = api.ChangeSelectedHotbarSlot(itemSpawnSlot)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 放入物品并使其旋转
	clickCount := 1 + int(f.FrameData.ItemRotation/45)
	for i := 0; i < clickCount; i++ {
		api.ClickBlock(GameInterface.UseItemOnBlocks{
			HotbarSlotID: itemSpawnSlot,
			BlockPos:     f.BlockEntity.AdditionalData.Position,
			BlockName:    f.BlockEntity.Block.Name,
			BlockStates:  f.BlockEntity.Block.States,
		})
	}
	// 返回值
	return nil
}
