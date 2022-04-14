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
	*BasicComponent
	Selector string   `json:"selector"`
	Duration int      `json:"check_duration_second"`
	Actions  []string `json:"actions"`
}

func (o *Kick) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

type Banned struct {
	Victim []string `json:"victim"`
}

func (o *Kick) kick(name string) {
	for _, a := range o.Actions {
		c := utils.FormateByRepalcment(a, map[string]interface{}{
			"[player]": name,
		})
		o.frame.GetGameControl().SendCmd(c)
	}
}

func (o *Kick) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Duration))
	for {
		<-t.C
		o.frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("/testfor %v", o.Selector), func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 && len(output.OutputMessages) > 0 {
				ban := &Banned{Victim: []string{}}
				err := json.Unmarshal([]byte(output.DataSet), &ban)
				if err != nil {
					o.frame.GetBackendDisplay().Write(fmt.Sprintf("fail to get kick info " + err.Error()))
				} else {
					o.frame.GetBackendDisplay().Write(fmt.Sprintf("try to kick %v", ban.Victim))
					for _, v := range ban.Victim {
						o.kick(v)
					}
				}
			}
		})
	}
}

func (o *Kick) Inject(frame defines.MainFrame) {
	o.frame = frame
}
