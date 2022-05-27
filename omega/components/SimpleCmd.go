package components

import (
	"encoding/json"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

type SimpleCmd struct {
	*BasicComponent
	Triggers     []string `json:"触发词"`
	FinalTrigger bool     `json:"总是触发"`
	ArgumentHint string   `json:"参数描述"`
	Usage        string   `json:"功能描述"`
	Cmds         []string `json:"触发时执行指令"`
}

func (sc *SimpleCmd) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, sc); err != nil {
		panic(err)
	}
	if sc.Cmds == nil {
		sc.Cmds = []string{}
	}
}

func (me *SimpleCmd) activate(chat *defines.GameChat) bool {
	for _, t := range me.Cmds {
		c := utils.FormatByReplacingOccurrences(t, map[string]interface{}{
			"[player]": chat.Name,
			"[args]":   strings.Join(chat.Msg, " "),
		})
		me.Frame.GetGameControl().SendCmd(c)
	}

	return true
}

func (sc *SimpleCmd) Inject(frame defines.MainFrame) {
	sc.Frame = frame
	sc.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     sc.Triggers,
			ArgumentHint: sc.ArgumentHint,
			Usage:        sc.Usage,
			FinalTrigger: sc.FinalTrigger,
		},
		OptionalOnTriggerFn: sc.activate,
	})
}
