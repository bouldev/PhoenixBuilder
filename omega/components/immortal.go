package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"
)

type Immortal struct {
	*BasicComponent
	Hint          string `json:"提示信息"`
	AskForRespawn bool   `json:"询问是否返回死亡点"`
}

func (b *Immortal) Init(cfg *defines.ComponentConfig) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, b); err != nil {
		panic(err)
	}
}

func (o *Immortal) doRespawn(name string, pos []int) {
	o.frame.GetBackendDisplay().Write(fmt.Sprintf("%v 死于 %v", name, pos))
	t := time.NewTicker(time.Second * 3)
	released := false
	c := make(chan interface{})
	go func() {
		for {
			<-t.C
			o.frame.GetGameControl().SendCmdAndInvokeOnResponse("testfor "+name, func(output *packet.CommandOutput) {
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
	msg := utils.FormateByRepalcment(o.Hint, map[string]interface{}{
		"[player]":   name,
		"[dead_pos]": pos,
	})
	o.frame.GetGameControl().SayTo(name, msg)
	if o.AskForRespawn {
		hint, resolver := utils.GenYesNoResolver()
		if o.frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			b, err := resolver(chat.Msg)
			if err != nil {
				o.frame.GetGameControl().SayTo(name, "无法理解你的回答，因为"+err.Error())
				return true
			}
			if b {
				o.frame.GetBackendDisplay().Write(fmt.Sprintf("%v 回到复活点 %v", name, pos))
				o.frame.GetGameControl().SendCmd(fmt.Sprintf("tp %v %v %v %v", name, pos[0], pos[1], pos[2]))
			}
			return true
		}) == nil {
			o.frame.GetGameControl().SayTo(name, hint)
		}
	}
}

func (o *Immortal) intercept(chat *defines.GameChat) bool {
	if chat.Type == packet.TextTypeRaw || chat.Type == packet.TextTypeChat {
		return false
	}
	pkt := chat.Aux.(*packet.Text)
	if strings.Contains(pkt.Message, "death") && len(pkt.Parameters) > 0 {
		victim := pkt.Parameters[0]
		go func() {
			//fmt.Println(victim)
			pos := <-o.frame.GetGameControl().GetPlayerKit(victim).GetPos("@a[name=[player]]")
			if pos != nil {
				o.doRespawn(victim, pos)
			}
		}()
	}
	return false
}

func (o *Immortal) Inject(frame defines.MainFrame) {
	o.frame = frame
	frame.GetGameListener().SetGameChatInterceptor(o.intercept)
}

func (o *Immortal) Activate() {
	o.frame.GetGameControl().SendCmd("gamerule doimmediaterespawn false")
	o.frame.GetBackendDisplay().Write("返回死亡点需要关闭立即重生")
}
