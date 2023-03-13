package blockNBT_Container

import "fmt"

// 放置一个容器并填充物品
func (c *Container) WriteDatas() error {
	err := c.BlockEntityDatas.API.SetBlock(c.BlockEntityDatas.Datas.Position, c.BlockEntityDatas.Block.Name, c.BlockEntityDatas.Datas.StatesString)
	if err != nil {
		return fmt.Errorf("WriteDatas: %v", err)
	}
	// 放置容器
	for _, value := range c.Items {
		err := c.BlockEntityDatas.API.ReplaceitemToContainer(
			c.BlockEntityDatas.Datas.Position,
			value.Slot,
			value.Name,
			value.Count,
			value.Damage,
			"",
		)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
	}
	// 向容器内填充物品
	return nil
	// 返回值
}
