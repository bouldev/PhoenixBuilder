package liliya

import (
	"encoding/json"
	"math"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"

	"github.com/google/uuid"
)

type AntiAFK struct {
	*defines.BasicComponent
	Duration int         `json:"检测周期_秒"`
	CmdsIn   interface{} `json:"符合条件时执行的指令"`
	cmds     []defines.Cmd
	lastYRot map[uuid.UUID]float64
}

func (o *AntiAFK) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	if cfg.Version == "0.0.1" {
		cfg.Configs["检测周期_秒"] = cfg.Configs["检测周期"].(float64) * 60
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
	var err error
	if o.cmds, err = utils.ParseAdaptiveCmd(o.CmdsIn); err != nil {
		panic(err)
	}
	if o.Duration < 1 {
		panic("检测周期不能小于1, 请更改配置文件")
	}
	o.lastYRot = make(map[uuid.UUID]float64)
}

func (o *AntiAFK) Inject(frame defines.MainFrame) {
	o.Frame = frame
}

func (o *AntiAFK) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Duration))
	for {
		<-t.C
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("querytarget @a", func(output *packet.CommandOutput) {
			if output.SuccessCount == 0 {
				return
			}
			var QueryResults []struct {
				UUID uuid.UUID `json:"uniqueId"`
				YRot float64   `json:"yRot"`
			}
			newMap := make(map[uuid.UUID]float64)
			for _, v := range output.OutputMessages {
				for _, j := range v.Parameters {
					err := json.Unmarshal([]byte(j), &QueryResults)
					if err != nil {
						continue
					}
					for _, u := range QueryResults {
						playerName := o.Frame.GetUQHolder().GetPlayersByUUID(u.UUID).Username
						if playerName == o.Frame.GetUQHolder().GetBotName() {
							continue
						}
						if v, ok := o.lastYRot[u.UUID]; ok && (v == u.YRot || math.Abs(v)+math.Abs(u.YRot) == 360) {
							replacement := map[string]interface{}{
								"[player]":                   "\"" + playerName + "\"",
								"[player_without_quotation]": playerName,
							}
							go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.cmds, replacement, o.Frame.GetBackendDisplay())
						} else {
							newMap[u.UUID] = u.YRot
						}
					}
				}
			}
			o.lastYRot = newMap
		})
	}
}
