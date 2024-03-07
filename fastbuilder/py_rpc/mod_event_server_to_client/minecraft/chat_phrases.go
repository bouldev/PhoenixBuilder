package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/interface/mod_event"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/minecraft/chat_phrases"
)

// 网易提供的快捷游戏短语
type ChatPhrases struct{ mod_event.Module }

// Return the module name of c
func (c *ChatPhrases) ModuleName() string {
	return "chatPhrases"
}

// Return a pool/map that contains all the event of c
func (c *ChatPhrases) EventPool() map[string]mod_event.Event {
	return map[string]mod_event.Event{
		"SyncNewPlayerPhrasesData": &chat_phrases.SyncNewPlayerPhrasesData{},
	}
}
