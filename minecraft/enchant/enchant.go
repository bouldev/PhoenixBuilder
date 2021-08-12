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
		player:=configuration.RespondUser
		lineUUID,_:=uuid.NewUUID()
		lineChan:=make(chan *packet.CommandOutput) //　蓮(れん)ちゃん(x
		command.UUIDMap.Store(lineUUID.String(),lineChan)
		command.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~ ~400 ~",player,conn.IdentityData().DisplayName),lineUUID,conn)
		<-lineChan
		close(lineChan)
		AddPlayerItemChannel=make(chan *protocol.ItemStack)
		dispUUID,_:=uuid.NewUUID()
		command.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~ ~2 ~",player,conn.IdentityData().DisplayName),dispUUID,conn)
		item:=<-AddPlayerItemChannel
		close(AddPlayerItemChannel)
		fmt.Printf("%+v\n",item)
	} ()
}