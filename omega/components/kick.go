package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type Kick struct {
	*defines.BasicComponent
	Selector  string      `json:"选择器"`
	Duration  int         `json:"检查周期"`
	ActionsIn interface{} `json:"制裁"`
	Actions   []defines.Cmd
}

func (o *Kick) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	var err error
	if o.Actions, err = utils.ParseAdaptiveCmd(o.ActionsIn); err != nil {
		panic(err)
	}
}

type Banned struct {
	Victim []string `json:"victim"`
}

func (o *Kick) kick(name string) {
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.Actions, map[string]interface{}{
		"[player]": "\"" + name + "\"",
	}, o.Frame.GetBackendDisplay())
}

func (o *Kick) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Duration))
	for {
		<-t.C
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("/testfor %v", o.Selector), func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 && len(output.OutputMessages) > 0 {
				ban := &Banned{Victim: []string{}}
				err := json.Unmarshal([]byte(output.DataSet), &ban)
				if err != nil {
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("fail to get kick info " + err.Error()))
				} else {
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("try to kick %v", ban.Victim))
					for _, v := range ban.Victim {
						o.kick(v)
					}
				}
			}
		})
	}
}

func (o *Kick) Inject(frame defines.MainFrame) {
	o.Frame = frame
}
