package blockNBT_API

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/pterm/pterm"
)

// 解析数据包并执行相应动作，如更新记录的背包数据等
func (o *PacketHandleResult) HandlePacket(pk *packet.Packet) {
	switch p := (*pk).(type) {
	case *packet.CommandOutput:
		uniqueId := p.CommandOrigin.UUID
		ok := o.commandDatas.testRequest(uniqueId)
		if !ok {
			return
		}
		o.commandDatas.writeResponce(uniqueId, *p)
		// send ws command with responce
	case *packet.InventoryContent:
		for key, value := range p.Content {
			if value.Stack.ItemType.NetworkID != -1 {
				o.Inventory.writeItemStackInfo(p.WindowID, uint8(key), value)
			}
		}
		// inventory contents(global)
	case *packet.InventoryTransaction:
		for _, value := range p.Actions {
			if value.SourceType == protocol.InventoryActionSourceCreative {
				continue
			}
			o.Inventory.writeItemStackInfo(uint32(value.WindowID), uint8(value.InventorySlot), value.NewItem)
		}
		// inventory contents(for enchant command...)
	case *packet.ItemStackResponse:
		for _, value := range p.Responses {
			err := o.ItemStackOperation.writeResponce(value.RequestID, value)
			if err != nil {
				panic("HandlePacket: Attempt to send packet.ItemStackRequest without using Bdump/blockNBT API")
			}
		}
		// item stack request
	case *packet.ContainerOpen:
		unsuccess, _ := o.ContainerResources.Occupy(true)
		if unsuccess {
			panic("HandlePacket: Attempt to send packet.ContainerOpen without using Bdump/blockNBT API")
		}
		o.ContainerResources.writeContainerCloseDatas(packet.ContainerClose{})
		o.ContainerResources.writeContainerOpenDatas(*p)
		o.ContainerResources.releaseAwaitGoRoutine()
		// while open a container
	case *packet.ContainerClose:
		if p.WindowID != 0 && p.WindowID != 119 && p.WindowID != 120 && p.WindowID != 124 {
			err := o.Inventory.deleteInventory(uint32(p.WindowID))
			if err != nil {
				pterm.Warning.Printf("HandlePacket: Try to removed an inventory which not existed; p.WindowID = %v\n", p.WindowID)
			}
		}
		if p.ServerSide == false {
			unsuccess, _ := o.ContainerResources.Occupy(true)
			if unsuccess {
				panic("HandlePacket: Attempt to send packet.ContainerClose without using Bdump/blockNBT API")
			}
		} else {
			o.ContainerResources.release()
		}
		o.ContainerResources.writeContainerOpenDatas(packet.ContainerOpen{})
		o.ContainerResources.writeContainerCloseDatas(*p)
		o.ContainerResources.releaseAwaitGoRoutine()
		// while a container is closed
	}
}
