package sunlife

import (
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"

	"github.com/pterm/pterm"
)

type CustomCmd struct {
	*defines.BasicComponent
	Task map[string][]string `json:"任务"`
}

func (b *CustomCmd) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}

}

// 注入
func (b *CustomCmd) Inject(frame defines.MainFrame) {
	b.Frame = frame
	b.BasicComponent.Inject(frame)
	b.Frame.GetGameListener().SetGameChatInterceptor(b.onChat)
}
func (b *CustomCmd) onChat(chat *defines.GameChat) (stop bool) {
	if len(chat.Msg) > 0 {
		if data, ok := b.Task[chat.Msg[0]]; ok {
			go func() {
				playerPos := <-GetPos(b.Frame, "@a")
				relist := map[string]interface{}{
					"player": chat.Name,
					"x":      playerPos[chat.Name][0],
					"y":      playerPos[chat.Name][1],
					"z":      playerPos[chat.Name][2],
				}
				for _, _cmd := range data {
					cmd := FormateMsg(b.Frame, relist, _cmd)
					b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
						if !(output.SuccessCount > 0) {
							pterm.Info.Println("执行%v指令失败\n错误信息为: %v", cmd, output.OutputMessages)
						}
					})
				}

			}()

			return true
		}
	}
	return false
}
