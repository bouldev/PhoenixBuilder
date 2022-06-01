package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"
	"strings"
	"time"
)

type StatusToggleLogic struct {
	Always                bool   `json:"总是触发"`
	ByArgument            string `json:"参数匹配时触发"`
	ByArgumentRegex       string `json:"参数符合正则表达式时触发"`
	compiledArgumentRegex *regexp.Regexp
	ByCondition           string      `json:"这条指令执行成功时触发"`
	ToggleCommandIn       interface{} `json:"触发指令"`
	toggleCommand         []defines.Cmd
	Hint                  string `json:"备注"`
}

type StatusToggle struct {
	*BasicComponent
	Triggers        []string             `json:"触发词"`
	Usage           string               `json:"提示信息"`
	ArgumentHint    string               `json:"若需要参数则参数提示为"`
	DefaultArgument string               `json:"没有给出参数时的默认参数"`
	Logic           []*StatusToggleLogic `json:"按顺序匹配以下切换逻辑并执行第一个成功项"`
}

func (o *StatusToggle) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	for _, l := range o.Logic {
		if l.toggleCommand, err = utils.ParseAdaptiveCmd(l.ToggleCommandIn); err != nil {
			panic(fmt.Errorf("在处理 %v 时，出现了错误 %v", l.ToggleCommandIn, err.Error()))
		} else {
			l.ToggleCommandIn = nil
		}
		if l.ByArgumentRegex != "" {
			l.compiledArgumentRegex = regexp.MustCompile(l.ByArgumentRegex)
		}
	}
}
func (o *StatusToggle) doMatch(player, argument string, idx int) {
	if idx == len(o.Logic) {
		return
	}
	success := false
	logic := o.Logic[idx]
	if logic.Always {
		success = true
	} else if logic.ByArgument != "" {
		success = logic.ByArgument == argument
	} else if logic.ByArgumentRegex != "" {
		success = !(logic.compiledArgumentRegex.Find([]byte(argument)) == nil)
	} else {
		t := time.NewTimer(time.Second)
		r := make(chan bool)
		testCmd := utils.FormatByReplacingOccurrences(logic.ByCondition, map[string]interface{}{
			"[player]": "\"" + player + "\"",
			"[args]":   argument,
		})
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(testCmd, func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				r <- true
			} else {
				r <- false
			}
		})
		select {
		case success = <-r:
		case <-t.C:
			success = false
		}
	}
	if success {
		utils.LaunchCmdsArray(o.Frame.GetGameControl(), logic.toggleCommand, map[string]interface{}{
			"[player]": "\"" + player + "\"",
			"[args]":   argument,
		}, o.Frame.GetBackendDisplay())
	} else {
		o.doMatch(player, argument, idx+1)
	}
}
func (o *StatusToggle) onTrigger(chat *defines.GameChat) (stop bool) {
	stop = true
	arguments := strings.Join(chat.Msg, " ")
	if arguments == "" {
		arguments = o.DefaultArgument
	}
	go o.doMatch(chat.Name, arguments, 0)
	return
}

func (o *StatusToggle) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: o.ArgumentHint,
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.onTrigger,
	})
}
