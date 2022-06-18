// +build is_tweak

package commands

import (
	"github.com/google/uuid"
	"sync"
)

/*
// Callee free
void phoenixbuilder_send_silent_command(char *command);
void phoenixbuilder_send_ws_command(char *command, char *uuid);
void phoenixbuilder_send_command(char *command, char *uuid);
void phoenixbuilder_send_chat(char *content);
*/
import "C"

func (sender *CommandSender) GetBlockUpdateSubscribeMap() *sync.Map {
	return &sender.BlockUpdateSubscribeMap
}

func (sender *CommandSender) GetUUIDMap() *sync.Map {
	return &sender.UUIDMap
}

func (sender *CommandSender) ClearUUIDMap() {
	sender.UUIDMap=sync.Map{}
}

func (sender *CommandSender) SendCommand(command string, UUID uuid.UUID) error {
	C.phoenixbuilder_send_command(C.CString(command), C.CString(UUID.String()))
	return nil
}

func (sender *CommandSender) SendWSCommand(command string, UUID uuid.UUID) error {
	C.phoenixbuilder_send_ws_command(C.CString(command), C.CString(UUID.String()))
	return nil
}

func (sender *CommandSender) SendSizukanaCommand(command string) error {
	C.phoenixbuilder_send_silent_command(C.CString(command))
	return nil
}

func (sender *CommandSender) SendChat(content string) error {
	C.phoenixbuilder_send_chat(C.CString(content))
	return nil
}