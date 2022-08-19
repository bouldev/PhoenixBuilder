package sunlife

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"time"

	"github.com/pterm/pterm"
)

type Lottery struct {
	*defines.BasicComponent

	DataOfLottery  map[string]*LotteryData
	Score          string                   `json:"货币计分板"`
	LotteryMenu    string                   `json:"抽奖菜单"`
	LotteryPool    map[string]*LotteryPools `json:"奖池(前面结构名字)"`
	PassWord       string                   `json:"激活密码"`
	Guarantees     int                      `json:"触发保底数"`
	LiteGuarantees []string                 `json:"小保底(会在下面的随机出结构名)"`
	BigGuarantees  string                   `json:"大保底(只出指定结构)"`
	Price          int                      `json:"抽一次的价格"`
}
type LotteryPools struct {
	Name       string `json:"奖品名字"`
	Guarantees int    `json:"概率(在0到100之间)"`
	Title      string `json:"抽中提示"`
}
type LotteryData struct {
	Num           int
	NowLotteryNum int
	Guarantees    int
}

func (b *Lottery) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	if b.PassWord != "985431651ASDFGHJKL" {
		pterm.Info.Println("[错误] 该组件激活码错误请在配置文件输入正确的激活码 否则请禁用该组件")
		panic("")
	}
	//初始化数据
	b.DataOfLottery = make(map[string]*LotteryData)

}
func (b *Lottery) Inject(frame defines.MainFrame) {
	b.Frame = frame
	b.BasicComponent.Inject(frame)
	CreateNameHash(b.Frame)
	b.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     []string{"抽奖系统"},
			ArgumentHint: " ",
			FinalTrigger: false,
			Usage:        "",
		},
		OptionalOnTriggerFn: b.Center,
	})
	//fmt.Println("-------", b.SnowsMenuTitle)
	b.Frame.GetJsonData("抽奖data.json", &b.DataOfLottery)
}
func (b *Lottery) Activate() {
	b.CmdSender(fmt.Sprintf("scoreboard objectives add %v dummy", b.Score))
}
func (b *Lottery) BackData(name string) {
	if _, ok := b.DataOfLottery[name]; !ok {
		b.DataOfLottery[name] = &LotteryData{
			Num:           0,
			NowLotteryNum: 0,
		}
	}
	msg := fmt.Sprintf("总抽奖次数:%v\n当前累计抽奖次数:%v\n保底数为:%v", b.DataOfLottery[name].Num, b.DataOfLottery[name].NowLotteryNum, b.Guarantees)
	b.sayto(name, msg)
}

func (b *Lottery) sayto(name string, msg string) {
	fmt.Println(msg)
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
}

// 根据次数抽奖
func (b *Lottery) LotteryByNum(name string, num int) string {
	rand.Seed(time.Now().Unix())
	_num := 0
	b.Frame.GetGameControl().SayTo("@a", "---------抽品列表----------\n中奖人:"+name+"\n奖品:\n")
	msg := ""
	for {
		prize := ""

		//如果次数超额则提前返回
		if _num >= num {
			b.Frame.GetGameControl().SayTo("@a", msg)
			b.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("\n累计次数:%v\n-------------------——", b.DataOfLottery[name].NowLotteryNum))
			return ""
		}
		//判断是否需要初始化
		if _, ok := b.DataOfLottery[name]; !ok {
			b.DataOfLottery[name] = &LotteryData{
				Num:           0,
				NowLotteryNum: 0,
			}
		}
		//查看是否保底
		if b.DataOfLottery[name].NowLotteryNum >= b.Guarantees {
			b.DataOfLottery[name].NowLotteryNum = 0
			randomNum := rand.Intn(len(b.LiteGuarantees))
			prize = b.LiteGuarantees[randomNum]
			//是否大保底
		} else if b.DataOfLottery[name].NowLotteryNum == b.Guarantees*2 {
			prize = b.BigGuarantees
			//否则就正常抽
		} else {
			lotterNum := 1
			for k, v := range b.LotteryPool {
				//如果最后一个还未抽中直接选择最后一个
				if lotterNum == len(b.LotteryPool) {
					prize = k
					break
				}
				if v.Guarantees >= 100 || v.Guarantees <= 0 {
					pterm.Info.Println("[错误！] 奖品的概率不能是<=0或者>=100的整数")
					b.sayto(name, "[错误！] 奖品的概率不能是<=0或者>=100的整数 请喊管理员修改配置文件")
					return ""
				}
				randomNum := rand.Intn(100)
				//如果是小于100的随机数则代表抽中
				if randomNum >= v.Guarantees {
					prize = k
					break
				}
				lotterNum++
			}
		}
		//判断奖品是否为奖品
		if _, ok := b.LotteryPool[prize]; !ok {
			pterm.Info.Printfln("不存在奖品名为:", prize, "的奖品")
			return ""
		}
		b.DataOfLottery[name].NowLotteryNum++
		b.DataOfLottery[name].Num++
		b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("execute @a[name=\"%v\"] ~~~ structure load %v ~~~", name, prize), func(output *packet.CommandOutput) {
			if !(output.SuccessCount > 0) {
				pterm.Info.Println("指令错误:错误信息为为:", output.OutputMessages, "\n错误指令为:"+fmt.Sprintf("execute @a[name=\"%v\"] ~~~ structure %v ~~~ ~~~", name, prize))
			}

		})
		msg = msg + fmt.Sprintf("%v/", prize)

		_num++
	}
	return ""

}

// 返回奖池
func (b *Lottery) backPool(name string) {
	msg := "-----------------"
	for k, v := range b.LotteryPool {
		msg = msg + fmt.Sprintf("\n奖品名字:%v 概率:%v\n----------------", k, v.Guarantees)

	}
	b.sayto(name, msg)
}
func (b *Lottery) CmdSender(str string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(str, func(output *packet.CommandOutput) {
		if !(output.SuccessCount > 0) {
			pterm.Info.Println("指令错误:错误信息为为:", output.OutputMessages, "\n错误指令为:"+str)
		}
	})
}
func (b *Lottery) Center(chat *defines.GameChat) bool {

	b.sayto(chat.Name, b.LotteryMenu)
	b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				go func() {
					b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players test @a[name=\"%v\"] %v %v *", chat.Name, b.Score, b.Price), func(output *packet.CommandOutput) {
						if output.SuccessCount > 0 {
							b.CmdSender(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", chat.Name, b.Score, b.Price))
							b.LotteryByNum(chat.Name, 1)
						} else {
							b.sayto(chat.Name, "[余额不足]")
						}

					})

				}()

			case "1":
				go func() {
					b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players test @a[name=\"%v\"] %v %v *", chat.Name, b.Score, b.Price*10), func(output *packet.CommandOutput) {
						if output.SuccessCount > 0 {
							b.CmdSender(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", chat.Name, b.Score, b.Price*10))
							b.LotteryByNum(chat.Name, 10)
						} else {
							b.sayto(chat.Name, "[余额不足]")
						}

					})

				}()

			case "2":
				b.BackData(chat.Name)
			case "3":
				b.backPool(chat.Name)
			}
		}
		return true
	})
	return false
}
