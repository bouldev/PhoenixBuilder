package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

type SetSpawnPoint struct {
	*BasicComponent
	Triggers             []string `json:"触发词"`
	HintOnSuccess        string   `json:"设置成功时的提示"`
	HintOnFailure        string   `json:"设置失败时的提示"`
	HintOnUnknownFailure string   `json:"出现未知错误时的提示"`
	Cmd                  string   `json:"设置方式"`
}

func (o *SetSpawnPoint) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *SetSpawnPoint) set(chat *defines.GameChat) bool {
	cmd := utils.FormatByReplacingOccurrences(o.Cmd, map[string]interface{}{
		"[player]": "\"" + chat.Name + "\"",
	})
	o.Frame.GetBackendDisplay().Write(cmd)
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(
		fmt.Sprintf(cmd), func(output *packet.CommandOutput) {
			if output.SuccessCount != 0 {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("success %v", output))
				o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnSuccess)
			} else {
				if len(output.OutputMessages) > 0 && output.OutputMessages[0].Message != "commands.generic.noTargetMatch" {
					o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnUnknownFailure)
				} else {
					o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnFailure)
				}
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("failure %v", output))
			}
		})
	return true
}

func (o *SetSpawnPoint) Inject(frame defines.MainFrame) {
	o.Frame = frame
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "设置重生点",
		},
		OptionalOnTriggerFn: o.set,
	})
}
