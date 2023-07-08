//go:build !is_tweak

package interfaces

import (
	game_interface "phoenixbuilder/game_control/game_interface"
)

type GameInterface interface {
	SendSettingsCommand(string, bool) error
	SendCommand(string) error
	SendWSCommand(string) error
	SendCommandWithResponse(string) game_interface.CommandRespond
	SendWSCommandWithResponse(string) game_interface.CommandRespond

	SetBlock([3]int32, string, string) error
	SetBlockAsync([3]int32, string, string) error

	SendChat(string) error
	Output(string) error
	Title(string) error
}
