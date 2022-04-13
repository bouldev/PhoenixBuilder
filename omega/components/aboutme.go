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
	Triggers []string            `json:"triggers"`
	Cmd      string              `json:"cmd"`
	Tags     map[string][]string `json:"tags_replacement"`
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
	//if player := o.frame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
	//
	//}
	hasTags := map[string]bool{}
	allTags := []string{}
	scores := map[string]string{}
	o.frame.GetGameControl().SendCmdAndInvokeOnResponse("tag "+chat.Name+" list", func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			for _, t := range strings.Split(output.OutputMessages[0].Parameters[2], ", ") {
				_t := utils.RemoveFormate(t)
				hasTags[_t] = true
				allTags = append(allTags, _t)
			}
		}
		o.frame.GetGameControl().SendCmdAndInvokeOnResponse("scoreboard players list "+chat.Name, func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				for _, p := range output.OutputMessages[1:] {
					if len(p.Parameters) == 3 {
						scores[p.Parameters[0]] = p.Parameters[2]
					}

				}
			}
			mappings := map[string]interface{}{
				"[tags]":   "[" + strings.Join(allTags, ", ") + "]",
				"[player]": chat.Name,
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
			cmd := utils.FormateByRepalcment(o.Cmd, mappings)
			o.frame.GetGameControl().SayTo(chat.Name, cmd)
		})
	})
	return true
}

func (o *AboutMe) Inject(frame defines.MainFrame) {
	o.frame = frame
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "显示有关我的信息",
		},
		OptionalOnTriggerFn: o.show,
	})
}
