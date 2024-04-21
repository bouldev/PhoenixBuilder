package ResourcesControl

import (
	"fmt"
	"phoenixbuilder/fastbuilder/py_rpc"
	stc "phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client"
	stc_mc "phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/pterm/pterm"
)

// 根据收到的数据包更新客户端的资源数据
func (r *Resources) handlePacket(pk *packet.Packet) {
	switch p := (*pk).(type) {
	case *packet.TickSync:
		r.Others.write_tick_sync_resp(*p)
		// sync game tick
	case *packet.CommandOutput:
		err := r.Command.try_to_write_response(p.CommandOrigin.UUID, *p)
		if err != nil {
			pterm.Error.Printf("handlePacket: %v\n", err)
		}
		// send ws command with response
	case *packet.PyRpc:
		if p.Value == nil {
			return
		}
		if p.Error != nil {
			panic(fmt.Sprintf("handlePacket: %v", p.Error))
		}
		// prepare
		content, err := py_rpc.Unmarshal(p.Value)
		if err != nil {
			pterm.Warning.Sprintf("handlePacket: %v", err)
			return
		}
		// unmarshal
		switch c := content.(type) {
		case *py_rpc.ModEvent:
			park, success := c.Package.(*stc.Minecraft)
			if !success {
				return
			}
			// minecraft package
			switch module := park.Module.(type) {
			case *stc_mc.AICommand:
				r.Command.on_ai_command(*module)
				// netease ai command
			}
		}
		// do some actions for some specific PyRpc packets
	case *packet.InventoryContent:
		for key, value := range p.Content {
			if value.Stack.ItemType.NetworkID != -1 {
				r.Inventory.write_item_stack_info(p.WindowID, uint8(key), value)
			}
		}
		// inventory contents(basic)
	case *packet.InventoryTransaction:
		for _, value := range p.Actions {
			if value.SourceType == protocol.InventoryActionSourceCreative {
				continue
			}
			r.Inventory.write_item_stack_info(uint32(value.WindowID), uint8(value.InventorySlot), value.NewItem)
		}
		// inventory contents(for enchant command...)
	case *packet.InventorySlot:
		r.Inventory.write_item_stack_info(p.WindowID, uint8(p.Slot), p.NewItem)
		// inventory contents(for chest...) [NOT TEST]
	case *packet.ItemStackResponse:
		for _, value := range p.Responses {
			if value.Status == protocol.ItemStackResponseStatusOK {
				r.ItemStackOperation.update_item_data(value, &r.Inventory)
			}
			// update local inventory data
			r.ItemStackOperation.write_response(value.RequestID, value)
			// write response
		}
		// item stack request
	case *packet.ContainerOpen:
		if !r.Container.GetOccupyStates() {
			panic("handlePacket: Attempt to send packet.ContainerOpen without using ResourcesControlCenter")
		}
		r.Container.write_container_closing_data(nil)
		r.Container.write_container_opening_data(p)
		r.Inventory.create_new_inventory(uint32(p.WindowID))
		r.Container.respond_to_container_operation()
		// when a container is opened
	case *packet.ContainerClose:
		if p.WindowID != 0 && p.WindowID != 119 && p.WindowID != 120 && p.WindowID != 124 {
			err := r.Inventory.delete_inventory(uint32(p.WindowID))
			if err != nil {
				panic(fmt.Sprintf("handlePacket: Try to removed an inventory which not existed; p.WindowID = %v", p.WindowID))
			}
		}
		if !p.ServerSide && !r.Container.GetOccupyStates() {
			panic("handlePacket: Attempt to send packet.ContainerClose without using ResourcesControlCenter")
		}
		r.Container.write_container_opening_data(nil)
		r.Container.write_container_closing_data(p)
		r.Container.respond_to_container_operation()
		// when a container has been closed
	case *packet.StructureTemplateDataResponse:
		if !r.Structure.GetOccupyStates() {
			panic("handlePacket: Attempt to send packet.StructureTemplateDataRequest without using ResourcesControlCenter")
		}
		r.Structure.writeResponse(*p)
		// used to request mcstructure data
	}
	// process packet
	r.Listener.distribute_packet(*pk)
	// distribute packet(for packet listener)
}
