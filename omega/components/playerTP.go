package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type PlayerTP struct {
	*BasicComponent
	Triggers             []string `json:"triggers"`
	HintOnNoTarget       string   `json:"hint_on_no_target"`
	HintOnNoPlayer       string   `json:"hint_on_no_player"`
	HintOnReq            string   `json:"hint_on_request_permission"`
	HintOnReqTooFrequent string   `json:"hint_on_req_too_frequent"`
	HintOnReqSent        string   `json:"hint_on_req_sent"`
	HintOnTargetBusy     string   `json:"hint_on_target_busy"`
	HintOnRefuse         string   `json:"hint_on_refuse"`
	CoolDownSecond       int      `json:"cool_down_second"`
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
	reqMsg := utils.FormateByRepalcment(o.HintOnReq, map[string]interface{}{
		"[src]": src,
		"[dst]": dst,
	})
	hint, ynResolver := utils.GenYesNoResolver()
	if o.frame.GetGameControl().SetOnParamMsg(dst, func(chat *defines.GameChat) (catch bool) {
		result, err := ynResolver(chat.Msg)
		if err != nil {
			o.frame.GetGameControl().SayTo(chat.Name, "抱歉，我不明白你的意思，因为输入"+err.Error())
			return true
		}
		if result {
			o.frame.GetGameControl().SendCmd(fmt.Sprintf("tp %v %v", src, dst))
			o.frame.GetBackendDisplay().Write(fmt.Sprintf("accept tp %v -> %v", src, dst))
			o.frame.GetGameControl().SayTo(src, "传送开始")
			o.frame.GetGameControl().SayTo(dst, "传送开始")
		} else {
			m := utils.FormateByRepalcment(o.HintOnRefuse, map[string]interface{}{
				"[src]": src,
				"[dst]": dst,
			})
			o.frame.GetBackendDisplay().Write(fmt.Sprintf("reject tp %v -> %v", src, dst))
			o.frame.GetGameControl().SayTo(src, m)
		}
		return true
	}) == nil {
		sendMsg := utils.FormateByRepalcment(o.HintOnReqSent, map[string]interface{}{
			"[src]": src,
			"[dst]": dst,
		})
		o.frame.GetGameControl().SayTo(src, sendMsg)
		o.frame.GetGameControl().SayTo(dst, reqMsg)
		o.frame.GetGameControl().SayTo(dst, hint)
		o.frame.GetBackendDisplay().Write(fmt.Sprintf("request tp %v -> %v", src, dst))
		o.lastRequestTime[src] = time.Now()
	} else {
		o.frame.GetGameControl().SayTo(src, o.HintOnTargetBusy)
	}
}

func (o *PlayerTP) check(chat *defines.GameChat) bool {
	if t, hask := o.lastRequestTime[chat.Name]; hask {
		if time.Now().Sub(t).Seconds() < float64(o.CoolDownSecond) {
			o.frame.GetGameControl().SayTo(chat.Name, o.HintOnReqTooFrequent)
			return true
		}
	}
	noTarget := len(chat.Msg) == 0
	var availablePlayers []string
	flag := false
	for _, p := range o.frame.GetUQHolder().PlayersByEntityID {
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
		o.frame.GetGameControl().SayTo(chat.Name, o.HintOnNoTarget)
	} else {
		o.frame.GetGameControl().SayTo(chat.Name, o.HintOnNoPlayer)
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(availablePlayers)
	if o.frame.GetGameControl().SetOnParamMsg(chat.Name,
		func(chat *defines.GameChat) (catch bool) {
			i, err := resolver(chat.Msg)
			if err != nil {
				o.frame.GetGameControl().SayTo(chat.Name, "无法传送，因为输入"+err.Error())
				return true
			}
			o.requestTp(chat.Name, availablePlayers[i])
			return true
		}) == nil {
		o.frame.GetGameControl().SayTo(chat.Name, "可选项有: "+hint+" 请输入喔:")
	}
	return true
}

func (o *PlayerTP) Inject(frame defines.MainFrame) {
	o.frame = frame
	o.frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[玩家名]",
			FinalTrigger: false,
			Usage:        "将自己传送到目标玩家身边，需要对方同意",
		},
		OptionalOnTriggerFn: o.check,
	})
}
