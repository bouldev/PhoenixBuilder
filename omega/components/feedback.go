package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

type FeedBack struct {
	*BasicComponent
	logger      defines.LineDst
	Response    string   `json:"response"`
	FileName    string   `json:"file_name"`
	Triggers    []string `json:"triggers"`
	HintOnEmpty string   `json:"hint_on_empty"`
}

func (me *FeedBack) record(chat *defines.GameChat) bool {
	msg := strings.Join(chat.Msg, "")
	if msg == "" {
		if player := me.frame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
			if player.SetOnParamMsg(func(c *defines.GameChat) bool {
				me.record(c)
				return true
			}) == nil {
				me.frame.GetGameControl().SayTo(chat.Name, me.HintOnEmpty)
			}
		}
		return true
	}
	me.logger.Write(fmt.Sprintf("[%v] %v:%v ", chat.Type, chat.Name, chat.Msg))
	m := utils.FormateByRepalcment(me.Response, map[string]interface{}{
		"[player]": chat.Name,
		"[msg]":    msg,
	})
	me.frame.GetGameControl().SendCmd(m)
	return true
}

func (me *FeedBack) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, me); err != nil {
		panic(err)
	}
}

func (me *FeedBack) Inject(frame defines.MainFrame) {
	me.frame = frame

	me.logger = &utils.MultipleLogger{Loggers: []defines.LineDst{
		me.frame.GetLogger(me.FileName),
		me.frame.GetBackendDisplay(),
	}}
	me.frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     me.Triggers,
			ArgumentHint: "[反馈]",
			Usage:        "给腐竹反馈信息",
			FinalTrigger: false,
		},
		OptionalOnTriggerFn: me.record,
	})
}
