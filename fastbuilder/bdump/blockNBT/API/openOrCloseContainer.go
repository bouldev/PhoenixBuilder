package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
)

/*
打开 pos 处名为 blockName 且方块状态为 blockStates 的容器，且只有当打开完成后才会返回值。
当 needOccupyContainerResources 为真时，此函数会主动占用容器资源，但一般情况下我建议此参数填 false ，
因为打开容器仅仅是一系列容器操作的一个步骤，因此此函数中不应该贸然修改容器资源，否则可能会造成潜在的问题。
除此外，此函数并不会检查打开容器时是否满足打开条件，因此您需要确保打开容器前从未打开容器或已关闭了上次打开的容器，
否则这可能会造成潜在的问题
*/
func (g *GlobalAPI) OpenContainer(
	pos [3]int32,
	blockName string,
	blockStates map[string]interface{},
	needOccupyContainerResources bool,
) error {
	var lock *sync.Mutex = &sync.Mutex{}
	if needOccupyContainerResources {
		_, lock = g.PacketHandleResult.ContainerResources.Occupy(false)
	}
	// lock down resources
	g.PacketHandleResult.ContainerResources.AwaitResponceBeforeSendPacket()
	// await responce before send packet
	err := g.UseItemOnBlocks(0, pos, blockName, blockStates, false)
	if err != nil {
		return fmt.Errorf("OpenContainer: %v", err)
	}
	// open container
	g.PacketHandleResult.ContainerResources.AwaitResponceAfterSendPacket()
	// wait changes
	if needOccupyContainerResources {
		lock.Unlock()
	}
	// unlock resources
	return nil
	// return
}

/*
关闭已经打开的容器，且只有当容器被关闭后才会返回值。当容器被成功关闭，公用容器资源会被强制释放，
因为关闭容器宣告着一系列的容器操作的结束。
此函数并不会检查此前是否已经打开过容器，因此您需要确保关闭容器前已经打开了一个容器，否则可能会造成潜在的问题
*/
func (g *GlobalAPI) CloseContainer() error {
	defer g.PacketHandleResult.ContainerResources.release()
	// release sharing resources
	g.PacketHandleResult.ContainerResources.AwaitResponceBeforeSendPacket()
	// await responce before send packet
	err := g.WritePacket(&packet.ContainerClose{
		WindowID:   g.PacketHandleResult.ContainerResources.GetContainerOpenDatas().WindowID,
		ServerSide: false,
	})
	if err != nil {
		return fmt.Errorf("CloseContainer: %v", err)
	}
	// close container
	g.PacketHandleResult.ContainerResources.AwaitResponceAfterSendPacket()
	// wait changes
	return nil
	// return
}
