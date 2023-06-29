package GameInterface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
)

// 向容器填充物品
func (g *GameInterface) ReplaceItemInContainer(
	pos [3]int32,
	chestSlot types.ChestSlot,
	method string,
) error {
	request := commands_generator.ReplaceItemRequest(
		&types.Module{
			Point: types.Position{
				X: int(pos[0]),
				Y: int(pos[1]),
				Z: int(pos[2]),
			},
			ChestSlot: &chestSlot,
		},
		method,
	)
	err := g.SendSettingsCommand(request, true)
	if err != nil {
		return fmt.Errorf("ReplaceitemToContainer: %v", err)
	}
	return nil
}
