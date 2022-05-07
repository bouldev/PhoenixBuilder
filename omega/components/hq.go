package components

import (
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type BackToHQ struct {
	*BasicComponent
	Triggers []string `json:"触发词"`
	ToAnchor string   `json:"世界锚点"`
	ToHQ     string   `json:"主城锚点"`
}

func (o *BackToHQ) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *BackToHQ) back(chat *defines.GameChat) bool {
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(
		utils.FormateByRepalcment(o.ToAnchor, map[string]interface{}{
			"[player]": chat.Name,
			"[bot]":    o.Frame.GetUQHolder().GetBotName(),
		}), func(output *packet.CommandOutput) {
			go func() {
				<-time.NewTimer(time.Second / 20).C
				o.Frame.GetGameControl().SendCmd(
					utils.FormateByRepalcment(o.ToHQ, map[string]interface{}{
						"[player]": chat.Name,
						"[bot]":    o.Frame.GetUQHolder().GetBotName(),
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
