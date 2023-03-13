package blockNBT_CommandBlock

import "fmt"

// 放置命令方块并写入 NBT 数据的主函数
func (c *CommandBlock) Main() error {
	err := c.Decode()
	if err != nil {
		return fmt.Errorf("Main: %v", err)
	}
	err = c.WriteDatas(true)
	if err != nil {
		return fmt.Errorf("Main: %v", err)
	}
	return nil
}
