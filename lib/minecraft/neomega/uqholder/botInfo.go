package uqholder

import (
	minecraft "fastbuilder-core/lib/minecraft/gophertunnel"
	"fastbuilder-core/lib/minecraft/neomega/omega"
	"fmt"
)

func init() {
	if false {
		func(omega.BotBasicInfoHolder) {}(&BotInfoHolder{})
		func(uq omega.PlayerUQ) {}(&BotInfoHolder{})
	}
}

type BotInfoHolder struct {
	debug        bool
	BotName      string
	BotRuntimeID uint64
	BotUniqueID  int64
}

func (b *BotInfoHolder) IsBot() bool {
	return true
}

func (b *BotInfoHolder) GetPlayerName() string {
	return b.GetBotName()
}

func (b *BotInfoHolder) GetBotName() string {
	return b.BotName
}

func (b *BotInfoHolder) GetBotRuntimeID() uint64 {
	return b.BotRuntimeID
}

func (b *BotInfoHolder) GetBotUniqueID() int64 {
	return b.BotUniqueID
}

func NewBotInfoHolder(conn *minecraft.Conn, debug bool) omega.BotBasicInfoHolder {
	h := &BotInfoHolder{
		debug: debug,
	}
	gd := conn.GameData()
	h.BotRuntimeID = gd.EntityRuntimeID
	h.BotUniqueID = gd.EntityUniqueID
	h.BotName = conn.IdentityData().DisplayName
	if h.debug {
		fmt.Printf("uqHolder.GetBotName()=%v\n", h.GetBotName())
		fmt.Printf("uqHolder.GetBotRuntimeID()=%v\n", h.GetBotRuntimeID())
		fmt.Printf("uqHolder.GetBotUniqueID()=%v\n", h.GetBotUniqueID())
	}
	return h
}
