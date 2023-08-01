//go:build !is_tweak

package interfaces

import (
	resources_control "phoenixbuilder/game_control/resources_control"
)

type GameInterface interface {
	SendSettingsCommand(string, bool) error
	SendCommand(string) error
	SendWSCommand(string) error
	SendCommandWithResponse(string, resources_control.CommandRequestOptions) resources_control.CommandRespond
	SendWSCommandWithResponse(string, resources_control.CommandRequestOptions) resources_control.CommandRespond

	SetBlock([3]int32, string, string) error
	SetBlockAsync([3]int32, string, string) error

	SendChat(string) error
	Output(string) error
	Title(string) error
}
