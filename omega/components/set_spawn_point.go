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
	Triggers             []string `json:"triggers"`
	HintOnSuccess        string   `json:"hint_on_success"`
	HintOnFailure        string   `json:"hint_on_failure"`
	HintOnUnknownFailure string   `json:"hint_on_unknown_failure"`
	Cmd                  string   `json:"cmd"`
}

func (o *SetSpawnPoint) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *SetSpawnPoint) set(chat *defines.GameChat) bool {
	cmd := utils.FormateByRepalcment(o.Cmd, map[string]interface{}{
		"[player]": chat.Name,
	})
	o.frame.GetBackendDisplay().Write(cmd)
	o.frame.GetGameControl().SendCmdAndInvokeOnResponse(
		fmt.Sprintf(cmd), func(output *packet.CommandOutput) {
			if output.SuccessCount != 0 {
				o.frame.GetBackendDisplay().Write(fmt.Sprintf("success %v", output))
				o.frame.GetGameControl().SayTo(chat.Name, o.HintOnSuccess)
			} else {
				if len(output.OutputMessages) > 0 && output.OutputMessages[0].Message != "commands.generic.noTargetMatch" {
					o.frame.GetGameControl().SayTo(chat.Name, o.HintOnUnknownFailure)
				} else {
					o.frame.GetGameControl().SayTo(chat.Name, o.HintOnFailure)
				}
				o.frame.GetBackendDisplay().Write(fmt.Sprintf("failure %v", output))
			}
		})
	return true
}

func (o *SetSpawnPoint) Inject(frame defines.MainFrame) {
	o.frame = frame
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
