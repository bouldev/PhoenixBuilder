package GameInterface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
)

// 在 pos 处以 setblock 命令放置名为 name 且方块状态为 states 的方块。
// 此实现是阻塞的，它将等待租赁服回应后再返回值
func (g *GameInterface) SetBlock(pos [3]int32, name string, states string) error {
	request := commands_generator.SetBlockRequest(&types.Module{
		Block: &types.Block{
			Name:        &name,
			BlockStates: states,
		},
		Point: types.Position{
			X: int(pos[0]),
			Y: int(pos[1]),
			Z: int(pos[2]),
		},
	}, &types.MainConfig{})
	// get setblock command
	resp := g.SendWSCommandWithResponse(
		request,
		ResourcesControl.CommandRequestOptions{
			TimeOut: ResourcesControl.CommandRequestDefaultDeadLine,
		},
	)
	if resp.Error != nil && resp.ErrorType == ResourcesControl.ErrCommandRequestTimeOut {
		err := g.SendSettingsCommand(request, true)
		if err != nil {
			return fmt.Errorf("SetBlock: %v", err)
		}
		err = g.AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf("SetBlock: %v", err)
		}
		return nil
	}
	if resp.Error != nil {
		return fmt.Errorf("SetBlock: %v", resp.Error)
	}
	// send setblock request
	return nil
	// return
}

// 在 pos 处以 setblock 命令放置名为 name 且方块状态为 states 的方块。
// 此实现不会等待租赁服响应，数据包被发送后将立即返回值
func (g *GameInterface) SetBlockAsync(pos [3]int32, name string, states string) error {
	request := commands_generator.SetBlockRequest(&types.Module{
		Block: &types.Block{
			Name:        &name,
			BlockStates: states,
		},
		Point: types.Position{
			X: int(pos[0]),
			Y: int(pos[1]),
			Z: int(pos[2]),
		},
	}, &types.MainConfig{})
	err := g.SendSettingsCommand(request, true)
	if err != nil {
		return fmt.Errorf("SetBlockAsync: %v", err)
	}
	return nil
}
