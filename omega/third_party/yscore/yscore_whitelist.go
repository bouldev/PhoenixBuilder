package yscore

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strconv"
	"time"
)

type WhiteList struct {
	*defines.BasicComponent
	DelayTime          int      `json:"延迟检测时间(秒)"`
	Triggers           []string `json:"触发语"`
	Usage              string   `json:"描述语"`
	WhiteListOfDefines []string `json:"默认白名单"`
	KicktTitle         string   `json:"踢出时提示语"`
	Data               *List
}
type List struct {
	WhiteData map[string]string
	Pending   map[string]string
	BlackList map[string]string
}

func (b *WhiteList) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}

	b.Data = &List{
		WhiteData: make(map[string]string),
		Pending:   make(map[string]string),
		BlackList: make(map[string]string),
	}

	//如果为空则初始化

}
func (b *WhiteList) Inject(frame defines.MainFrame) {
	b.Frame = frame
	b.BasicComponent.Inject(frame)
	b.Listener.AppendLoginInfoCallback(b.onLogin)
	CreateNameHash(b.Frame)
	b.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     b.Triggers,
			ArgumentHint: " ",
			FinalTrigger: false,
			Usage:        b.Usage,
		},
		OptionalOnTriggerFn: b.AddWhiteList,
	})
	b.Frame.GetJsonData("白名单.json", &b.Data)
}
func (b *WhiteList) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return b.Frame.WriteJsonDataWithTMP("白名单.json", ".ckpt", &b.Data)
	}
	return nil
}
func (b *WhiteList) sayto(name string, str string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), str)
}
func (b *WhiteList) checkInMap(list map[string]string, name string) bool {
	for _, v := range list {
		if name == v {
			return true
		}
	}
	return false
}

// 加入白名单
func (b *WhiteList) AddWhiteList(chat *defines.GameChat) bool {
	go func() {
		list := <-b.GetPlayerName("@a[m=1]")
		fmt.Println(list)
		if b.checkInMap(list, chat.Name) {
			b.sayto(chat.Name, "白名单菜单 输入对应号数 处理对应玩家 输入 allow则全部同意\n输入 deny则拒绝")
			list := make(map[string]string)
			msg := ""
			num := 0
			for k, _ := range b.Data.Pending {
				list[strconv.Itoa(num)] = k
				msg = msg + strconv.Itoa(num) + " " + k + "\n"
				num++
			}
			b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(Newchat *defines.GameChat) (catch bool) {
				if len(Newchat.Msg) > 0 {
					if username, ok := list[chat.Msg[0]]; ok {
						b.sayto(Newchat.Name, "输入 0 加入白名单\n输入 1 加入黑名单")
						b.Frame.GetGameControl().SetOnParamMsg(Newchat.Name, func(chat *defines.GameChat) (catch bool) {
							if len(chat.Msg) > 0 {
								switch chat.Msg[0] {
								case "0":
									b.Data.WhiteData[username] = "白名单"
									delete(b.Data.Pending, username)
									b.sayto(chat.Name, "[成功加入白名单]")
								case "1":
									b.Data.BlackList[username] = "黑名单"
									delete(b.Data.Pending, username)
									b.sayto(chat.Name, "[成功加入黑名单]")
								case "allow":
									for k, _ := range b.Data.Pending {
										delete(b.Data.Pending, k)
										b.Data.WhiteData[k] = "白名单"
									}
								case "deny":
									for k, _ := range b.Data.Pending {
										delete(b.Data.Pending, k)
										b.Data.BlackList[k] = "黑名单"
									}

								}
							}

							return true
						})
					}
				}

				return true
			})

		} else {
			b.sayto(chat.Name, "[权限不足 需要创造]")
		}
	}()

	return false
}
func (b *WhiteList) checkInWhite(name string) bool {
	for _, v := range b.WhiteListOfDefines {
		if v == name {
			return true
		}
	}
	return false
}

// 上线处理
func (b *WhiteList) onLogin(entry protocol.PlayerListEntry) {
	go func() {
		time.Sleep(time.Second * time.Duration(b.DelayTime))
		if _, ok := b.Data.WhiteData[entry.Username]; ok {

		} else if _, ok := b.Data.BlackList[entry.Username]; ok {
			b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("kick \"%v\" %v", entry.Username, b.KicktTitle), func(output *packet.CommandOutput) {
				if output.SuccessCount > 0 {
					b.Frame.GetGameControl().SayTo("@a", "成功踢出非黑名单玩家 可在待处理里面处理该玩家")
				}
			})

		} else if b.Frame.GetUQHolder().GetBotName() == entry.Username {

		} else if b.checkInWhite(entry.Username) {

		} else {
			b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("kick \"%v\" %v", entry.Username, b.KicktTitle), func(output *packet.CommandOutput) {
				if output.SuccessCount > 0 {
					b.Data.Pending[entry.Username] = "这是一个处理"
					b.Frame.GetGameControl().SayTo("@a", "成功踢出非白名单玩家 可在待处理里面处理该玩家")
				}
			})
		}
	}()
}

// 获取指定限制器的玩家名字 返回通道值 key 为数字 v为玩家
func (b *WhiteList) GetPlayerName(name string) (listChan chan map[string]string) {
	type User struct {
		Name []string `json:"victim"`
	}
	var Users User
	//var UsersListChan chan []string
	UsersListChan := make(chan map[string]string, 2)
	//OkChan := make(chan bool, 2)
	//fmt.Print("test")
	//isok := false
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse("testfor "+name, func(output *packet.CommandOutput) {
		//fmt.Print(",,,,,,,,,,,,,,,,,,")
		//fmt.Print(output.DataSet)
		if output.SuccessCount > 0 {
			json.Unmarshal([]byte(output.DataSet), &Users)

			//var mapName map[string]string
			//fmt.Print("Users:", Users)
			mapName := make(map[string]string, 40)
			for k, v := range Users.Name {
				mapName[strconv.Itoa(k)] = v
			}

			//isok = true
			//fmt.Print("isok:", isok)
			UsersListChan <- mapName
			//OkChan <- true
		}

	})

	//fmt.Print("isok:", isok)
	return UsersListChan
}
