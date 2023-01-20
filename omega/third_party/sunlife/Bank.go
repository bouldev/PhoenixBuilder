package sunlife

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"regexp"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

// 银行系统
type Bank struct {
	*defines.BasicComponent
	Triggers        []string                       `json:"触发词"`
	Use             string                         `json:"描述"`
	Menu            map[string]string              `json:"菜单提示类"`
	InvestRegularly map[string]*InvestRegularlyMap `json:"定期投资"`
	StableRegularly map[string]float64             `json:"稳定投资"`
	VentureCapital  map[string]*VentureCapitalMap  `json:"风险投资"`
	TitleWord       map[string]string              `json:"各种提示词"`
	Score           string                         `json:"银行所用计分板"`
	Data            map[string]*BankData
}

// 风险投资字典json
type VentureCapitalMap struct {
	Star  float64 `json:"波动起始"`
	Range float64 `json:"波动范围"`
}

// 银行数据
type BankData struct {
	InvestRegularly *InvestRegularlyData
	StableRegularly *StableRegularlyData
	VentureCapital  *VentureCapitalData
}

// 风险投资
type VentureCapitalData struct {
	Comon   string
	Score   int
	ReScore float64
	Day     int
}

// 未定投资
type StableRegularlyData struct {
	Day     int
	Combo   string
	Score   int
	ReScore float64
}

// 稳定投资
type InvestRegularlyData struct {
	Combo   string
	Score   int
	Day     int
	ReScore float64
}

// 稳定投资的字典提取json使用
type InvestRegularlyMap struct {
	Time     int     `json:"周期"`
	Interest float64 `json:"利率"`
}

func (o *Bank) Init(cfg *defines.ComponentConfig) {
	if cfg.Version == "0.0.1" {
		cfg.Configs["赞助人名单"] = []string{
			"萌宝宝",
			"像曹一样",
			"sunlife飘怅",
			"一只佳鑫酱",
			"天劫飛飛",
			"狗头梦幻",
			"Junio_炖鱼",
			"天空之城",
			"弑",
			"春风",
			"空中的风",
			"ACCA的会计师",
		}
		cfg.Configs["各种提示词"] = map[string]string{
			"输入存入金额提示":  "[请输入你要存取的金额]",
			"余额不足提示词":   "余额不足",
			"提现金额成功提示":  "成功取出[金额]\n余额:[余额]",
			"取出余额成功提示":  "成功取出[金额]\n余额:[余额]",
			"稳定投资取钱提示词": "当前余额:[余额]\n请输入要取出的数额",
			"取风险投资时提示":  "请输入取出金额\n(当前套餐:[套餐名],当前余额:[余额])",
			"还未到日期时提示":  "你当前套餐[套餐]\n日期还有:[日期]天",
		}
		cfg.Configs["菜单提示类"] = map[string]string{
			"主菜单提示":      "输入 0 定期投资\n输入 1 稳定投资\n输入 2 风险投资",
			"稳定投资存钱提示词":  "输入存入金额",
			"定期投资菜单模板":   "输入 [i] 购买[方案名字] 周期为:[周期] 利率为:[利率]",
			"定期投资菜单":     "输入 0 购买套餐\n输入 1 取出金额",
			"稳定投资菜单显示":   "输入 0 存入金额\n输入 1 取出金额",
			"风险投资菜单显示":   "输入 0 购买套餐\n输入 1 存入金额\n输入 2 取出金额",
			"购买风险投资套餐提示": "输入 [i] 购买[方案名字] 波动起点:[波动起点],波动范围:[波动范围]",
		}
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
	o.Data = make(map[string]*BankData)
}
func (o *Bank) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetJsonData("银行数据.json", &o.Data)
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        o.Use,
		},
		OptionalOnTriggerFn: o.BankCenter,
	})
}

