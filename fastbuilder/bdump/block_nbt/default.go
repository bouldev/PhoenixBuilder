package blockNBT

import "fmt"

// Default 结构体用于描述一个完整的方块实体数据。
// 任何未被支持的方块实体都会被重定向为此结构体
type Default struct {
	BlockEntity *BlockEntity // 该方块实体的详细数据
}

// 这只是为了保证接口一致而设
func (d *Default) Decode() error {
	return nil
}

// 放置目标方块但不写入方块实体数据
func (d *Default) WriteData() error {
	err := d.BlockEntity.Interface.SetBlockAsync(d.BlockEntity.AdditionalData.Position, d.BlockEntity.Block.Name, d.BlockEntity.AdditionalData.BlockStates)
	if err != nil {
		return fmt.Errorf("placeBlockWithNBTData: %v", err)
	}
	return nil
}
