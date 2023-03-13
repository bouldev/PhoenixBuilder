package blockNBT_API

import "fmt"

// 向容器填充物品
func (g *GlobalAPI) ReplaceitemToContainer(pos [3]int32, slot uint8, itemName string, count uint8, itemData uint16, method string) error {
	request := fmt.Sprintf(
		"replaceitem block %d %d %d slot.container %d %s %d %d %v",
		pos[0],
		pos[1],
		pos[2],
		slot,
		itemName,
		count,
		itemData,
		method,
	)
	err := g.SendSettingsCommand(request, true)
	if err != nil {
		return fmt.Errorf("ReplaceitemToContainer: %v", err)
	}
	return nil
}
