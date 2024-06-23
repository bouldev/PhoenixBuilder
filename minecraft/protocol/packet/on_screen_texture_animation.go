package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// OnScreenTextureAnimation is sent by the server to show a certain animation on the screen of the player.
// The packet is used, as an example, for when a raid is triggered and when a raid is defeated.
type OnScreenTextureAnimation struct {

	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		AnimationType is the type of the animation to show. The packet provides no further extra data to allow
		modifying the duration or other properties of the animation.s

		For netease, the data type of this field is uint32,
		but on standard minecraft, this is int32.
	*/
	AnimationType uint32
	// AnimationType int32
}

// ID ...
func (*OnScreenTextureAnimation) ID() uint32 {
	return IDOnScreenTextureAnimation
}

func (pk *OnScreenTextureAnimation) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		io.Uint32(&pk.AnimationType)
		// io.Int32(&pk.AnimationType)
	}
}
