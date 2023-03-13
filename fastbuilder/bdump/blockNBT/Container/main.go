package blockNBT_Container

import "fmt"

// 放置容器并填入物品的主函数
func (c *Container) Main() error {
	err := c.Decode()
	if err != nil {
		return fmt.Errorf("Main: %v", err)
	}
	err = c.WriteDatas()
	if err != nil {
		return fmt.Errorf("Main: %v", err)
	}
	return nil
}
