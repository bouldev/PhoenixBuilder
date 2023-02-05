package example

import (
	"encoding/json"
	"phoenixbuilder/omega/defines"
	"strings"
	"time"
)

type EchoMiao struct {
	*defines.BasicComponent
	Suffix   string   `json:"附加的文本"`
	Triggers []string `json:"菜单触发词"`
}

func (o *EchoMiao) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
}

func (o *EchoMiao) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetGameChatInterceptor(o.onChat)
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[参数喵]",
			FinalTrigger: false,
			Usage:        "示例菜单项",
		},
		OptionalOnTriggerFn: o.onMenu,
	})
}

func (o *EchoMiao) onChat(chat *defines.GameChat) (stop bool) {
	o.Frame.GetGameControl().SayTo(chat.Name, strings.Join(chat.Msg, " ")+o.Suffix)
	// 返回 false 代表让后续组件继续处理这个信息
	// 返回 true 则后续组件都看不到这个信息
	return false
}

func (o *EchoMiao) onMenu(chat *defines.GameChat) (stop bool) {
	if len(chat.Msg) == 0 {
		// 向用户请求参数
		o.Frame.GetGameControl().SayTo(chat.Name, "喵~ ?")
		o.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
			o.interact(chat.Name, chat.Msg)
			return true
		})
	} else {
		o.interact(chat.Name, chat.Msg)
	}
	// 此处一般为 true 只有很少的情况为 false (菜单被误触)
	return true
}

func (o *EchoMiao) interact(name string, msg []string) {
	o.Frame.GetGameControl().SayTo(name, "喵喵! "+strings.Join(msg, " ")+o.Suffix)
}

// 所有组件 Inject 之后，会调用 BeforeActivate，在这个函数里可以去寻找其他组件的接口了，因为注入接口的过程是在 Inject 中完成的
// func (o *EchoMiao) BeforeActivate() {

// }

// 这个函数会在一个单独的协程中运行，可以自由的 sleep 或者阻塞
func (o *EchoMiao) Activate() {
	time.Sleep(1 * time.Second)
	o.Frame.GetBackendDisplay().Write("喵喵喵！")
}

// 如果需要处理主框架发来的信号
// func (o *EchoMiao) Signal(signal int) error {
// 	switch signal {
// 	case defines.SIGNAL_DATA_CHECKPOINT:
// 		if o.fileChange {
// 			o.fileChange = false
// 			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", &o.data)
// 		}
// 	}
// 	return nil
// }

// 如果需要在被正常关闭时执行某些功能
// func (o *EchoMiao) Stop() error {
// 	fmt.Println("正在保存 " + o.FileName)
// 	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", &o.data)
// }
