package liliya

import (
	"encoding/json"
	"math"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

type PlayersSleepingPercentage struct {
	*defines.BasicComponent
	Percentage               int         `json:"入睡比例"`
	CmdsWhenChangeIn         interface{} `json:"变更时执行的指令"`
	CmdsWhenSatisfyIn        interface{} `json:"满足入睡比例时执行的指令"`
	cmdsWhenChange           []defines.Cmd
	cmdsWhenSatisfy          []defines.Cmd
	cmdsWhenChangeExecuting  bool
	cmdsWhenSatisfyExecuting bool
}

func (o *PlayersSleepingPercentage) Init(cfg *defines.ComponentConfig) {
	var err error
	marshal, _ := json.Marshal(cfg.Configs)
	if err = json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
	if o.cmdsWhenChange, err = utils.ParseAdaptiveCmd(o.CmdsWhenChangeIn); err != nil {
		panic(err)
	}
	if o.cmdsWhenSatisfy, err = utils.ParseAdaptiveCmd(o.CmdsWhenSatisfyIn); err != nil {
		panic(err)
	}
	if o.Percentage < 0 {
		panic("入睡比例不能小于0, 请更改配置文件")
	}
	o.cmdsWhenChangeExecuting, o.cmdsWhenSatisfyExecuting = false, false
}

func (o *PlayersSleepingPercentage) getCountOfPlayersInOtherDimensions() (result int32) {
	waitChan := make(chan struct{})
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("querytarget @a", func(output *packet.CommandOutput) {
		if output.SuccessCount != 0 {
			var QueryResults []struct {
				Dimension int `json:"dimension"`
			}
			for _, v := range output.OutputMessages {
				for _, j := range v.Parameters {
					err := json.Unmarshal([]byte(j), &QueryResults)
					if err != nil {
						continue
					}
					for _, u := range QueryResults {
						if u.Dimension != 0 {
							result++
						}
					}
				}
			}
		}
		close(waitChan)
	})
	<-waitChan
	return result
}

func (o *PlayersSleepingPercentage) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDLevelEvent, func(p packet.Packet) {
		go func() {
			pkt := p.(*packet.LevelEvent)
			if pkt.EventType != packet.LevelEventSleepingPlayers {
				return
			}
			total, sleeping := pkt.EventData>>16-1, pkt.EventData%65536
			if sleeping == o.getCountOfPlayersInOtherDimensions() {
				return
			}
			need := math.Ceil(float64(total)/100*float64(o.Percentage) - float64(sleeping))
			replacement := map[string]interface{}{
				"[total]":    total,
				"[sleeping]": sleeping,
				"[need]":     need,
			}
			if need > 0 {
				if !o.cmdsWhenChangeExecuting {
					o.cmdsWhenChangeExecuting = true
					utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.cmdsWhenChange, replacement, o.Frame.GetBackendDisplay())
					o.cmdsWhenChangeExecuting = false
				}
			} else {
				if !o.cmdsWhenSatisfyExecuting {
					o.cmdsWhenSatisfyExecuting = true
					utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.cmdsWhenSatisfy, replacement, o.Frame.GetBackendDisplay())
					o.cmdsWhenSatisfyExecuting = false
				}
			}
		}()
	})
}
