package components

import (
	"encoding/json"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

type SimpleCmd struct {
	*defines.BasicComponent
	Triggers     []string    `json:"触发词"`
	FinalTrigger bool        `json:"总是触发"`
	ArgumentHint string      `json:"参数描述"`
	Usage        string      `json:"功能描述"`
	CmdsIn       interface{} `json:"触发时执行指令"`
	Cmds         []defines.Cmd
}

func (sc *SimpleCmd) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, sc); err != nil {
		panic(err)
	}
	var err error
	if sc.Cmds, err = utils.ParseAdaptiveCmd(sc.CmdsIn); err != nil {
		panic(err)
	}
}

func (o *SimpleCmd) activate(chat *defines.GameChat) bool {
	mappings := map[string]interface{}{
		"[player]": "\"" + chat.Name + "\"",
		"[args]":   strings.Join(chat.Msg, " "),
	}
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.Cmds, mappings, o.Frame.GetBackendDisplay())
	// o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("scoreboard players list \""+chat.Name+"\"", func(output *packet.CommandOutput) {
	// 	scores := map[string]string{}
	// 	if output.SuccessCount > 0 {
	// 		for _, p := range output.OutputMessages[1:] {
	// 			if len(p.Parameters) == 3 {
	// 				scores[p.Parameters[0]] = p.Parameters[2]
	// 			}
	// 		}
	// 	}

	// 	for n, s := range scores {
	// 		mappings[fmt.Sprintf("[score<%v>]", s)] = n
	// 	}
	// 	for _, t := range o.Cmds {
	// 		c := utils.FormatByReplacingOccurrences(t, mappings)
	// 		o.Frame.GetGameControl().SendCmd(c)
	// 	}
	// })

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
