package playerKit

import (
	"encoding/json"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/utils"
	"time"
)

func init() {
	if false {
		func(kit omega.PlayerKit) {}(&PlayerKit{})
	}
}

type PlayerKit struct {
	name string
	ctrl interface {
		omega.CmdSender
		omega.InfoSender
	}
	uq omega.PlayerUQReader
}

func NewPlayerKit(name string, ctrl interface {
	omega.CmdSender
	omega.InfoSender
}, uq omega.PlayerUQReader) omega.PlayerKit {
	pk := &PlayerKit{
		name: name,
		ctrl: ctrl,
		uq:   uq,
	}
	return pk
}

func (p *PlayerKit) Say(msg string) {
	p.ctrl.SayTo(p.name, msg)
}

func (p *PlayerKit) RawSay(msg string) {
	p.ctrl.RawSayTo(p.name, msg)
}

func (p *PlayerKit) ActionBar(msg string) {
	p.ctrl.ActionBarTo(p.name, msg)
}

func (p *PlayerKit) Title(msg string) {
	p.ctrl.TitleTo(p.name, msg)
}

func (p *PlayerKit) SubTitle(msg string) {
	p.ctrl.SubTitleTo(p.name, msg)
}

func (p *PlayerKit) GetRelatedUQ() omega.PlayerUQReader {
	return p.uq
}

//func (p *PlayerKit) SetOnParamMsg(f func(chat omega.GameChat) (catch bool)) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (p *PlayerKit) GetOnParamMsg() func(chat omega.GameChat) (catch bool) {
//	//TODO implement me
//	panic("implement me")
//}

func (p *PlayerKit) GetPos(selector string) chan *define.CubePos {
	s := utils.FormatByReplacingOccurrences(selector, map[string]interface{}{
		"[player]": "\"" + p.name + "\"",
	})
	c := make(chan *define.CubePos)
	sent := false
	send := func(d *define.CubePos) {
		if sent {
			return
		}
		sent = true
		c <- d
	}
	var QueryResults []struct {
		Position *struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"position"`
		Uuid string `json:"uniqueId"`
	}

	p.ctrl.SendWSCmdAndInvokeOnResponse("querytarget "+s, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			for _, v := range output.OutputMessages {
				for _, j := range v.Parameters {
					err := json.Unmarshal([]byte(j), &QueryResults)
					if err != nil {
						send(nil)
					}
					for _, u := range QueryResults {
						send(&define.CubePos{
							int(u.Position.X),
							int(u.Position.Y),
							int(u.Position.Z),
						})
					}
				}
			}

		}
		send(nil)
	})
	go func() {
		<-time.NewTicker(time.Second * 3).C
		send(nil)
	}()
	return c
}
