package ResourcesControl

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/pterm/pterm"
)

// 根据收到的数据包更新客户端的资源数据
func (r *Resources) handlePacket(pk *packet.Packet) {
	switch p := (*pk).(type) {
	case *packet.TickSync:
		r.Others.writeTickSyncPacketResponse(*p)
		// sync game tick
	case *packet.CommandOutput:
		err := r.Command.tryToWriteResponse(p.CommandOrigin.UUID, *p)
		if err != nil {
			pterm.Error.Printf("handlePacket: %v\n", err)
		}
		// send ws command with response
	case *packet.InventoryContent:
		for key, value := range p.Content {
			if value.Stack.ItemType.NetworkID != -1 {
				r.Inventory.writeItemStackInfo(p.WindowID, uint8(key), value)
			}
		}
		// inventory contents(basic)
	case *packet.InventoryTransaction:
		for _, value := range p.Actions {
			if value.SourceType == protocol.InventoryActionSourceCreative {
				continue
			}
			r.Inventory.writeItemStackInfo(uint32(value.WindowID), uint8(value.InventorySlot), value.NewItem)
		}
		// inventory contents(for enchant command...)
	case *packet.InventorySlot:
		r.Inventory.writeItemStackInfo(p.WindowID, uint8(p.Slot), p.NewItem)
		// inventory contents(for chest...) [NOT TEST]
	case *packet.ItemStackResponse:
		for _, value := range p.Responses {
			if value.Status == protocol.ItemStackResponseStatusOK {
				r.ItemStackOperation.updateItemData(value, &r.Inventory)
			}
			// update local inventory data
			r.ItemStackOperation.writeResponse(value.RequestID, value)
			// write response
		}
		// item stack request
	case *packet.ContainerOpen:
		if !r.Container.GetOccupyStates() {
			panic("handlePacket: Attempt to send packet.ContainerOpen without using ResourcesControlCenter")
		}
		r.Container.writeContainerClosingData(nil)
		r.Container.writeContainerOpeningData(p)
		r.Inventory.createNewInventory(uint32(p.WindowID))
		r.Container.respondToContainerOperation()
		// when a container is opened
	case *packet.ContainerClose:
		if p.WindowID != 0 && p.WindowID != 119 && p.WindowID != 120 && p.WindowID != 124 {
			err := r.Inventory.deleteInventory(uint32(p.WindowID))
			if err != nil {
				panic(fmt.Sprintf("handlePacket: Try to removed an inventory which not existed; p.WindowID = %v", p.WindowID))
			}
		}
		if !p.ServerSide && !r.Container.GetOccupyStates() {
			panic("handlePacket: Attempt to send packet.ContainerClose without using ResourcesControlCenter")
		}
		r.Container.writeContainerOpeningData(nil)
		r.Container.writeContainerClosingData(p)
		r.Container.respondToContainerOperation()
		// when a container has been closed
	case *packet.StructureTemplateDataResponse:
		if !r.Structure.GetOccupyStates() {
			panic("handlePacket: Attempt to send packet.StructureTemplateDataRequest without using ResourcesControlCenter")
		}
		r.Structure.writeResponse(*p)
		// used to request mcstructure data
	}
	// process packet
	err := r.Listener.distributePacket(*pk)
	if err != nil {
		panic(fmt.Sprintf("handlePacket: %v", err))
	}
	// distribute packet(for packet listener)
}
