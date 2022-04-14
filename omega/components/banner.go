package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
)

type Banner struct {
	*BasicComponent
	Cmds [][]string `json:"启动时信息显示" yaml:"启动时信息显示"`
}

func (b *Banner) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	for _, cmd := range b.Cmds {
		if len(cmd) != 2 {
			panic(fmt.Errorf("%v should be [target,msg]", cmd))
		}
	}
}

func (b *Banner) Activate() {
	for _, cmd := range b.Cmds {
		b.frame.GetGameControl().SayTo(cmd[0], cmd[1])
	}
}
