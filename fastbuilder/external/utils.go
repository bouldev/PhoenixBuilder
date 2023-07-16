package external

import (
	GameInterface "phoenixbuilder/game_control/game_interface"
	_ "unsafe"

	"github.com/google/uuid"
)

//go:linkname sendCommand phoenixbuilder/game_control/game_interface.(*GameInterface).sendCommand
func sendCommand(
	g *GameInterface.GameInterface,
	command string,
	uniqueId uuid.UUID,
	origin uint32,
)
