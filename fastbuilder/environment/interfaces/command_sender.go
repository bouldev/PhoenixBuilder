package interfaces

import (
	"github.com/google/uuid"
	"sync"
)

type CommandSender interface {
	GetUUIDMap() *sync.Map
	ClearUUIDMap()
	GetBlockUpdateSubscribeMap() *sync.Map
	SendCommand(string,uuid.UUID) error
	SendWSCommand(string,uuid.UUID) error
	SendSizukanaCommand(string) error
	SendChat(string) error
	Output(string) error
	WorldChatOutput(string, string) error
	Title(string) error
}