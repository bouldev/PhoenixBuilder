package listen

import (
	"context"
	"phoenixbuilder/minecraft/protocol/packet"
)

type LuaGoListen interface {
	UserInputChan() <-chan string
	MakeMCPacketFeeder(ctx context.Context, wants []string) <-chan packet.Packet
}