// 正则表达检查字符串是否为数字
func (b *Bank) CheckIsNum(str string) bool {
	ok, _ := regexp.MatchString("^\\+?[1-9][0-9]*$", str)
	return ok
}
func (b *Bank) randFloats(min, max float64) float64 {

	res := min + rand.Float64()*(max-min)

	return res
}
func (b *Bank) Work() {
	for k, v := range b.Data {
		fmt.Println("test")
		//v.StableRegularly.Combo
		//每日零点加倍率
		if v.InvestRegularly == nil {
			b.Data[k].InvestRegularly = &InvestRegularlyData{}
		}
		if v.StableRegularly == nil {
			b.Data[k].StableRegularly = &StableRegularlyData{}
		}
		if v.VentureCapital == nil {
			b.Data[k].VentureCapital = &VentureCapitalData{}
		}
		//稳定投资
		if v.StableRegularly.Score > 0 {
			//如果存在套餐且分数不为0
			b.Data[k].StableRegularly.ReScore = float64(b.Data[k].StableRegularly.Score) + v.StableRegularly.ReScore*b.StableRegularly["利率"]
			b.Data[k].StableRegularly.Score = int(b.Data[k].StableRegularly.ReScore)
			pterm.Info.Printf("稳定投资:%v %v %v\n", k, b.Data[k].StableRegularly.ReScore, b.Data[k].StableRegularly.Score)
		}

		if _, ok := b.InvestRegularly[v.InvestRegularly.Combo]; ok && v.InvestRegularly.Score > 0 {
			b.Data[k].InvestRegularly.ReScore = float64(b.Data[k].InvestRegularly.Score) + b.InvestRegularly[v.InvestRegularly.Combo].Interest*v.InvestRegularly.ReScore
			b.Data[k].InvestRegularly.Score = int(b.Data[k].InvestRegularly.ReScore)
			pterm.Info.Printf("定期投资:%v %v %v\n", k, b.Data[k].InvestRegularly.ReScore, b.Data[k].InvestRegularly.Score)
		}
		if data, ok := b.VentureCapital[v.VentureCapital.Comon]; ok && v.VentureCapital.Score > 0 {
			interestRate := b.randFloats(data.Star, data.Star+data.Range)
			pterm.Info.Println("当天风险投资利率为:", interestRate)
			b.Data[k].VentureCapital.ReScore = float64(v.VentureCapital.Score) + v.VentureCapital.ReScore*interestRate
			b.Data[k].VentureCapital.Score = int(b.Data[k].VentureCapital.ReScore)

		}
	}

}
func (o *Bank) Activate() {
	for {
		now := time.Now()                                                                    //获取当前时间，放到now里面，要给next用
		next := now.Add(time.Hour * 24)                                                      //通过now偏移24小时
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()) //获取下一个凌晨的日期
		t := time.NewTimer(next.Sub(now))                                                    //计算当前时间到凌晨的时间间隔，设置一个定时器
		<-t.C
		o.Work()
		//以下为定时执行的操作

	}
}
func (b *Bank) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return b.Frame.WriteJsonDataWithTMP("银行数据.json", ".ckpt", &b.Data)
	}
	return nil
}

// 笨办法
func (b *Bank) formateMsg(str string, re string, afterstr string) (newstr string) {

	res := regexp.MustCompile("\\[" + re + "\\]")
	return res.ReplaceAllString(str, afterstr)

}

// 返回全部金额
func (b *Bank) GetNum(num int, sNum float64) int {
	return (num * int(sNum+1))
}

// 好点的
func (b *Bank) FormateMsg(list map[string]interface{}, msg string) string {

	for k, v := range list {
		msg = b.formateMsg(msg, k, fmt.Sprintf("%v", v))
	}
	return msg
}

// 命令发送者
func (b *Bank) CmdSender(cmd string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {

		} else {
			pterm.Info.Printfln(fmt.Sprintf("执行指令失败 错误原因:%v\n错误指令为%v", output.OutputMessages, cmd))
		}
	})
}

