package yscore

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"regexp"
	"time"
)

type Talk struct {
	*defines.BasicComponent
	Talkmap        map[string][]string `json:"精准触正则表达式与触发指令"`
	VagueAnswer    map[string][]string `json:"模糊回答"`
	VagueQuestion  map[string][]string `json:"模糊问题"`
	ComponentsName string
}

func (b *Talk) Init(cfg *defines.ComponentConfig) {

	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}

	b.ComponentsName = cfg.Name
}
func (b *Talk) Inject(frame defines.MainFrame) {
	b.Frame = frame

	b.BasicComponent.Inject(frame)
	b.Listener.SetGameChatInterceptor(b.Talker)
	//fmt.Println("-------", b.SnowsMenuTitle)

	//(*frame.GetContext())[collaborate.INTERFACE_FB_USERNAME] = b.FbNameSearcher
	//frame.QuerySensitiveInfo(defines.SENSITIVE_INFO_USERNAME_HASH)
	CreateNameHash(b.Frame)
}
func (b *Talk) Activate() {
	//监听并处理

}
func (b *Talk) Talker(chat *defines.GameChat) bool {
	if len(chat.Msg) > 0 {
		if b.PreciseDialogue(chat) == false {
			b.FuzzyCommunication(chat)
		}
	}
	return false
}

// 精准触发
func (b *Talk) PreciseDialogue(chat *defines.GameChat) bool {
	//先遍历所有消息
	for _, v := range chat.Msg {
		//i是问题

		for i, h := range b.Talkmap {

			if b.IsFinded(v, i) {

				for _, cmd := range h {
					cmd = b.FormateMsg(cmd, "触发对象名字", chat.Name)
					b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
						if output.SuccessCount > 0 {

						} else {
							fmt.Println("[执行错误错误] 指令报错如下", output.OutputMessages, "\n指令为:", cmd)
						}

					})

				}
				return true
			}
		}
	}
	return false
}

// 返回是否是模糊回答
func (b *Talk) FuzzyCommunication(chat *defines.GameChat) bool {
	//先在模糊问答中找关键词列表再迭代
	for k, v := range b.VagueQuestion {
		//从关键词列表中找关键词
		for _, j := range v {
			//如果关键词符合则随机回复对应的列表
			for _, msg := range chat.Msg {
				if b.IsFinded(msg, j) {
					if reply, ok := b.VagueAnswer[k]; ok {
						rand.Seed(time.Now().UnixNano())
						randomNum := rand.Intn(len(reply))
						ReReply := reply[randomNum]
						ReReply = b.FormateMsg(ReReply, "触发对象名字", chat.Name)
						b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(ReReply, func(output *packet.CommandOutput) {
							if output.SuccessCount > 0 {

							} else {
								fmt.Println("[执行错误错误] 指令报错如下", output.OutputMessages, "\n指令为:", ReReply)
							}
						})
						return true
					}
				}
			}
		}

	}
	return false
}

// 正则匹配是否能在msg中找到re
func (b *Talk) IsFinded(msg string, re string) bool {
	//fmt.Println("开始寻找")
	ok, _ := regexp.MatchString(msg, re)
	return ok
}

// 格式化信息
func (b *Talk) FormateMsg(str string, re string, afterstr string) (newstr string) {

	res := regexp.MustCompile("\\[" + re + "\\]")
	return res.ReplaceAllString(str, afterstr)

}
