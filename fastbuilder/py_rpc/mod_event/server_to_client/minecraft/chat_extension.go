package minecraft

import (
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft/chat_extension"
)

// 聊天扩展
type ChatExtension struct{ mei.Module }

// Return the module name of c
func (c *ChatExtension) ModuleName() string {
	return "chatExtension"
}

// Return a pool/map that contains all the event of c
func (c *ChatExtension) EventPool() map[string]mei.Event {
	return map[string]mei.Event{
		"PlayerAddRoom": &chat_extension.PlayerAddRoom{},
	}
}