// 购买套餐
func (b *Bank) BuyInverstCombo(name string) {
	if _, ok := b.Data[name]; !ok {
		b.Data[name] = &BankData{}
		b.Data[name].InvestRegularly = &InvestRegularlyData{
			Combo:   "",
			Score:   0,
			Day:     time.Now().Day(),
			ReScore: 0,
		}
	}
	if b.Data[name].InvestRegularly == nil {
		b.Data[name].InvestRegularly = &InvestRegularlyData{
			Combo:   "",
			Score:   0,
			Day:     time.Now().Day(),
			ReScore: 0,
		}
	}
	//检查套餐是否为0

	if _, ok := b.InvestRegularly[b.Data[name].InvestRegularly.Combo]; !ok || b.Data[name].InvestRegularly.Score == 0 {
		b.InvestRegularlyMenu(name)
	} else if ok {
		b.sayto(name, "你套餐余额不为0不能购买其他套餐")
	} else {
		b.sayto(name, "没有套餐")
	}

}

// 定期投资菜单
func (b *Bank) InvestRegularlyCenter(name string) {
	b.sayto(name, b.Menu["定期投资菜单"])
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.BuyInverstCombo(chat.Name)
			case "1":
				b.getInvesetRegularlyMenu(chat.Name)
			}
		}
		return true
	})

}

// 购买套餐
func (b *Bank) InvestRegularlyMenu(name string) {
	msg := ""
	num := 0
	Dic := make(map[string]string)
	for k, v := range b.InvestRegularly {

		Dic[strconv.Itoa(num)] = k
		list := map[string]interface{}{
			"i":    num,
			"方案名字": k,
			"周期":   v.Time,
			"利率":   v.Interest,
		}
		msg = msg + b.FormateMsg(list, b.Menu["定期投资菜单模板"]) //
		msg = msg + "\n"
		num++
	}
	b.sayto(name, msg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if comboName, ok := Dic[chat.Msg[0]]; ok {
				b.sayto(name, b.TitleWord["输入存入金额提示"])
				b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
					if len(chat.Msg) > 0 && b.CheckIsNum(chat.Msg[0]) {
						Price := chat.Msg[0]

						//b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd)

						b.CheckMoney(name, Price, func(output *packet.CommandOutput) {
							if output.SuccessCount > 0 {
								price, _ := strconv.Atoi(chat.Msg[0])
								b.Data[name].InvestRegularly = &InvestRegularlyData{
									Combo: comboName,
									Score: price,
									Day:   time.Now().Day(),
								}
								b.sayto(name, "成功购买套餐")

							} else {
								b.sayto(name, b.TitleWord["余额不足提示词"])
								pterm.Info.Println(fmt.Sprintf("执行错误 错误原因:%v\n", output.OutputMessages))
							}
						})

					} else {
						b.sayto(name, "输入有效数字")
					}
					return true
				})
			} else {
				b.sayto(name, "请输入有效数字")
			}
		}
		return true
	})
}

// 检查余额是否充足 第二个参数为需要的钱
func (b *Bank) CheckMoney(name string, num string, DealWay func(output *packet.CommandOutput)) {
	cmd := fmt.Sprintf("scoreboard players remove @a[name=\"%v\",scores={\"%v\"=%v..}] %v %v *", name, b.Score, num, b.Score, num)
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, DealWay)
}

