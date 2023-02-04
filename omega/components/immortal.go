package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"
)

type Immortal struct {
	*defines.BasicComponent
	Hint          string      `json:"提示信息"`
	AskForRespawn bool        `json:"询问是否返回死亡点"`
	Selector      string      `json:"死亡玩家选择器"`
	CmdsIn        interface{} `json:"复活时执行指令"`
	Cmds          []defines.Cmd
}

func (b *Immortal) Init(cfg *defines.ComponentConfig) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, b); err != nil {
		panic(err)
	}
	if b.Selector == "" {
		b.Selector = "@a[name=[player]]"
	}
	var err error
	if b.CmdsIn == nil {
		b.CmdsIn = []string{}
	}
	if b.Cmds, err = utils.ParseAdaptiveCmd(b.CmdsIn); err != nil {
		panic(err)
	}
}

func (o *Immortal) doRespawn(name string, pos []int) {
	o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 死于 %v", name, pos))
	t := time.NewTicker(time.Second * 3)
	released := false
	c := make(chan interface{})
	go func() {
		for {
			<-t.C
			o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("testfor \""+name+"\"", func(output *packet.CommandOutput) {
				if output.SuccessCount != 0 {
					if !released {
						c <- true
						released = true
					}
				}
			})
			if released {
				return
			}
		}
	}()
	<-c
	utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.Cmds, map[string]interface{}{
		"[player]":   "\"" + name + "\"",
		"[dead_pos]": pos,
	}, o.Frame.GetBackendDisplay())
	msg := utils.FormatByReplacingOccurrences(o.Hint, map[string]interface{}{
		"[player]":   "\"" + name + "\"",
		"[dead_pos]": pos,
	})
	o.Frame.GetGameControl().SayTo(name, msg)
	if collaborate_func, hasK := o.Frame.GetContext(collaborate.INTERFACE_GEN_YES_NO_RESOLVER); hasK && o.AskForRespawn {
		hint, resolver := collaborate_func.(collaborate.GEN_YES_NO_RESOLVER)()
		if o.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			b, err := resolver(chat.Msg)
			if err != nil {
				o.Frame.GetGameControl().SayTo(name, "无法理解你的回答，因为"+err.Error())
				return true
			}
			if b {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 回到复活点 %v", name, pos))
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp \"%v\" %v %v %v", name, pos[0], pos[1], pos[2]))
			}
			return true
		}) == nil {
			o.Frame.GetGameControl().SayTo(name, hint)
		}
	}
}

func (o *Immortal) intercept(chat *defines.GameChat) bool {
	if chat.Type == packet.TextTypeRaw || chat.Type == packet.TextTypeChat {
		return false
	}
	if strings.Contains(chat.RawMsg, "death") && len(chat.RawParameters) > 0 {
		victim := chat.RawParameters[0]
		go func() {
			//fmt.Println(victim)
			pos := <-o.Frame.GetGameControl().GetPlayerKit(victim).GetPos(o.Selector)
			if pos != nil {
				o.doRespawn(victim, []int{pos.X(), pos.Y(), pos.Z()})
			}
		}()
	}
	return false
}

func (o *Immortal) Inject(frame defines.MainFrame) {
	o.Frame = frame
	frame.GetGameListener().SetGameChatInterceptor(o.intercept)
}

func (o *Immortal) Activate() {
	o.Frame.GetGameControl().SendCmd("gamerule doimmediaterespawn false")
	o.Frame.GetBackendDisplay().Write("返回死亡点需要关闭立即重生")
}
