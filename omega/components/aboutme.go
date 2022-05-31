package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

type AboutMe struct {
	*BasicComponent
	Triggers []string            `json:"触发词"`
	Cmd      string              `json:"显示信息模版"`
	Tags     map[string][]string `json:"标签替换"`
	Usage    string              `json:"提示信息"`
}

func (o *AboutMe) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	if o.Tags == nil {
		o.Tags = map[string][]string{}
	}
}

func (o *AboutMe) show(chat *defines.GameChat) bool {
	//if player := o.Frame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
	//
	//}
	hasTags := map[string]bool{}
	allTags := []string{}
	scores := map[string]string{}
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("tag \""+chat.Name+"\" list", func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 && len(output.OutputMessages) > 0 && len(output.OutputMessages[0].Parameters) > 2 {
			for _, t := range strings.Split(output.OutputMessages[0].Parameters[2], ", ") {
				_t := utils.RemoveFormat(t)
				hasTags[_t] = true
				allTags = append(allTags, _t)
			}
		}
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("scoreboard players list \""+chat.Name+"\"", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				for _, p := range output.OutputMessages[1:] {
					if len(p.Parameters) == 3 {
						scores[p.Parameters[0]] = p.Parameters[2]
					}

				}
			}
			mappings := map[string]interface{}{
				"[tags]":   strings.Join(allTags, ", "),
				"[player]": "\"" + chat.Name + "\"",
			}
			for tagName, replacement := range o.Tags {
				ac, rj := replacement[0], replacement[1]
				if _, hasK := hasTags[tagName]; hasK {
					mappings[fmt.Sprintf("[tag<%v>]", tagName)] = ac
				} else {
					mappings[fmt.Sprintf("[tag<%v>]", tagName)] = rj
				}
			}
			for n, s := range scores {
				mappings[fmt.Sprintf("[score<%v>]", s)] = n
			}
			cmd := utils.FormatByReplacingOccurrences(o.Cmd, mappings)
			o.Frame.GetBackendDisplay().Write("玩家信息: " + cmd)
			o.Frame.GetGameControl().SayTo(chat.Name, cmd)
		})
	})
	return true
}

func (o *AboutMe) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if o.Usage == "" {
		o.Usage = "显示有关我的信息"
	}
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.show,
	})
}