// 提现定期投资
func (b *Bank) getInvesetRegularlyMenu(name string) {
	if data, ok := b.Data[name]; ok {
		if data.InvestRegularly == nil {
			data.InvestRegularly = &InvestRegularlyData{}
		}
		if _, ok := b.InvestRegularly[data.InvestRegularly.Combo]; ok {
			if time.Now().Day() >= (data.InvestRegularly.Day + b.InvestRegularly[data.InvestRegularly.Combo].Time) {
				b.sayto(name, b.formateMsg(b.TitleWord["提现金额成功提示"], "price", strconv.Itoa(b.GetNum(data.InvestRegularly.Score, b.InvestRegularly[data.InvestRegularly.Combo].Interest))))
				b.CmdSender(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", name, b.Score, strconv.Itoa(b.GetNum(data.InvestRegularly.Score, b.InvestRegularly[data.InvestRegularly.Combo].Interest))))
				b.Data[name].InvestRegularly.Score = 0
				b.Data[name].InvestRegularly.Combo = ""
				b.Data[name].InvestRegularly.Day = 0
			} else {
				Relist := map[string]interface{}{
					"套餐": b.Data[name].InvestRegularly.Combo,
					"日期": time.Now().Day() - (b.Data[name].InvestRegularly.Day + b.InvestRegularly[data.InvestRegularly.Combo].Time),
				}

				b.sayto(name, b.FormateMsg(Relist, b.TitleWord["还未到日期时提示"]))
			}

		} else {
			b.sayto(name, "你没有套餐")
		}

	} else {
		b.sayto(name, "你没有存款")
	}
}

/*

稳定投资

*/

// 稳定投资的存取
func (b *Bank) StableInvestmentAccess(name string, num string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players remove @a[name=\"%v\",scores={%v=%v..}] %v %v", name, b.Score, num, b.Score, num), func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			price, _ := strconv.Atoi(num)
			if price <= 0 {
				price = price * -1
			}
			if _, ok := b.Data[name]; !ok {
				b.Data[name] = &BankData{}
				b.Data[name].StableRegularly = &StableRegularlyData{
					Day:     0,
					Score:   0,
					ReScore: 0,
				}
			}
			b.Data[name].StableRegularly.Score += price
			b.sayto(name, "存入成功")
		} else {
			b.sayto(name, b.TitleWord["余额不足提示词"])
			pterm.Info.Println(fmt.Sprintf("执行错误 错误原因:%v\n错误指令:%v", output.OutputMessages, fmt.Sprintf("scoreboard players remove @a[name=\"%v\",scores={%v=%v..}] %v %v", name, b.Score, num, b.Score, num)))
		}
	})
}

// 稳定取钱取钱
func (b *Bank) StableInvestmentTake(name string, num string) {
	cmd := fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", name, b.Score, num)
	price, _ := strconv.Atoi(num)
	if price <= b.Data[name].StableRegularly.Score {
		b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				if price <= 0 {
					price = price * -1
				}
				b.Data[name].StableRegularly.Score -= price
				//确保是最低
				b.Data[name].StableRegularly.ReScore = float64(b.Data[name].StableRegularly.Score)
				b.sayto(name, "取款成功")
			} else {
				b.sayto(name, fmt.Sprintf("执行错误 错误原因:%v\n错误指令:%v", output.OutputMessages, cmd))
				pterm.Info.Println(fmt.Sprintf("执行错误 错误原因:%v\n错误指令:%v", output.OutputMessages, cmd))
			}
		})
	} else {
		b.sayto(name, b.TitleWord["余额不足提示词"])
	}

}

// 取钱中心
func (b *Bank) StableInvestmentTakeCenter(name string) {
	if _, ok := b.Data[name]; !ok {
		b.Data[name] = &BankData{
			StableRegularly: &StableRegularlyData{},
		}
	}
	Relist := map[string]interface{}{
		"余额": b.Data[name].StableRegularly.Score,
	}
	msg := b.FormateMsg(Relist, b.TitleWord["稳定投资取钱提示词"])
	b.sayto(name, msg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 && b.CheckIsNum(chat.Msg[0]) {
			price, _ := strconv.Atoi(chat.Msg[0])
			if price > b.Data[name].StableRegularly.Score {
				b.sayto(name, b.TitleWord["余额不足提示词"])
			} else {
				b.StableInvestmentTake(name, chat.Msg[0])
			}

		} else {
			b.sayto(name, "输入有效数字")
		}
		return true
	})
}

// 稳定投资菜单
func (b *Bank) StableInvestmentCenter(name string) {
	b.sayto(name, b.Menu["稳定投资菜单显示"])
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.StableInvestmentMenu(name)
			case "1":
				b.StableInvestmentTakeCenter(name)
			}
		}
		return true
	})
}

