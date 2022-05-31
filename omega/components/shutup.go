package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"sync"
	"time"
)

type ShutUp struct {
	*BasicComponent
	Round          int `json:"发言检测周期(秒)"`
	MaxWords       int `json:"周期内字数限制"`
	MaxMsgs        int `json:"周期内条数限制"`
	mu             sync.Mutex
	statisticWords map[string]int
	statisticMsgs  map[string]int
	shutupCmds     []defines.Cmd
}

func (o *ShutUp) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	o.statisticMsgs = map[string]int{}
	o.statisticWords = map[string]int{}
	var err error
	o.shutupCmds, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"发言过快反制"})
	if err != nil {
		panic(err)
	}
}

func (o *ShutUp) shutup(player string) {
	o.Frame.GetBackendDisplay().Write(fmt.Sprintf("玩家 %v 发言过快，启动反制指令", player))
	utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.shutupCmds, map[string]interface{}{
		"[player]": "\"" + player + "\"",
	}, o.Frame.GetBackendDisplay())
}

func (o *ShutUp) counter(c *defines.GameChat) bool {
	if c.Name == o.Frame.GetUQHolder().GetBotName() {
		return false
	}
	if c.Name == "外部" {
		return false
	}
	o.mu.Lock()
	o.statisticMsgs[c.Name]++
	o.statisticWords[c.Name] += len(c.Msg)
	if o.statisticMsgs[c.Name] > o.MaxMsgs || o.statisticWords[c.Name] > o.MaxWords {
		o.shutup(c.Name)
	}
	o.mu.Unlock()
	return false
}

func (o *ShutUp) Activate() {
	o.Ctrl = o.Frame.GetGameControl()
	go func() {
		t := time.NewTicker(time.Second * time.Duration(o.Round))
		for {
			<-t.C
			o.mu.Lock()
			o.statisticMsgs = map[string]int{}
			o.statisticWords = map[string]int{}
			o.mu.Unlock()
		}
	}()
}
func (o *ShutUp) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetGameChatInterceptor(o.counter)
}
