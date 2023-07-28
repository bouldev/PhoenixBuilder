package GameInterface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
)

// 指定物品生成的位置。
// 仅被 ReplaceItemInInventory 函数所使用
type ItemGenerateLocation struct {
	// 指代物品应当在哪个库存生成，
	// 一个样例是 slot.weapon.mainhand
	Path string
	// 指代物品应当在哪个槽位上生成
	Slot uint8
}

// 向容器填充物品。
// chestSlot 指代该物品的基本信息，
// method 指代该物品的物品组件信息
func (g *GameInterface) ReplaceItemInContainer(
	pos [3]int32,
	chestSlot types.ChestSlot,
	method string,
) error {
	request := commands_generator.ReplaceItemInContainerRequest(
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

// 向背包填充物品。
// itemBasicData 指代该物品的基本信息，
// generateLocation 指代该物品的实际生成位置，
// method 指代该物品的物品组件信息
func (g *GameInterface) ReplaceItemInInventory(
	target string,
	generateLocation ItemGenerateLocation,
	itemBasicData types.ChestSlot,
	method string,
) error {
	request := commands_generator.ReplaceItemInInventoryRequest(
		&itemBasicData,
		target,
		fmt.Sprintf("%s %d", generateLocation.Path, generateLocation.Slot),
		method,
	)
	resp := g.SendWSCommandWithResponse(request)
	if resp.Error != nil {
		return fmt.Errorf("ReplaceitemToContainer: %v", resp.Error)
	}
	return nil
}