// 稳定投资存入
func (b *Bank) StableInvestmentMenu(name string) {
	b.sayto(name, b.Menu["稳定投资存钱提示词"])
	//如果不存在则初始化
	if b.Data[name] == nil {
		b.Data[name] = &BankData{}
		b.Data[name].StableRegularly = &StableRegularlyData{
			Combo:   "",
			Score:   0,
			ReScore: 0,
			Day:     time.Now().Day(),
		}
	}
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 && b.CheckIsNum(chat.Msg[0]) {
			b.StableInvestmentAccess(name, chat.Msg[0])
		} else {
			b.sayto(name, "输入有效数字")
		}
		return true
	})
}

/*

风险投资

*/
// 风险投资存取（）
func (b *Bank) VentureCapitalAccess(name string, comboName string) {
	b.sayto(name, "[输入存入金额]")
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 && b.CheckIsNum(chat.Msg[0]) {
			b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", name, b.Score, chat.Msg[0]), func(output *packet.CommandOutput) {
				if output.SuccessCount > 0 {
					price, _ := strconv.Atoi(chat.Msg[0])
					if price <= 0 {
						price = price * -1
					}
					b.Data[name].VentureCapital.Score += price
					b.sayto(name, "存入成功")
				} else {
					b.sayto(name, b.TitleWord["余额不足提示词"])
					pterm.Info.Println(fmt.Sprintf("执行错误 错误原因:%v\n错误指令:%v", output.OutputMessages, fmt.Sprintf("scoreboard players remove @a[name=\"%v\",scores={%v=%v..}] %v %v", name, b.Score, chat.Msg[0], b.Score, chat.Msg[0])))
				}
			})
		} else {
			b.sayto(name, "输入有效数字")
		}
		return true
	})
}

// 购买风险套餐（）
func (b *Bank) BuyVentureCombo(name string) {
	//b.sayto(name, b.Menu["风险投资菜单模板"])
	num := 0
	list := make(map[string]string)
	msg := ""
	for k, v := range b.VentureCapital {
		msg = msg + b.Menu["购买风险投资套餐提示"]
		ReList := map[string]interface{}{
			"i":    num,
			"方案名字": k,
			"波动起点": v.Star,
			"波动范围": v.Range,
		}
		list[strconv.Itoa(num)] = k
		msg = b.FormateMsg(ReList, msg)
		msg = msg + "\n"
		num++
	}
	b.sayto(name, msg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if comboName, ok := list[chat.Msg[0]]; ok {
				if b.Data[name] == nil {
					b.Data[name] = &BankData{}
				}
				if b.Data[name].VentureCapital == nil {
					b.Data[name].VentureCapital = &VentureCapitalData{}
				}
				if _, isok := b.Data[name]; isok && (b.Data[name].VentureCapital.Comon == list[chat.Msg[0]]) {
					b.VentureCapitalAccess(name, comboName)
				} else if b.Data[name].VentureCapital.Score == 0 {
					//诺套餐为空则自动创建
					b.Data[name].VentureCapital.Comon = comboName
					b.Data[name].VentureCapital.ReScore = 0

					b.VentureCapitalAccess(name, comboName)
				} else {
					b.sayto(name, "你已经在其他套餐 \n如果需要转移套餐请取出所有钱")
				}

			}
		} else {
			b.sayto(name, "输入有效数字")
		}

		return true
	})
}

