package micro_omega

import (
	"encoding/json"
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol/packet"
	"fastbuilder-core/lib/minecraft/mirror/define"
	"fastbuilder-core/lib/minecraft/omega/omega"
	"fastbuilder-core/lib/minecraft/omega/uq_holder"
	"fastbuilder-core/lib/minecraft/omega/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PlayerKitOmega struct {
	uq         *uq_holder.UQHolder
	ctrl       *GameCtrl
	name       string
	OnParamMsg func(chat *omega.GameChat) (catch bool)
	playerUQ   *uq_holder.Player
	Permission map[string]bool
}

func (p *PlayerKitOmega) HasPermission(key string) bool {
	if auth, hasK := p.Permission[key]; hasK && auth {
		return true
	}
	return false
}
func (b *PlayerKitOmega) GetPlayerNameByUUid(Theuuid string) string {
	UUID, err := uuid.Parse(Theuuid)
	if err != nil {
		fmt.Println(err)
	}
	if player := b.ctrl.GetPlayerKitByUUID(UUID); player != nil {
		username := player.GetRelatedUQ().Username
		return username
	}
	return ""
}
func (p *PlayerKitOmega) GetPos(selector string) chan *define.CubePos {
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

	p.ctrl.SendCmdAndInvokeOnResponse("querytarget "+s, func(output *packet.CommandOutput) {
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

func (p *PlayerKitOmega) SetPermission(key string, b bool) {
	p.Permission[key] = b
}

func (p *PlayerKitOmega) SetOnParamMsg(f func(chat *omega.GameChat) (catch bool)) error {
	if p.OnParamMsg != nil {
		return fmt.Errorf("player busy")
	}
	p.OnParamMsg = f
	return nil
}

func (p *PlayerKitOmega) GetOnParamMsg() func(chat *omega.GameChat) (catch bool) {
	f := p.OnParamMsg
	p.OnParamMsg = nil
	return f
}

func newPlayerKitOmega(uq *uq_holder.UQHolder, ctrl *GameCtrl, name string) *PlayerKitOmega {
	pko, k := ctrl.perPlayerStorage[name]
	if k {
		return pko
	}
	player := &PlayerKitOmega{
		uq:         uq,
		ctrl:       ctrl,
		name:       name,
		OnParamMsg: nil,
	}
	ctrl.perPlayerStorage[name] = player
	return player
}

func (p *PlayerKitOmega) Say(msg string) {
	p.ctrl.SayTo(p.name, msg)
}

func (p *PlayerKitOmega) RawSay(msg string) {
	p.ctrl.RawSayTo(p.name, msg)
}

func (p *PlayerKitOmega) ActionBar(msg string) {
	p.ctrl.ActionBarTo(p.name, msg)
}

func (p *PlayerKitOmega) Title(msg string) {
	p.ctrl.TitleTo(p.name, msg)
}

func (p *PlayerKitOmega) SubTitle(msg string) {
	p.ctrl.SubTitleTo(p.name, msg)
}

func (p *PlayerKitOmega) GetRelatedUQ() *uq_holder.Player {
	if p.playerUQ != nil {
		return p.playerUQ
	}
	for _, player := range p.uq.PlayersByEntityID {
		if player.Username == p.name {
			p.playerUQ = player
			return player
		}
	}
	return nil
}
