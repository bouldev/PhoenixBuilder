package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

/*
打开 pos 处名为 blockName 且方块状态为 blockStates 的容器，且只有当打开完成后才会返回值。
hotBarSlotID 字段代表玩家此时手持的物品栏，因为打开容器实际上是一次方块点击事件。
返回值的第一项代表执行结果，为真时容器被成功打开，否则反之。

请确保在使用此函数前占用了容器资源，否则会造成程序惊慌。
*/
func (g *GlobalAPI) OpenContainer(
	pos [3]int32,
	blockName string,
	blockStates map[string]interface{},
	hotBarSlotID uint8,
) (bool, error) {
	g.Resources.Container.AwaitChangesBeforeSendPacket()
	// await responce before send packet
	err := g.ClickBlock(hotBarSlotID, pos, blockName, blockStates, false)
	if err != nil {
		return false, fmt.Errorf("OpenContainer: %v", err)
	}
	// open container
	g.Resources.Container.AwaitChangesAfterSendPacket()
	// wait changes
	if g.Resources.Container.GetContainerOpenDatas() == nil {
		return false, nil
	}
	// if unsuccess
	return true, nil
	// return
}

/*
打开背包。
返回值的第一项代表执行结果，为真时背包被成功打开，否则反之。

请确保打开前占用了容器资源，否则会造成程序惊慌。
*/
func (g *GlobalAPI) OpenInventory() (bool, error) {
	g.Resources.Container.AwaitChangesBeforeSendPacket()
	// await responce before send packet
	err := g.WritePacket(&packet.Interact{
		ActionType:            packet.InteractActionOpenInventory,
		TargetEntityRuntimeID: g.BotInfo.BotRunTimeID,
	})
	if err != nil {
		return false, fmt.Errorf("OpenInventory: %v", err)
	}
	// open inventory
	g.Resources.Container.AwaitChangesAfterSendPacket()
	// wait changes
	if g.Resources.Container.GetContainerOpenDatas() == nil {
		return false, nil
	}
	// if unsuccess
	return true, nil
	// return
}

// 用于关闭容器时检测到容器从未被打开时的报错信息
var ErrContainerNerverOpened error = fmt.Errorf("CloseContainer: Container have been nerver opened")

/*
关闭已经打开的容器，且只有当容器被关闭后才会返回值。
您应该确保容器被关闭后，对应的容器公用资源被释放。

返回值的第一项代表执行结果，为真时容器被成功关闭，否则反之
*/
func (g *GlobalAPI) CloseContainer() (bool, error) {
	g.Resources.Container.AwaitChangesBeforeSendPacket()
	// await responce before send packet
	if g.Resources.Container.GetContainerOpenDatas() == nil {
		return false, ErrContainerNerverOpened
	}
	// if the container have been nerver opened
	err := g.WritePacket(&packet.ContainerClose{
		WindowID:   g.Resources.Container.GetContainerOpenDatas().WindowID,
		ServerSide: false,
	})
	if err != nil {
		return false, fmt.Errorf("CloseContainer: %v", err)
	}
	// close container
	g.Resources.Container.AwaitChangesAfterSendPacket()
	// wait changes
	if g.Resources.Container.GetContainerCloseDatas() == nil {
		return false, nil
	}
	// if unsuccess
	return true, nil
	// return
}