// 风险存钱
func (b *Bank) VentureCapitalDeposit(name string, comboName string) {
	b.sayto(name, "请输入存入金额")
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			b.CheckIsNum(chat.Msg[0])
			num := chat.Msg[0]
			cmd := fmt.Sprintf("scoreboard players remove @a[name=\"%v\",scores={%v=%v..}] %v %v", name, b.Score, num, b.Score, num)
			b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
				if output.SuccessCount > 0 {
					price, _ := strconv.Atoi(num)
					if price <= 0 {
						price = price * -1
					}
					b.Data[name].VentureCapital.Score += price
					b.sayto(name, "存入成功")
				} else {
					b.sayto(name, b.TitleWord["余额不足提示词"])
					pterm.Info.Println(fmt.Sprintf("执行错误 错误原因:%v\n错误指令:%v", output.OutputMessages, cmd))
				}
			})

		}

		return true
	})

}

// 风险投资存钱中心
func (b *Bank) VentureCapitalDepositCenter(name string) {
	if _, ok := b.Data[name]; !ok {
		//防止没有
		b.Data[name] = &BankData{}
		b.Data[name].VentureCapital = &VentureCapitalData{
			Comon:   "",
			Score:   0,
			ReScore: 0,
			Day:     time.Now().Day(),
		}

	}
	//如果没有套餐
	if _, ok := b.VentureCapital[b.Data[name].VentureCapital.Comon]; ok {
		b.VentureCapitalDeposit(name, b.Data[name].VentureCapital.Comon)
	} else {
		b.sayto(name, "暂时未购买套餐")
	}
}

// 风险取钱
func (b *Bank) VentureCapitalTake(name string) {
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if b.CheckIsNum(chat.Msg[0]) {
				price, _ := strconv.Atoi(chat.Msg[0])
				go func() {

					if price <= b.Data[name].VentureCapital.Score {
						b.Data[name].VentureCapital.Score -= price
						b.Data[name].VentureCapital.ReScore = float64(b.Data[name].VentureCapital.Score)
						Relist := map[string]interface{}{
							"金额": price,
							"余额": b.Data[name].VentureCapital.Score,
						}
						b.CmdSender(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", name, b.Score, price))
						b.sayto(name, b.FormateMsg(Relist, b.TitleWord["取出余额成功提示"]))
					} else {
						b.sayto(name, b.TitleWord["余额不足提示词"])
					}

				}()

			} else {
				b.sayto(name, "输入正确数字")
			}
		}

		return true
	})

}

// 风险取钱中心
func (b *Bank) VentureCapitalTakeCenter(name string) {
	if _, ok := b.Data[name]; !ok {
		b.Data[name] = &BankData{}
		b.Data[name].VentureCapital = &VentureCapitalData{
			Comon:   "",
			Score:   0,
			ReScore: 0,
			Day:     time.Now().Day(),
		}
	}
	if b.Data[name].VentureCapital == nil {
		b.Data[name].VentureCapital = &VentureCapitalData{}
	}
	if _, ok := b.VentureCapital[b.Data[name].VentureCapital.Comon]; ok {
		Relist := map[string]interface{}{
			"套餐名": b.Data[name].VentureCapital.Comon,
			"余额":  b.Data[name].VentureCapital.Score,
		}
		msg := b.FormateMsg(Relist, b.TitleWord["取风险投资时提示"])
		b.sayto(name, msg)
		b.VentureCapitalTake(name)
	} else {
		b.sayto(name, "没有购买套餐")
	}
}

// 风险投资中心
func (b *Bank) VentureCapitalCenter(name string) {
	b.sayto(name, b.Menu["风险投资菜单显示"])

	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.BuyVentureCombo(name)
			case "1":
				b.VentureCapitalDepositCenter(name)
			case "2":
				b.VentureCapitalTakeCenter(name)

			}
		}

		return true
	})
}

func (b *Bank) BankCenter(chat *defines.GameChat) (stop bool) {
	b.sayto(chat.Name, b.Menu["主菜单提示"])
	b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.InvestRegularlyCenter(chat.Name)
			case "1":
				b.StableInvestmentCenter(chat.Name)
			case "2":
				b.VentureCapitalCenter(chat.Name)
			case "3":
				//b.Work()
			}
		}

		return true
	})
	return true
}
func (b *Bank) sayto(name string, str string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), str)
	fmt.Println("SATTOstr:", str)
}
