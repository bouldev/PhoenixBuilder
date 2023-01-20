package sunlife

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

type Enchant struct {
	*defines.BasicComponent
	Tirgger     []string                `json:"触发语"`
	Usage       string                  `json:"描述"`
	Menu        map[string]string       `json:"菜单"`
	EnchantPool map[string]*EnchantPool `json:"附魔"`
	Score       string                  `json:"购买所用计分板"`
	TitleWord   map[string]string       `json:"提示词"`
}
type EnchantPool struct {
	Name string           `json:"附魔商品名"`
	Rank map[string]*Rank `json:"等级"`
}
type Rank struct {
	Price       int      `json:"价格"`
	FCmds       string   `json:"购买前执行指令"`
	Scmds       []string `json:"购买后执行指令"`
	Title       string   `json:"购买前执行指令失败提示"`
	Probability float64  `json:"成功概率"`
}

func (b *Enchant) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}

}

// 注入
func (b *Enchant) Inject(frame defines.MainFrame) {
	b.Frame = frame
	b.BasicComponent.Inject(frame)

	b.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     b.Tirgger,
			ArgumentHint: " ",
			FinalTrigger: false,
			Usage:        b.Usage,
		},
		OptionalOnTriggerFn: b.Center,
	})

}

func (b *Enchant) Center(chat *defines.GameChat) bool {
	msg := ""
	num := 0
	list := make(map[string]string)
	for k, v := range b.EnchantPool {

		relist := map[string]interface{}{
			"i":   num,
			"附魔名": v.Name,
		}
		msg += FormateMsg(b.Frame, relist, b.Menu["主菜单模板"])
		list[strconv.Itoa(num)] = k
		num++
	}
	Sayto(b.Frame, chat.Name, msg)
	b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(Newchat *defines.GameChat) (catch bool) {
		if len(Newchat.Msg) > 0 {
			if enchantName, ok := list[Newchat.Msg[0]]; ok {
				msg = ""
				for k, v := range b.EnchantPool[enchantName].Rank {
					relist := map[string]interface{}{
						"等级":     k,
						"价格":     v.Price,
						"player": chat.Name,
					}
					msg += FormateMsg(b.Frame, relist, b.Menu["附魔子菜单模板"])
					num++
				}
				Sayto(b.Frame, chat.Name, msg)
				//购买
				b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
					//查看输入的是否为等级
					if data, isok := b.EnchantPool[enchantName].Rank[chat.Msg[0]]; isok {
						relist := map[string]interface{}{
							"player": chat.Name,
							"附魔":     enchantName,
							"i":      chat.Msg[0],
							"计分板":    b.Score,
							"价格":     data.Price,
						}
						fcmd := FormateMsg(b.Frame, relist, data.FCmds)
						b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fcmd, func(output *packet.CommandOutput) {
							if output.SuccessCount > 0 {
								//概率一下
								rand.Seed(time.Now().Unix())
								RandomNum := int(100 * data.Probability)
								if rand.Intn(100) <= RandomNum {
									//执行第二次指令
									for _, cmd := range data.Scmds {
										b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(FormateMsg(b.Frame, relist, cmd), func(output *packet.CommandOutput) {
											if output.SuccessCount > 0 {
												pterm.Info.Printfln("执行%v指令成功", FormateMsg(b.Frame, relist, cmd))
											} else {
												Sayto(b.Frame, chat.Name, fmt.Sprintf("执行%v指令失败 返回指令失败信息为:%v\n一般来说出现这种错误是因为你没有手持正确的武器", FormateMsg(b.Frame, relist, cmd), output.OutputMessages))
												pterm.Info.Printfln("执行%v指令失败 返回指令失败信息为:%v\n一般来说出现这种错误是因为你没有手持正确的武器", FormateMsg(b.Frame, relist, cmd), output.OutputMessages)
											}
										})
									}
									Sayto(b.Frame, chat.Name, b.TitleWord["附魔成功提示词"])
								} else {
									Sayto(b.Frame, chat.Name, b.TitleWord["附魔失败"])
								}

							} else {
								pterm.Info.Println("执行的前置指令为:", fcmd)
								Sayto(b.Frame, chat.Name, data.Title)
							}
						})
					} else {
						Sayto(b.Frame, chat.Name, b.TitleWord["无效输入"])
					}

					return true
				})
			} else {
				Sayto(b.Frame, chat.Name, b.TitleWord["无效输入"])
			}
		}

		return true
	})
	return true
}
