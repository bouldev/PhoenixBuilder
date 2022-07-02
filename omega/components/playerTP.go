package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type PlayerTP struct {
	*defines.BasicComponent
	Triggers             []string `json:"触发词"`
	HintOnNoTarget       string   `json:"没有指定玩家时提示"`
	HintOnNoPlayer       string   `json:"没有对应玩家时提示"`
	HintOnReq            string   `json:"询问是否同意传送"`
	HintOnReqTooFrequent string   `json:"请求过于频繁时提示"`
	HintOnReqSent        string   `json:"请求发送时提示"`
	HintOnTargetBusy     string   `json:"目标玩家忙碌时提示"`
	HintOnRefuse         string   `json:"目标玩家拒绝时提示"`
	CoolDownSecond       int      `json:"请求冷却时间"`
	TPCmd                string   `json:"传送指令"`
	lastRequestTime      map[string]time.Time
}

func (o *PlayerTP) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	o.lastRequestTime = map[string]time.Time{}
}

func (o *PlayerTP) requestTp(src, dst string) {
	reqMsg := utils.FormatByReplacingOccurrences(o.HintOnReq, map[string]interface{}{
		"[src]": "\"" + src + "\"",
		"[dst]": "\"" + dst + "\"",
	})
	hint, ynResolver := utils.GenYesNoResolver()
	if o.Frame.GetGameControl().SetOnParamMsg(dst, func(chat *defines.GameChat) (catch bool) {
		result, err := ynResolver(chat.Msg)
		if err != nil {
			o.Frame.GetGameControl().SayTo(chat.Name, "抱歉，我不明白你的意思，因为输入"+err.Error())
			return true
		}
		if result {
			tpCmd := utils.FormatByReplacingOccurrences(o.TPCmd, map[string]interface{}{
				"[src]": "\"" + src + "\"",
				"[dst]": "\"" + dst + "\"",
			})
			o.Frame.GetGameControl().SendCmd(tpCmd)
			// fmt.Println(tpCmd)
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("accept tp %v -> %v", src, dst))
			o.Frame.GetGameControl().SayTo(src, "传送开始")
			o.Frame.GetGameControl().SayTo(dst, "传送开始")
		} else {
			m := utils.FormatByReplacingOccurrences(o.HintOnRefuse, map[string]interface{}{
				"[src]": "\"" + src + "\"",
				"[dst]": "\"" + dst + "\"",
			})
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("reject tp %v -> %v", src, dst))
			o.Frame.GetGameControl().SayTo(src, m)
		}
		return true
	}) == nil {
		sendMsg := utils.FormatByReplacingOccurrences(o.HintOnReqSent, map[string]interface{}{
			"[src]": "\"" + src + "\"",
			"[dst]": "\"" + dst + "\"",
		})
		o.Frame.GetGameControl().SayTo(src, sendMsg)
		o.Frame.GetGameControl().SayTo(dst, reqMsg)
		o.Frame.GetGameControl().SayTo(dst, hint)
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("request tp %v -> %v", src, dst))
		o.lastRequestTime[src] = time.Now()
	} else {
		o.Frame.GetGameControl().SayTo(src, o.HintOnTargetBusy)
	}
}

func (o *PlayerTP) check(chat *defines.GameChat) bool {
	if t, hask := o.lastRequestTime[chat.Name]; hask {
		timeLeft := time.Now().Sub(t).Seconds()
		if timeLeft < float64(o.CoolDownSecond) {
			o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnReqTooFrequent+fmt.Sprintf("(还剩%v秒)", int(timeLeft)))
			return true
		}
	}
	noTarget := len(chat.Msg) == 0
	var availablePlayers []string
	flag := false
	for _, p := range o.Frame.GetUQHolder().PlayersByEntityID {
		availablePlayers = append(availablePlayers, p.Username)
		if !noTarget && chat.Msg[0] == p.Username {
			flag = true
		}
	}
	if flag {
		o.requestTp(chat.Name, chat.Msg[0])
		return true
	}
	if noTarget {
		o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnNoTarget)
	} else {
		o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnNoPlayer)
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(availablePlayers)
	if o.Frame.GetGameControl().SetOnParamMsg(chat.Name,
		func(chat *defines.GameChat) (catch bool) {
			i, cancel, err := resolver(chat.Msg)
			if cancel {
				o.Frame.GetGameControl().SayTo(chat.Name, "已取消")
				return true
			}
			if err != nil {
				o.Frame.GetGameControl().SayTo(chat.Name, "无法传送，因为输入"+err.Error())
				return true
			}
			o.requestTp(chat.Name, availablePlayers[i])
			return true
		}) == nil {
		o.Frame.GetGameControl().SayTo(chat.Name, "可选项有: "+hint+" 请输入喔:")
	}
	return true
}

func (o *PlayerTP) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[玩家名]",
			FinalTrigger: false,
			Usage:        "将自己传送到目标玩家身边，需要对方同意",
		},
		OptionalOnTriggerFn: o.check,
	})
}
