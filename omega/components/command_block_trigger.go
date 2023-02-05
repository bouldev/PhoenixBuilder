package components

import (
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

type CBTrigger struct {
	*defines.BasicComponent
	TriggerPatterns []string    `json:"信息匹配规则"`
	TriggerHint     string      `json:"功能名"`
	CmdsIn          interface{} `json:"触发时执行指令"`
	DebugMode       bool        `json:"调试模式"`
	Cmds            []defines.Cmd

	CompiledTriggerChecker func(input string) (keys map[string]string, hit bool)
}

func (o *CBTrigger) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	var err error
	if o.Cmds, err = utils.ParseAdaptiveCmd(o.CmdsIn); err != nil {
		panic(err)
	}
}

func (o *CBTrigger) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetGameChatInterceptor(func(chat *defines.GameChat) (stop bool) {
		// is command block?
		if chat.Type != packet.TextTypeObject {
			return false
		}
		return false
	})

}
