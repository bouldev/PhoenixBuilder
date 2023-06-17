package blockNBT

import "fmt"

// Default 结构体用于描述一个完整的方块实体数据。
// 任何未被支持的方块实体都会被重定向为此结构体
type Default struct {
	Package *Package // 该方块实体的详细数据
}

// 这只是为了保证接口一致而设
func (d *Default) Decode() error {
	return nil
}

// 放置目标方块但不写入方块实体数据
func (d *Default) WriteDatas() error {
	err := d.Package.API.SetBlockFastly(d.Package.Datas.Position, d.Package.Block.Name, d.Package.Datas.StatesString)
	if err != nil {
		return fmt.Errorf("placeBlockWithNBTData: %v", err)
	}
	return nil
}
