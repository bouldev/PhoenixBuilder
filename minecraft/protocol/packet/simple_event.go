package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

const (
	SimpleEventCommandsEnabled = iota + 1
	SimpleEventCommandsDisabled
	SimpleEventUnlockWorldTemplateSettings
)

// SimpleEvent is sent by the server to send a 'simple event' to the client, meaning an event without any
// additional event data. The event is typically used by the client for telemetry.
type SimpleEvent struct {
	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		EventType is the type of the event to be called. It is one of the constants that may be found above.

		For netease, the data type of this field is uint16,
		but on standard minecraft, this is int16.
	*/
	EventType uint16
	// EventType int16
}

// ID ...
func (*SimpleEvent) ID() uint32 {
	return IDSimpleEvent
}

func (pk *SimpleEvent) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Uint16(&pk.EventType)
		// io.Int16(&pk.EventType)
	}
}
