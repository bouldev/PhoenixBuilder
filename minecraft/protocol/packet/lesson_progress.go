package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

const (
	LessonActionStart = iota
	LessonActionComplete
	LessonActionRestart
)

// LessonProgress is a packet sent by the server to the client to inform the client of updated progress on a lesson.
// This packet only functions on the Minecraft: Education Edition version of the game.
type LessonProgress struct {
	// Identifier is the identifier of the lesson that is being progressed.
	Identifier string

	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		Action is the action the client should perform to show progress. This is one of the constants defined above.

		For netease, the data type of this field is int32,
		but on standard minecraft, this is uint8.
	*/
	Action int32
	// Action uint8

	// Score is the score the client should use when displaying the progress.
	Score int32
}

// ID ...
func (*LessonProgress) ID() uint32 {
	return IDLessonProgress
}

func (pk *LessonProgress) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Varint32(&pk.Action)
		// io.Uint8(&pk.Action)
	}
	io.Varint32(&pk.Score)
	io.String(&pk.Identifier)
}
