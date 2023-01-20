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

type NewLottery struct {
	*defines.BasicComponent
	DataOfLottery   map[string]*NewLotteryData
	Tirgger         []string                `json:"触发词"`
	Usage           string                  `json:"描述"`
	Menu            map[string]string       `json:"菜单"`
	LotteryMainMenu string                  `json:"抽奖菜单"`
	Score           string                  `json:"货币计分板"`
	PassWord        string                  `json:"激活密码"`
	Pool            map[string]*LotteryPool `json:"奖池(前面结构名字)"`
	Price           int                     `json:"抽一次的价格"`
	TitleWord       map[string]string       `json:"提示词"`
	Guarantees      int                     `json:"触发保底数"`
	FGuaranteesType string                  `json:"小保底指定类型"`
	SGuaranteesType string                  `json:"大保底指定类型"`
	//随机数范围
	RandomRange int `json:"随机数范围"`
	//类型列表
	TypeList map[string]string
}
type LotteryPool struct {
	Name      string   `json:"奖品名字"`
	StarScore int      `json:"起始数字"`
	EndScore  int      `json:"结尾数字"`
	PoolType  string   `json:"类型"`
	Cmds      []string `json:"抽中执行指令"`
}

// 玩家数据
type NewLotteryData struct {
	//各个类型的累计数字
	NumDic map[string]int `json:"各个类型的累计数字"`
	//当前有效累计
	NowLotteryNum int `json:"当前有效累计"`
	//是否保底
	IsGuarantees bool `json:"是否小保底保底"`
}

func (b *NewLottery) Init(cfg *defines.ComponentConfig) {
	if cfg.Version == "0.0.1" {
		cfg.Configs["抽一次的价格"] = 200
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	} else if cfg.Version == "0.0.2" {
		cfg.Configs["触发词"] = []string{"抽奖"}
		cfg.Configs["描述"] = ""
		cfg.Configs["抽一次的价格"] = 200
		cfg.Version = "0.0.3"
		cfg.Upgrade()
	} else if cfg.Version == "0.0.3" {
		cfg.Configs["菜单"] = map[string]string{
			"奖池详细信息模板": "\n奖品名字:[奖品名字] 概率:[概率]\n-------------------",
			"抽奖头":      "---------奖品列表--------\n中奖人:[中奖人]\n奖品:",
			"抽奖尾":      "\n累计次数:[累计次数]\n--------------------------",
			"抽奖体":      "[奖品]/",
		}
		cfg.Version = "0.0.4"
		cfg.Upgrade()
	} else if cfg.Version == "0.0.4" {
		delete(cfg.Configs, "小保底(会在下面的随机出结构名)")
		delete(cfg.Configs, "大保底(只出指定结构)")
		PoolData := map[string]interface{}{
			"二号": map[string]interface{}{
				"奖品名字": "奖品二号",
				"起始数字": 0,
				"结尾数字": 10,
				"类型":   "大保底奖品",
				"抽中执行指令": []string{
					"execute @a[name=\"[player]\"] ~~~ structure load [奖品结构名字] ~~~",
					"title @a[name=\"[player]\"] title 抽中了 [奖品名字]",
				},
			},
			"三号": map[string]interface{}{
				"奖品名字": "奖品三号",
				"起始数字": 10,
				"结尾数字": 80,
				"类型":   "普通奖品",
				"抽中执行指令": []string{
					"execute @a[name=\"[player]\"] ~~~ structure load [奖品结构名字] ~~~",
					"title @a[name=\"[player]\"] title 抽中了 [奖品名字]",
				},
			},
			"四号": map[string]interface{}{
				"奖品名字": "奖品四号",
				"起始数字": 80,
				"结尾数字": 100,
				"类型":   "小保底奖品",
				"抽中执行指令": []string{
					"execute @a[name=\"[player]\"] ~~~ structure load [奖品结构名字] ~~~",
					"title @a[name=\"[player]\"] title 抽中了 [奖品名字]",
				},
			},
		}
		cfg.Configs["奖池(前面结构名字)"] = PoolData
		cfg.Configs["小保底指定类型"] = "小保底奖品"
		cfg.Configs["大保底指定类型"] = "大保底奖品"
		cfg.Configs["菜单"] = map[string]string{
			"奖池详细信息模板": "\n[奖品类型]:[奖品名字] 概率:[概率]\n-------------------",
			"抽奖头":      "---------奖品列表--------\n中奖人:[中奖人]\n奖品:",
			"累计显示":     "\n[抽奖类型]:[累计次数]",
			"抽奖尾部":     "当前有效抽奖数累计:[累计数],小保底数:90,大保底数为:180",
			"抽奖体":      "[奖品类型]:[奖品]/",
		}
		cfg.Configs["提示词"] = map[string]string{
			"余额不足提示词": "§c余额不足",
		}
		cfg.Configs["抽奖菜单"] = "输入 0 抽一次\n输入 1 十连\n输入 2 查看奖品池子"
		cfg.Configs["随机数范围"] = 100
		cfg.Version = "0.0.5"
		//cfg.Name = "sunlife抽奖"
		cfg.Upgrade()

	} else if cfg.Version == "0.0.5" {
		delete(cfg.Configs, "激活密码")
		cfg.Description = "免费的组件"
		cfg.Version = "0.0.6"
		cfg.Upgrade()
	}
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}

	//初始化数据

	b.TypeList = make(map[string]string)
	for _, v := range b.Pool {
		b.TypeList[v.PoolType] = "这是一个类型"
	}
	b.DataOfLottery = make(map[string]*NewLotteryData)

}
func (b *NewLottery) Inject(frame defines.MainFrame) {
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

	b.Frame.GetJsonData("新抽奖data.json", &b.DataOfLottery)
}

