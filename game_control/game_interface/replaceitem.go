package GameInterface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
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

/*
向 pos 处的容器填充物品。

chestSlot 指代该物品的基本信息，
method 指代该物品的物品组件信息。

此实现不会等待租赁服响应，
数据包被发送后将立即返回值
*/
func (g *GameInterface) ReplaceItemInContainerAsync(
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
		return fmt.Errorf("ReplaceitemToContainerAsync: %v", err)
	}
	return nil
}

/*
向背包填充物品。

target 指代被填充物品的目标，是一个目标选择器；
itemBasicData 指代该物品的基本信息；
generateLocation 指代该物品的实际生成位置；
method 指代该物品的物品组件信息；

blocked 指代是否使用以阻塞的方式运行此函数，
如果为真，它将等待租赁服响应后再返回值
*/
func (g *GameInterface) ReplaceItemInInventory(
	target string,
	generateLocation ItemGenerateLocation,
	itemBasicData types.ChestSlot,
	method string,
	blocked bool,
) error {
	request := commands_generator.ReplaceItemInInventoryRequest(
		&itemBasicData,
		target,
		fmt.Sprintf("%s %d", generateLocation.Path, generateLocation.Slot),
		method,
	)
	// generate replaceitem request
	if blocked {
		resp := g.SendWSCommandWithResponse(
			request,
			ResourcesControl.CommandRequestOptions{
				TimeOut: ResourcesControl.CommandRequestDefaultDeadLine,
			},
		)
		if resp.Error != nil && resp.ErrorType == ResourcesControl.ErrCommandRequestTimeOut {
			err := g.SendSettingsCommand(request, true)
			if err != nil {
				return fmt.Errorf("ReplaceitemToContainer: %v", err)
			}
			err = g.AwaitChangesGeneral()
			if err != nil {
				return fmt.Errorf("ReplaceitemToContainer: %v", err)
			}
		}
		return nil
	}
	// if need to wait response
	err := g.SendSettingsCommand(request, true)
	if err != nil {
		return fmt.Errorf("ReplaceitemToContainer: %v", err)
	}
	// if there is no need to wait response
	return nil
	// return
}
