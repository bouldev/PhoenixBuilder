package components

import (
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"time"
)

type Crash struct {
	*defines.BasicComponent
	Delay int `json:"强制退出延迟"`
}

func (b *Crash) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
}
func (b *Crash) Stop() error {
	go func() {
		<-time.NewTimer(time.Duration(b.Delay) * time.Second).C
		b.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback("kick @s", func(output *packet.CommandOutput) {})
		<-time.NewTimer(time.Duration(b.Delay) * time.Second).C
		panic("强制退出")
	}()
	return nil
}
func (b *Crash) Activate() {
}
