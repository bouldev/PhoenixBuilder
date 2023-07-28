package NBTAssigner

// 取得用于放置目标方块实体的 接口/方法
func GetPlaceBlockMethod(block *BlockEntity) GeneralBlockNBT {
	switch block.AdditionalData.Type {
	case "CommandBlock":
		return &CommandBlock{BlockEntity: block, ShouldPlaceBlock: true}
	case "Container":
		return &Container{BlockEntity: block}
	case "Sign":
		return &Sign{BlockEntity: block}
	default:
		return &DefaultBlock{BlockEntity: block}
		// 其他尚且未被支持的方块实体
	}
}

// 取得用于生成目标 NBT 物品的 接口/方法
func GetGenerateItemMethod(item *ItemPackage) GeneralItemNBT {
	switch item.AdditionalData.Type {
	case "Book":
		return &Book{ItemPackage: item}
	default:
		return &DefaultItem{ItemPackage: item}
		// 其他尚且未被支持的 NBT 物品
	}
}
