package nbtconstructor

import (
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/command"
	"github.com/google/uuid"
	"fmt"
	"time"
)

var isOccupied bool = false
var AddVillagerChannel chan *packet.AddActor = nil
var InventoryContentChannel chan *packet.InventoryContent = nil
// ^ WindowID=0 only
var ItemStackResponseChannel chan *packet.ItemStackResponse = nil
//var TradeWindowIDChannel chan byte = nil
var IsWorking bool = false
var TradeWindowID byte

func StartSessionWithCustomNBT(conn *minecraft.Conn, itemNetworkID int32, metadataValue uint32, nbtContent map[string]interface{}) {
	if(isOccupied) {
		command.Tellraw(conn, "There's already a working nbt construction session.")
		return
	}
	isOccupied=true
	reqID:=int32(-3)
	InventoryContentChannel=make(chan *packet.InventoryContent)
	AddVillagerChannel=make(chan *packet.AddActor)
	command.SendWSCommand("summon villager ~ ~ ~ minecraft:become_butcher", uuid.New(), conn)
	villagerPkt:=<-AddVillagerChannel
	close(AddVillagerChannel)
	AddVillagerChannel=nil
	IsWorking=true
	err:=conn.WritePacket(&packet.InventoryTransaction{
		LegacyRequestID:    0,
		LegacySetItemSlots: []protocol.LegacySetItemSlot{},
		Actions:            []protocol.InventoryAction{},
		TransactionData: &protocol.UseItemOnEntityTransactionData{
			TargetEntityRuntimeID: villagerPkt.EntityRuntimeID,
			ActionType:            0,
			HotBarSlot:            0,
			HeldItem: protocol.ItemInstance{
				StackNetworkID: 0,
				Stack: protocol.ItemStack{
					ItemType: protocol.ItemType{
						NetworkID:     0,
						MetadataValue: 0,
					},
					BlockRuntimeID: 0,
					Count:          0,
					NBTData:        map[string]interface{}{},
					CanBePlacedOn:  []string{},
					CanBreak:       []string{},
					HasNetworkID:   false,
				},
			},
			Position:        villagerPkt.Position,
			ClickedPosition: villagerPkt.Position,
		},
	})
	if err!=nil {
		panic(err)
	}
	command.SendWSCommand("clear @s", uuid.New(), conn)
	<-InventoryContentChannel
	command.SendWSCommand("give @s emerald 64", uuid.New(), conn)
	content:=<-InventoryContentChannel
	emeraldStackID:=content.Content[0].StackNetworkID
	if emeraldStackID == 0 {
		isOccupied=false
		command.Tellraw(conn, "Failed to get the emerald stack.")
		return
	}
	close(InventoryContentChannel)
	InventoryContentChannel=nil
	ItemStackResponseChannel=make(chan *packet.ItemStackResponse)
	placeReq:=&packet.ItemStackRequest {
		Requests: []protocol.ItemStackRequest {
			{
				RequestID: reqID,
				Actions: []protocol.StackRequestAction {
					&protocol.PlaceStackRequestAction {},
				},
			},
		},
	}
	placeReq.Requests[0].Actions[0].(*protocol.PlaceStackRequestAction).Count=64
	placeReq.Requests[0].Actions[0].(*protocol.PlaceStackRequestAction).Source=protocol.StackRequestSlotInfo {
		ContainerID: 12,
		Slot: 0,
		StackNetworkID:emeraldStackID,
	}
	placeReq.Requests[0].Actions[0].(*protocol.PlaceStackRequestAction).Destination=protocol.StackRequestSlotInfo {
		ContainerID: 47,
		Slot: 4,
		StackNetworkID: 0,
	}
	time.Sleep(time.Second)
	conn.WritePacket(placeReq)
	resp:=<-ItemStackResponseChannel
	if resp.Responses[0].Status != 0 {
		isOccupied=false
		command.Tellraw(conn, fmt.Sprintf("Unexpected item stack response: %+v",resp.Responses[0]))
		return
	}
	nbtdatar:=nbtContent
	tradereq:=&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				RequestID: reqID - 2,
				Actions: []protocol.StackRequestAction{
					&protocol.CraftRecipeStackRequestAction{
						RecipeNetworkID: 1145,
					},
					&protocol.CraftResultsDeprecatedStackRequestAction{
						ResultItems: []protocol.ItemStack{
							{
								ItemType: protocol.ItemType{
									NetworkID:     itemNetworkID,
									MetadataValue: metadataValue,
								},
								BlockRuntimeID: 0,
								Count:          1,
								NBTData: nbtdatar,
								CanBePlacedOn: []string{},
								CanBreak:      []string{},
								HasNetworkID:  false,
							},
						},
						TimesCrafted: 1,
					},
					&protocol.ConsumeStackRequestAction{
						DestroyStackRequestAction: protocol.DestroyStackRequestAction{
							Count: 64,
							Source: protocol.StackRequestSlotInfo{
								ContainerID:    47,
								Slot:           4,
								StackNetworkID: emeraldStackID,
							},
						},
					},
					&protocol.PlaceStackRequestAction{},
				},
			},
		},
	}
	tradereq.Requests[0].Actions[3].(*protocol.PlaceStackRequestAction).Count = 1
	tradereq.Requests[0].Actions[3].(*protocol.PlaceStackRequestAction).Source = protocol.StackRequestSlotInfo{
		ContainerID:    60,
		Slot:           50,
		StackNetworkID: reqID - 2,
	}
	tradereq.Requests[0].Actions[3].(*protocol.PlaceStackRequestAction).Destination = protocol.StackRequestSlotInfo{
		ContainerID:    12,
		Slot:           0,
		StackNetworkID: 0,
	}
	command.Tellraw(conn, "Fetching item")
	conn.WritePacket(tradereq)
	resp=<-ItemStackResponseChannel
	if resp.Responses[0].Status != 0 {
		isOccupied=false
		command.Tellraw(conn, fmt.Sprintf("[2] Unexpected item stack response: %+v",resp.Responses[0]))
		return
	}
	dropItem := &packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				RequestID: reqID - 4,
				Actions: []protocol.StackRequestAction{
					&protocol.DropStackRequestAction{
						Count: 1,
						Source: protocol.StackRequestSlotInfo{
							ContainerID:    12,
							Slot:           0,
							StackNetworkID: emeraldStackID + 1,
						},
						Randomly: false,
					},
				},
				CustomNames: nil,
			},
		},
	}
	command.Tellraw(conn, "Dropping the item")
	conn.WritePacket(dropItem)
	resp=<-ItemStackResponseChannel
	if resp.Responses[0].Status != 0 {
		isOccupied=false
		command.Tellraw(conn, fmt.Sprintf("[3] Unexpected item stack response: %+v",resp.Responses[0]))
		return
	}
	close(ItemStackResponseChannel)
	ItemStackResponseChannel=nil
	IsWorking=false
	conn.WritePacket(&packet.ContainerClose {
		WindowID: TradeWindowID,
		ServerSide: false,
	})
	isOccupied=false
	command.Tellraw(conn, "[NBTConstructor] Process finished.")
}