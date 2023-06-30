// +build !is_tweak

package interfaces

import "phoenixbuilder/minecraft/protocol/packet"

type GameInterface interface {
	SendSettingsCommand(string, bool) error
	SendCommand(string) error
	SendWSCommand(string) error
	SendCommandWithResponse(string) (packet.CommandOutput, error)
	SendWSCommandWithResponse(string) (packet.CommandOutput, error)
	
	SetBlock([3]int32,string,string) error
	SetBlockAsync([3]int32,string,string) error
	
	SendChat(string) error
	Output(string) error
	Title(string) error
}