package enchant

import (
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/command"
	"github.com/google/uuid"
	"strconv"
	"fmt"
	"strings"
	"time"
)

var isOccupied bool = false
var AddVillagerChannel chan *packet.AddActor = nil
var InventoryContentChannel chan *packet.InventoryContent = nil
// ^ WindowID=0 only
var ItemStackResponseChannel chan *packet.ItemStackResponse = nil
//var TradeWindowIDChannel chan byte = nil
var PacketToResend *packet.InventoryTransaction = nil
var TradeWindowID byte

func StartSession(conn *minecraft.Conn) {
	if(isOccupied) {
		command.Tellraw(conn, "There's already a started enchant session.")
		return
	}
	isOccupied=true
	reqID:=int32(-3)
	command.SendWSCommand("clear @s", uuid.New(), conn)
	AddVillagerChannel=make(chan *packet.AddActor)
	command.SendWSCommand("summon villager ~ ~ ~ minecraft:become_butcher", uuid.New(), conn)
	villagerPkt:=<-AddVillagerChannel
	close(AddVillagerChannel)
	AddVillagerChannel=nil
	//TradeWindowIDChannel=make(chan byte)
	PacketToResend=&packet.InventoryTransaction{
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
	}
	err:=conn.WritePacket(PacketToResend)
	if err!=nil {
		panic(err)
	}
	//command.Tellraw(conn, fmt.Sprintf("Start trading, window id: %d",int(tradeWindowID)))
	//close(TradeWindowIDChannel)
	//TradeWindowIDChannel=nil
	InventoryContentChannel=make(chan *packet.InventoryContent)
	command.Tellraw(conn, "Now drop me the item to enchant.")
	content:=<-InventoryContentChannel
	enchantItemID:=content.Content[0].Stack.ItemType.NetworkID
	enchantItemData:=content.Content[0].Stack.ItemType.MetadataValue
	finalEnch:=make([]map[string]interface{},0)
	finalLore:=make([]string,0)
	command.Tellraw(conn, fmt.Sprintf("Got it. Its NetworkID is %d while its metadata value is %d.",enchantItemID,enchantItemData))
	command.SendWSCommand("clear @s", uuid.New(), conn)
	<-InventoryContentChannel
	for {
		command.Tellraw(conn, "Please drop me enchant books with enchant level as their name now (e.g. 3a2a7a6a7), to assign a lore, drop me an unenchanted item with the content of the lore, to stop, drop me an unenchanted item w/o name.")
		content=<-InventoryContentChannel
		//fmt.Printf("%+v",content.Content[0])
		if len(content.Content[0].Stack.NBTData)==0 {
			command.Tellraw(conn, "OK, now I'm making it, please wait a while.")
			break
		}
		nbtdata:=content.Content[0].Stack.NBTData
		level:=int64(70000)
		displayField, hasDisplayField:=nbtdata["display"]
		enchField, hasEnchField:=nbtdata["ench"]
		if hasDisplayField {
			displayRI, ok := displayField.(map[string]interface{})
			if ok {
				n,h:=displayRI["Name"]
				if h {
					lvl,tr:=n.(string)
					if tr {
						if hasEnchField {
							level, err=strconv.ParseInt(strings.Replace(lvl,"a","",-1), 10, 64)
							if err != nil || level>32767 || level< -32767 {
								command.Tellraw(conn, "Invalid level assigned for this item.")
							}
						}else{
							finalLore=append(finalLore,lvl)
							command.Tellraw(conn, fmt.Sprintf("Assigned Lore as %+v.",finalLore))
						}
					}
				}
			}
		}
		if hasEnchField {
			enchF:=enchField.([]interface{})
			for _, item := range enchF {
				itmmap:=item.(map[string]interface{})
				enchId:=itmmap["id"].(int16)
				enchLvl:=int16(level)
				if level==70000 {
					enchLvl=itmmap["lvl"].(int16)
				}
				finalEnch=append(finalEnch,map[string]interface{}{
					"id": enchId,
					"lvl": enchLvl,
				})
				command.Tellraw(conn, fmt.Sprintf("Applied enchant: ID: %d, Level: %d.",enchId,enchLvl))
			}
		}
		command.SendWSCommand("clear @s", uuid.New(), conn)
		<-InventoryContentChannel
	}
	command.SendWSCommand("clear @s", uuid.New(), conn)
	<-InventoryContentChannel
	command.SendWSCommand("give @s emerald 64", uuid.New(), conn)
	content=<-InventoryContentChannel
	emeraldStackID:=content.Content[0].StackNetworkID
	if emeraldStackID == 0 {
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
		command.Tellraw(conn, fmt.Sprintf("Unexpected item stack response: %+v",resp.Responses[0]))
		return
	}
	nbtdatar:=make(map[string]interface{})
	if len(finalLore) != 0 {
		nbtdatar["display"]=map[string]interface{}{
			"Lore": finalLore,
		}
	}
	if len(finalEnch) !=0 {
		nbtdatar["ench"]=finalEnch
	}
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
									NetworkID:     enchantItemID, // the network id of red stone
									MetadataValue: enchantItemData,
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
		command.Tellraw(conn, fmt.Sprintf("[3] Unexpected item stack response: %+v",resp.Responses[0]))
		return
	}
	close(ItemStackResponseChannel)
	ItemStackResponseChannel=nil
	PacketToResend=nil
	conn.WritePacket(&packet.ContainerClose {
		WindowID: TradeWindowID,
		ServerSide: false,
	})
	isOccupied=false
	command.Tellraw(conn, "[Enchant] Process finished.")
}