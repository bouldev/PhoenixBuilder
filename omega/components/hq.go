package components

import (
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type BackToHQ struct {
	*defines.BasicComponent
	Triggers       []string      `json:"触发词"`
	ToAnchor       string        `json:"世界锚点"`
	ToHQ           string        `json:"主城锚点"`
	cmdsBeforeBack []defines.Cmd `json:"回城前执行"`
}

func (o *BackToHQ) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	var err error
	if v, hask := cfg.Configs["回城前执行"]; hask && v != nil {
		if o.cmdsBeforeBack, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"回城前执行"}); err != nil {
			panic(err)
		}
	}

	if o.cmdsBeforeBack == nil {
		o.cmdsBeforeBack = make([]defines.Cmd, 0)
	}
}

func (o *BackToHQ) back(chat *defines.GameChat) bool {
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.cmdsBeforeBack, map[string]interface{}{
		"[player]": "\"" + chat.Name + "\"",
	}, o.Frame.GetBackendDisplay())
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(
		utils.FormatByReplacingOccurrences(o.ToAnchor, map[string]interface{}{
			"[player]": "\"" + chat.Name + "\"",
			"[bot]":    "\"" + o.Frame.GetUQHolder().GetBotName() + "\"",
		}), func(output *packet.CommandOutput) {
			go func() {
				<-time.NewTimer(time.Second / 20).C
				o.Frame.GetGameControl().SendCmd(
					utils.FormatByReplacingOccurrences(o.ToHQ, map[string]interface{}{
						"[player]": "\"" + chat.Name + "\"",
						"[bot]":    "\"" + o.Frame.GetUQHolder().GetBotName() + "\"",
					}),
				)
			}()
		})
	return true
}

func (o *BackToHQ) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "返回主城",
		},
		OptionalOnTriggerFn: o.back,
	})
}
