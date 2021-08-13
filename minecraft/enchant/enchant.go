package enchant

import (
	"fmt"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/minecraft/command"
	"phoenixbuilder/minecraft/configuration"
	"github.com/google/uuid"
)

var AddPlayerItemChannel chan *protocol.ItemStack = nil

func Run(conn *minecraft.Conn) {
	go func() {
		pos:=configuration.GlobalFullConfig().Main().Position
		conn.WritePacket(&packet.StructureBlockUpdate {
			Position: protocol.BlockPos { int32(pos.X), int32(pos.Y), int32(pos.Z) },
			StructureName: "mystructure:testa",
			StructureBlockType: packet.StructureBlockData,
			DataField: "igloo",
			Settings: protocol.StructureSettings {
				PaletteName: "default",
				Integrity: 100,
				//Size: protocol.BlockPos { 5, 5, 5 },
			},
			ShouldTrigger: true,
		})
		return
		player:=configuration.RespondUser
		lineUUID,_:=uuid.NewUUID()
		lineChan:=make(chan *packet.CommandOutput) //　蓮(れん)ちゃん(x
		command.UUIDMap.Store(lineUUID.String(),lineChan)
		command.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~400 ~ ~",player,conn.IdentityData().DisplayName),lineUUID,conn)
		<-lineChan
		close(lineChan)
		AddPlayerItemChannel=make(chan *protocol.ItemStack)
		dispUUID,_:=uuid.NewUUID()
		command.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~ ~2 ~",player,conn.IdentityData().DisplayName),dispUUID,conn)
		item:=<-AddPlayerItemChannel
		close(AddPlayerItemChannel)
		invact:=protocol.InventoryAction {
			SourceType: protocol.InventoryActionSourceWorld,
			WindowID: protocol.WindowIDOffHand,
			SourceFlags: 0,
			InventorySlot: 0,
			OldItem: protocol.ItemStack {
				ItemType: protocol.ItemType {
					NetworkID: 0,
				},
			},
			NewItem: *item,
			StackNetworkID: 0,
		}
		conn.WritePacket(&packet.InventoryTransaction {
			HasNetworkIDs: false,
			Actions: []protocol.InventoryAction { invact },
		})
		fmt.Printf("ok\n")
	} ()
}