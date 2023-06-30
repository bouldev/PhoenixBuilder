package GameInterface

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

/*
打开 pos 处名为 blockName 且方块状态为 blockStates 的容器，且只有当打开完成后才会返回值。
hotBarSlotID 字段代表玩家此时手持的物品栏，因为打开容器实际上是一次方块点击事件。
返回值的第一项代表执行结果，为真时容器被成功打开，否则反之。

请确保在使用此函数前占用了容器资源，否则会造成程序 panic
*/
func (g *GameInterface) OpenContainer(
	pos [3]int32,
	blockName string,
	blockStates map[string]interface{},
	hotBarSlotID uint8,
) (bool, error) {
	g.Resources.Container.AwaitChangesBeforeSendingPacket()
	// await responce before send packet
	err := g.ClickBlock(
		UseItemOnBlocks{
			HotbarSlotID: hotBarSlotID,
			BlockPos:     pos,
			BlockName:    blockName,
			BlockStates:  blockStates,
		},
	)
	if err != nil {
		return false, fmt.Errorf("OpenContainer: %v", err)
	}
	// open container
	g.Resources.Container.AwaitChangesAfterSendingPacket()
	// wait changes
	if g.Resources.Container.GetContainerOpeningData() == nil {
		return false, nil
	}
	// if unsuccess
	return true, nil
	// return
}

/*
关闭已经打开的容器，且只有当容器被关闭后才会返回值。
您应该确保容器被关闭后，对应的容器公用资源被释放。

返回值的第一项代表执行结果，为真时容器被成功关闭，否则反之
*/
func (g *GameInterface) CloseContainer() (bool, error) {
	g.Resources.Container.AwaitChangesBeforeSendingPacket()
	// await responce before send packet
	if g.Resources.Container.GetContainerOpeningData() == nil {
		return false, ErrContainerNerverOpened
	}
	// if the container have been nerver opened
	err := g.WritePacket(&packet.ContainerClose{
		WindowID:   g.Resources.Container.GetContainerOpeningData().WindowID,
		ServerSide: false,
	})
	if err != nil {
		return false, fmt.Errorf("CloseContainer: %v", err)
	}
	// close container
	g.Resources.Container.AwaitChangesAfterSendingPacket()
	// wait changes
	if g.Resources.Container.GetContainerClosingData() == nil {
		return false, nil
	}
	// if unsuccess
	return true, nil
	// return
}
