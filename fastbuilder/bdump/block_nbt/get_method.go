package blockNBT

// 取得用于放置目标方块实体的 接口/方法
func getMethod(block *BlockEntity) GeneralBlockNBT {
	switch block.AdditionalData.Type {
	case "CommandBlock":
		return &CommandBlock{BlockEntity: block, ShouldPlaceBlock: true}
	case "Container":
		return &Container{BlockEntity: block}
	case "Sign":
		return &Sign{BlockEntity: block}
	default:
		return &Default{BlockEntity: block}
		// 其他尚且未被支持的方块实体
	}
}
