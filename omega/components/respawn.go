package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type Respawn struct {
	*BasicComponent
	Triggers                   []string `json:"触发词"`
	CmdsAfterNormalUserSpawned []string `json:"非创造玩家重生后执行的指令"`
	CmdsToKillCreativeUser     []string `json:"创造玩家重生后执行的指令"`
	DefaultImmediateRespawn    bool
}

func (o *Respawn) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *Respawn) respawn(chat *defines.GameChat) bool {
	killPlayer := fmt.Sprintf("kill @a[name=%v,m=!c]", chat.Name)
	if respawn := o.frame.GetUQHolder().GameRules["doimmediaterespawn"]; respawn != nil {
		o.DefaultImmediateRespawn = respawn.Value.(bool)
	}
	isCreative := false
	afterKill := func() {
		cmdsArr := o.CmdsAfterNormalUserSpawned
		if isCreative {
			cmdsArr = o.CmdsToKillCreativeUser
		}
		c := make(chan interface{})
		go func() {
			<-c
			for _, cmdT := range cmdsArr {
				cmd := utils.FormateByRepalcment(cmdT, map[string]interface{}{
					"[player]": chat.Name,
				})
				o.frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {})
			}
		}()
		go func() {
			t := time.NewTicker(time.Second * 2)
			released := false
			for {
				<-t.C
				o.frame.GetGameControl().SendCmdAndInvokeOnResponse("testfor "+chat.Name, func(output *packet.CommandOutput) {
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
	}
	kill := func(cb func()) {
		o.frame.GetGameControl().SendCmdAndInvokeOnResponse(
			fmt.Sprintf(killPlayer), func(output *packet.CommandOutput) {
				if len(output.OutputMessages) > 0 && output.OutputMessages[0].Message == "commands.generic.noTargetMatch" {
					//o.frame.GetBackendDisplay().Write(fmt.Sprintf("%v is created", chat.Name))
					o.frame.GetGameControl().SendCmd("gamemode a " + chat.Name)
					isCreative = true
					o.frame.GetGameControl().SendCmdAndInvokeOnResponse("kill "+chat.Name, func(output *packet.CommandOutput) {
						cb()
					})
				} else {
					cb()
				}
			})
	}
	if !o.DefaultImmediateRespawn {
		o.frame.GetGameControl().SendCmdAndInvokeOnResponse("gamerule doimmediaterespawn true",
			func(output *packet.CommandOutput) {
				kill(func() {
					o.frame.GetGameControl().SendCmd("gamerule doimmediaterespawn false")
					afterKill()
				})
			})
	} else {
		kill(func() {
			afterKill()
		})
	}

	return true
}

func (o *Respawn) Inject(frame defines.MainFrame) {
	o.frame = frame
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "返回重生点",
		},
		OptionalOnTriggerFn: o.respawn,
	})
}