// 保存数据
func (b *NewLottery) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return b.Frame.WriteJsonDataWithTMP("新抽奖data.json", ".ckpt", &b.DataOfLottery)
	}
	return nil
}
func (b *NewLottery) Activate() {
	cmd := fmt.Sprintf("scoreboard objectives add %v dummy", b.Score)
	//b.CmdSender(fmt.Sprintf("scoreboard objectives add %v dummy", b.Score))
	//初始化计分板
	pterm.Info.Println("检查抽奖计分板是否存在")
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			pterm.Info.Println("不存在 已自动创立计分板" + b.Score)
		} else {
			pterm.Info.Printfln("已存在计分板 %v", b.Score)
		}
	})

}
func (b *NewLottery) Center(chat *defines.GameChat) bool {
	Sayto(b.Frame, chat.Name, b.LotteryMainMenu)
	b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.LotteryByNum(chat.Name, 1)
			case "1":
				b.LotteryByNum(chat.Name, 10)
			case "2":
				b.getBackPoolData(chat.Name)

			}
		}
		return true
	})
	return true
}
func (b *NewLottery) getBackPoolData(name string) {
	for _, v := range b.Pool {
		relist := map[string]interface{}{
			"奖品类型": v.PoolType,
			"奖品名字": v.Name,
			"概率":   (float64(v.EndScore) - float64(v.StarScore)) / float64(b.RandomRange),
		}
		Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["奖池详细信息模板"]))
	}
}
func (b *NewLottery) LotteryByNum(name string, num int) {
	if b.DataOfLottery[name] == nil {
		b.DataOfLottery[name] = &NewLotteryData{
			NumDic:        make(map[string]int),
			NowLotteryNum: 0,
			IsGuarantees:  false,
		}
	}
	price := b.getPrice(num)
	go func() {
		PlayerScoreDic := <-GetScore(b.Frame)
		if PlayerScoreDic[name][b.Score] < price {
			Sayto(b.Frame, name, b.TitleWord["余额不足提示词"])
		} else {
			//扣钱
			b.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", name, b.Score, price))
			relist := map[string]interface{}{
				"中奖人": name,
			}
			Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["抽奖头"]))
			//循环固定次数
			msg := ""
			for i := 0; i < num; i++ {
				rand.Seed(time.Now().Unix() + int64(i))
				randomNum := rand.Intn(b.RandomRange)
				prizeName := b.FindPrizeByNum(randomNum)
				//检查是否保底
				if data, isok := b.DataOfLottery[name]; isok {
					if data.NowLotteryNum == b.Guarantees {
						for k, v := range b.Pool {
							if v.PoolType == b.FGuaranteesType {
								prizeName = k
							}
						}
					} else if data.NowLotteryNum == b.Guarantees*2 {
						for k, v := range b.Pool {
							if v.PoolType == b.SGuaranteesType {
								prizeName = k
							}
						}
					}

				} else {
					pterm.Info.Println("请确保你全部抽奖品的随机范围能占满你的总随机数范围 \n现在已经遇到无法匹配的结果 已自动跳过该次抽奖")
					Sayto(b.Frame, name, "请确保你全部抽奖品的随机范围能占满你的总随机数范围 \n现在已经遇到无法匹配的结果 已自动跳过该次抽奖")
					break
				}
				//判断保底类型
				switch b.Pool[prizeName].PoolType {
				case b.FGuaranteesType:
					//小保底
					b.DataOfLottery[name].IsGuarantees = true
					b.DataOfLottery[name].NumDic[b.Pool[prizeName].PoolType] += 1
					b.DataOfLottery[name].NowLotteryNum = b.Guarantees + 1
				case b.SGuaranteesType:
					//大保底
					b.DataOfLottery[name].IsGuarantees = false
					b.DataOfLottery[name].NumDic[b.Pool[prizeName].PoolType] += 1
					b.DataOfLottery[name].NowLotteryNum = 0
				default:
					//常规情况
					b.DataOfLottery[name].NumDic[b.Pool[prizeName].PoolType] += 1
					b.DataOfLottery[name].NowLotteryNum++
					//执行指令

				}
				//执行指令
				for _, v := range b.Pool[prizeName].Cmds {
					relist = map[string]interface{}{
						"player": name,
						"奖品结构名字": prizeName,
						"奖品名字":   b.Pool[prizeName].Name,
					}
					b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(FormateMsg(b.Frame, relist, v), func(output *packet.CommandOutput) {
						pterm.Info.Printfln("已执行指令%v 结果如下:%v", v, output.OutputMessages)
					})
				}
				//显示
				relist = map[string]interface{}{
					"奖品类型": b.Pool[prizeName].PoolType,
					"奖品":   b.Pool[prizeName].Name,
				}
				msg += FormateMsg(b.Frame, relist, b.Menu["抽奖体"])
			}
			//发送抽奖体
			Sayto(b.Frame, name, msg)
			//发送抽奖累计次数
			msg = ""
			for k, _ := range b.TypeList {
				//如果既在列表中存在 也在玩家信息中存在则参与循环
				if num, ok := b.DataOfLottery[name].NumDic[k]; ok {
					relist = map[string]interface{}{
						"抽奖类型": k,
						"累计次数": num,
					}
					msg += FormateMsg(b.Frame, relist, b.Menu["累计显示"])
				}
			}
			Sayto(b.Frame, name, msg)
			//发送抽奖尾部
			relist = map[string]interface{}{
				"累计数": b.DataOfLottery[name].NowLotteryNum,
			}
			Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["抽奖尾部"]))
		}
	}()
}

// 获取价格
func (b *NewLottery) getPrice(num int) int {
	return b.Price * num
}

// 根据随机数字找到对应的奖品
func (b *NewLottery) FindPrizeByNum(num int) string {
	for k, v := range b.Pool {
		if num >= v.StarScore && num < v.EndScore {
			return k
		}
	}
	return ""
}
