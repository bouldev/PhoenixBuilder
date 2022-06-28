package components

import (
	"encoding/json"
	"fmt"
	"math"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

type PlayerShopDataGood struct {
	Src           string           `json:"售卖人"`
	UID           string           `json:"UUID"`
	Name          string           `json:"物品名"`
	TimeStamp     string           `json:"售卖时间"`
	StructureName string           `json:"结构方块名"`
	Price         int              `json:"价格"`
	Currency      defines.Currency `json:"货币"`
}

type PlayerShopData struct {
	DeferedCmds map[string][]*defines.CmdsWithName `json:"下次上线时执行的指令"`
	Goods       map[string]*PlayerShopDataGood     `json:"在售物品"`
}

type PlayerShop struct {
	*BasicComponent
	fileData                 *PlayerShopData
	fileChange               bool
	FileName                 string   `json:"数据文件"`
	LoginDelay               int      `json:"登录时延迟发送"`
	BuyTriggers              []string `json:"购买触发词"`
	BuyUsage                 string   `json:"购买提示信息"`
	SaleTriggers             []string `json:"售卖触发词"`
	SaleUsage                string   `json:"售卖提示信息"`
	PreExecuteCmdsOnSale     []defines.Cmd
	PreExecuteCmdsOnSaleIn   interface{} `json:"打开售卖菜单前执行的指令"`
	SaleAuthSelector         string      `json:"售卖条件选择器"`
	HintOnCannotSale         []defines.Cmd
	HintOnCannotSaleIn       interface{} `json:"没有售卖权限时执行的指令"`
	HintOnCanSaleStep1       []defines.Cmd
	HintOnCanSaleStep1In     interface{}         `json:"售卖指导指令1"`
	DefaultCurrency          *defines.Currency   `json:"默认货币"`
	AllCurrency              []*defines.Currency `json:"可选择的货币"`
	HintOnCanSaleStep2       []defines.Cmd
	HintOnCanSaleStep2In     interface{} `json:"售卖指导指令2"`
	HintOnCanSaleStep3       []defines.Cmd
	HintOnCanSaleStep3In     interface{} `json:"售卖指导指令3"`
	StructureBlockNamePrefix string      `json:"结构方块前导名"`
	PackagePlatform          []int       `json:"打包平台"`
	SelectCmd                string      `json:"物品转移器"`
	Tax                      float64     `json:"所得税"`
	HintOnSale               []defines.Cmd
	HintOnSaleIn             interface{} `json:"放入玩家商店时的指令"`
	MaxGoodsPerPage          int         `json:"每页最多显示商品数"`
	GoodsListHead            string      `json:"物品清单开头"`
	GoodsListNextPage        string      `json:"还有下一页时末尾"`
	GoodsListEnd             string      `json:"物品清单结尾提示"`
	GoodEntryFmt             string      `json:"物品渲染格式"`
	FinalCmds                []defines.Cmd
	FinalCmdsIn              interface{} `json:"成功售出时的加钱指令"`
	mu                       sync.RWMutex
}

func (o *PlayerShop) executeCmds(player string, mapping map[string]interface{}, cmds []defines.Cmd) (success bool) {
	resultChan := make(chan bool)
	utils.GetPlayerList(o.Frame.GetGameControl(), "@a[name=\""+player+"\"]", func(s []string) {
		if len(s) == 0 {
			resultChan <- false
		} else {
			utils.LaunchCmdsArray(o.Frame.GetGameControl(), cmds, mapping, o.Frame.GetBackendDisplay())
			resultChan <- true
		}
	})
	return <-resultChan
}

func (o *PlayerShop) executeCmdsWithDefer(player, uid string, cmds []defines.Cmd, mapping map[string]interface{}) {
	go func() {
		if o.executeCmds(player, mapping, cmds) {
		} else {
			o.mu.Lock()
			if _, hasK := o.fileData.DeferedCmds[uid]; hasK {
				o.fileData.DeferedCmds[uid] = append(o.fileData.DeferedCmds[uid], &defines.CmdsWithName{Name: player, Cmds: cmds, Params: mapping})
			} else {
				o.fileData.DeferedCmds[uid] = []*defines.CmdsWithName{&defines.CmdsWithName{Name: player, Cmds: cmds, Params: mapping}}
			}
			o.mu.Unlock()
			o.fileChange = true
		}
	}()
}

func (o *PlayerShop) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	o.mu = sync.RWMutex{}
	if len(o.PackagePlatform) != 3 {
		panic(fmt.Errorf("打包平台 %v 坐标无效，应该为  [x,y,z] ", o.PackagePlatform))
	}
}

func (o *PlayerShop) formatPackage(idx int) string {
	return fmt.Sprintf(o.StructureBlockNamePrefix+"%v", idx)
}

func (o *PlayerShop) acquireEmptyNumber() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for i := 1; ; i++ {
		pkgName := o.formatPackage(i)
		if _, hasK := o.fileData.Goods[pkgName]; !hasK {
			return i
		}
	}
}

func (o *PlayerShop) Inject(frame defines.MainFrame) {
	o.Frame = frame
	err := frame.GetJsonData(o.FileName, &o.fileData)
	if err != nil {
		panic(err)
	}
	if o.fileData == nil {
		o.fileData = &PlayerShopData{}
	}
	if o.fileData.DeferedCmds == nil {
		o.fileData.DeferedCmds = make(map[string][]*defines.CmdsWithName)
	}
	if o.fileData.Goods == nil {
		o.fileData.Goods = make(map[string]*PlayerShopDataGood)
	}
	if o.PreExecuteCmdsOnSale, err = utils.ParseAdaptiveCmd(o.PreExecuteCmdsOnSaleIn); err != nil {
		panic("打开售卖菜单前执行的指令" + err.Error())
	}
	if o.HintOnCannotSale, err = utils.ParseAdaptiveCmd(o.HintOnCannotSaleIn); err != nil {
		panic("没有售卖权限时执行的指令" + err.Error())
	}
	if o.HintOnCanSaleStep1, err = utils.ParseAdaptiveCmd(o.HintOnCanSaleStep1In); err != nil {
		panic("售卖指导指令1" + err.Error())
	}
	if o.HintOnCanSaleStep2, err = utils.ParseAdaptiveCmd(o.HintOnCanSaleStep2In); err != nil {
		panic("售卖指导指令2" + err.Error())
	}
	if o.HintOnCanSaleStep3, err = utils.ParseAdaptiveCmd(o.HintOnCanSaleStep3In); err != nil {
		panic("售卖指导指令2" + err.Error())
	}
	if o.HintOnSale, err = utils.ParseAdaptiveCmd(o.HintOnSaleIn); err != nil {
		panic("放入玩家商店时的指令" + err.Error())
	}
	if o.FinalCmds, err = utils.ParseAdaptiveCmd(o.FinalCmdsIn); err != nil {
		panic("成功售出时的加钱指令" + err.Error())
	}
	o.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		name := utils.ToPlainName(entry.Username)
		uid := entry.UUID.String()
		o.mu.RLock()
		if deferCmds, hasK := o.fileData.DeferedCmds[uid]; hasK {
			o.mu.RUnlock()
			timer := time.NewTimer(time.Duration(o.LoginDelay) * time.Second)
			go func() {
				<-timer.C
				flag := true
				for cmd_i, cmd := range deferCmds {
					cmd.Params["[出售者]"] = "\"" + name + "\""
					if !o.executeCmds(name, cmd.Params, cmd.Cmds) {
						if cmd_i != 0 {
							o.mu.Lock()
							o.fileData.DeferedCmds[uid] = deferCmds[cmd_i:]
							o.mu.Unlock()
						}
						flag = false
						break
					}
				}
				if flag {
					o.mu.Lock()
					delete(o.fileData.DeferedCmds, uid)
					o.fileChange = true
					o.mu.Unlock()
				}
			}()
		} else {
			o.mu.RUnlock()
		}
	})
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.BuyTriggers,
			ArgumentHint: "",
			Usage:        o.BuyUsage,
			FinalTrigger: false,
		},
		OptionalOnTriggerFn: o.onBuyTrigger,
	})
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.SaleTriggers,
			ArgumentHint: "",
			Usage:        o.SaleUsage,
			FinalTrigger: false,
		},
		OptionalOnTriggerFn: o.onSaleTrigger,
	})
}

func (o *PlayerShop) showGoods(player string, goods []*PlayerShopDataGood, currentI int) {
	pk := o.Frame.GetGameControl().GetPlayerKit(player)
	pk.Say(o.GoodsListHead)
	i := currentI
	availableAction := map[string]string{}
	for i = currentI; i < len(goods) && i < currentI+o.MaxGoodsPerPage; i++ {
		l := utils.FormatByReplacingOccurrences(o.GoodEntryFmt, map[string]interface{}{
			"[I]":      i + 1,
			"[商品]":     goods[i].Name,
			"[出售者]":    "\"" + goods[i].Src + "\"",
			"[价格]":     goods[i].Price,
			"[货币名]":    goods[i].Currency.CurrencyName,
			"[货币计分板名]": goods[i].Currency.ScoreboardName,
		})
		availableAction[fmt.Sprintf("%v", i+1)] = goods[i].StructureName
		pk.Say(l)
	}
	if i == len(goods) {
		pk.Say(o.GoodsListEnd)
	} else {
		pk.Say(o.GoodsListNextPage)
	}
	hint := fmt.Sprintf("输入想购买的商品序号[%v-%v]", currentI+1, i)
	if currentI != 0 {
		hint += ", 输入 p 查看上一页"
	}
	if i != len(goods) {
		hint += ", 输入 n 查看下一页"
	}
	hint += ", 输入取消以取消购买"
	pk.Say(hint)
	pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) < 1 {
			pk.Say("无法理解的输入")
			return true
		}
		m := chat.Msg[0]
		if m == "取消" {
			pk.Say("已取消")
			return true
		}
		if m == "p" && currentI != 0 {
			newI := currentI - o.MaxGoodsPerPage
			if newI < 0 {
				newI = 0
			}
			go o.showGoods(player, goods, newI)
			return true
		}
		if m == "n" && i != len(goods) {
			go o.showGoods(player, goods, currentI+o.MaxGoodsPerPage)
			return true
		}
		if goodId, hasK := availableAction[m]; hasK {
			o.buyGood(player, goodId)
			return true
		} else {
			pk.Say("无法理解的输入")
			return true
		}
	})
}

func (o *PlayerShop) startBuy(player, goodId string, good *PlayerShopDataGood) {
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(
		fmt.Sprintf("scoreboard players add \"%v\" %v 0", player, good.Currency.ScoreboardName), func(output *packet.CommandOutput) {
			//fmt.Println(output)
			if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 4 {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("购买时发现玩家没有记分板%v %v", good.Currency.ScoreboardName, output))
				o.Frame.GetGameControl().SayTo(player, "似乎没有相关记分板，或者你这个记分板没有分数")
				return
			}
			hasMoney, err := strconv.Atoi(output.OutputMessages[0].Parameters[3])
			if err != nil {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("购买时解析出错 %v", err))
				o.Frame.GetGameControl().SayTo(player, "出于未知原因，无法完成购买")
				return
			}
			if hasMoney > good.Price {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("玩家 %v 花费 %v / %v 购买了 %v %v", player, good.Price, hasMoney, goodId, good))
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreboard players remove \"%v\" %v %v", player, good.Currency.ScoreboardName, good.Price))
				cmd := fmt.Sprintf("execute \"%v\" ~~~ structure load %v ~~~ 0_degrees none true false", player, good.StructureName)
				o.Frame.GetGameControl().SendCmd(cmd)
				shouldGet := int(math.Floor(float64(good.Price) * (float64(1) - o.Tax)))
				totalTax := good.Price - shouldGet
				mapping := map[string]interface{}{
					"[购买者]":    "\"" + player + "\"",
					"[商品]":     good.Name,
					"[出售者]":    "\"" + good.Src + "\"",
					"[价格]":     good.Price,
					"[货币名]":    good.Currency.CurrencyName,
					"[货币计分板名]": good.Currency.ScoreboardName,
					"[所得税]":    totalTax,
					"[实际收益]":   shouldGet,
				}
				o.executeCmdsWithDefer(good.Src, good.UID, o.FinalCmds, mapping)
				o.Frame.GetGameControl().SayTo(player, "§6购买成功")
				o.mu.Lock()
				delete(o.fileData.Goods, goodId)
				o.mu.Unlock()
				o.fileChange = true
			} else {
				o.Frame.GetGameControl().SayTo(player, "很遗憾,你钱不够")
			}
		},
	)
}

func (o *PlayerShop) buyGood(player, goodId string) {
	if good, hasK := o.fileData.Goods[goodId]; hasK {
		o.startBuy(player, goodId, good)
	} else {
		o.Frame.GetGameControl().SayTo(player, "这个物品就在刚刚被买走了")
	}
}

func (o *PlayerShop) onBuyTrigger(chat *defines.GameChat) (stop bool) {
	stop = true
	plainGoods := make([]*PlayerShopDataGood, 0)
	for _, v := range o.fileData.Goods {
		plainGoods = append(plainGoods, v)
	}
	go o.showGoods(chat.Name, plainGoods, 0)
	return
}

func (o *PlayerShop) tryGetAmountAndCurrentInStr(in string) (amount int, currency *defines.Currency, cancel bool, err error) {
	if in == "取消" {
		return 0, nil, true, nil
	}
	re := regexp.MustCompile("^[-]?[0-9]+")
	val := re.FindAllString(in, 1)
	if len(val) == 0 {
		return 0, nil, false, fmt.Errorf("不是一个数字")
	}
	v, _ := strconv.Atoi(val[0])
	if v < 0 {
		return 0, nil, false, fmt.Errorf("不可以是负数")
	}
	for _, currency = range o.AllCurrency {
		if strings.Contains(in, currency.CurrencyName) {
			return v, currency, false, nil
		}
	}
	return v, o.DefaultCurrency, false, nil
}

func (o *PlayerShop) onSale(good *PlayerShopDataGood) {
	shouldGet := int(math.Floor(float64(good.Price) * (float64(1) - o.Tax)))
	totalTax := good.Price - shouldGet
	utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.HintOnSale, map[string]interface{}{
		"[player]": "\"" + good.Src + "\"",
		"[商品]":     good.Name,
		"[价格]":     good.Price,
		"[所得税]":    totalTax,
		"[实际收益]":   shouldGet,
		"[货币名]":    good.Currency.CurrencyName,
		"[货币计分板名]": good.Currency.ScoreboardName,
	}, o.Frame.GetBackendDisplay())
}

func (o *PlayerShop) packupGood(good *PlayerShopDataGood) {
	cmd := utils.FormatByReplacingOccurrences(o.SelectCmd, map[string]interface{}{"[player]": "\"" + good.Src + "\""})
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount == 0 {
			o.Frame.GetGameControl().SayTo(good.Src, "物品转移失败")
		} else {
			ox, oy, oz := o.PackagePlatform[0], o.PackagePlatform[1], o.PackagePlatform[2]
			sx, sy, sz := ox-1, oy, oz-1
			ex, ey, ez := ox+1, oy+1, oz+1
			sid := o.acquireEmptyNumber()
			good.StructureName = o.formatPackage(sid)
			cmd = fmt.Sprintf("structure save %v %v %v %v %v %v %v true disk false",
				good.StructureName, sx, sy, sz, ex, ey, ez)
			o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
				if output.SuccessCount != 0 {
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @e[r=3,x=%v,y=%v,z=%v] ~ -40 ~", ox, oy, oz))
					if m, err := json.Marshal(good); err == nil {
						descStr := string(m)
						o.Frame.GetBackendDisplay().Write("出售物品，信息: " + descStr)
					} else {
						o.Frame.GetBackendDisplay().Write(err.Error())
					}
					o.mu.Lock()
					o.fileData.Goods[good.StructureName] = good
					o.fileChange = true
					o.onSale(good)
					o.mu.Unlock()
				} else {
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @e[r=3,x=%v,y=%v,z=%v] %v", ox, oy, oz, good.Src))
					o.Frame.GetGameControl().SayTo(good.Src, "打包失败，尝试退回物品")
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v: 打包失败 %v (%v)", good.Src, descStr, good.StructureName))
				}

			})
		}
	})
}

func (o *PlayerShop) askForGoods(playerName, goodName string, price int, currency *defines.Currency) {
	//dstPlayer := chat.Msg[0]
	if o.Frame.GetGameControl().SetOnParamMsg(playerName, func(c *defines.GameChat) bool {

		uid := o.Frame.GetGameControl().GetPlayerKit(playerName).GetRelatedUQ().UUID.String()
		g := &PlayerShopDataGood{
			Src:           playerName,
			Name:          goodName,
			TimeStamp:     utils.TimeToString(time.Now()),
			StructureName: "",
			Price:         price,
			Currency:      *currency,
			UID:           uid,
		}
		o.packupGood(g)
		return true
	}) == nil {
		utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.HintOnCanSaleStep3, map[string]interface{}{
			"[player]": "\"" + playerName + "\"",
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *PlayerShop) askForGoodPrice(playerName, goodName string) {
	if o.Frame.GetGameControl().SetOnParamMsg(playerName, func(c *defines.GameChat) bool {
		priceStr := c.Msg[0]
		price, currency, cancel, err := o.tryGetAmountAndCurrentInStr(priceStr)
		if cancel {
			o.Frame.GetGameControl().SayTo(playerName, "已取消")
			return true
		}
		if err != nil {
			o.Frame.GetGameControl().SayTo(playerName, err.Error()+" 请重新输入，或者输入 取消")
			go o.askForGoodPrice(playerName, goodName)
			return true
		}
		o.askForGoods(playerName, goodName, price, currency)
		return true
	}) == nil {
		utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.HintOnCanSaleStep2, map[string]interface{}{
			"[player]": "\"" + playerName + "\"",
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *PlayerShop) askForGoodName(name string) {
	if o.Frame.GetGameControl().SetOnParamMsg(name, func(c *defines.GameChat) bool {
		goodName := c.Msg[0]
		o.askForGoodPrice(name, goodName)
		return true
	}) == nil {
		utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.HintOnCanSaleStep1, map[string]interface{}{
			"[player]": "\"" + name + "\"",
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *PlayerShop) onSaleTrigger(chat *defines.GameChat) (stop bool) {
	stop = true
	go func() {
		utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.PreExecuteCmdsOnSale, map[string]interface{}{
			"[player]": "\"" + chat.Name + "\"",
		}, o.Frame.GetBackendDisplay())
		if !<-utils.CheckPlayerMatchSelector(o.Frame.GetGameControl(), chat.Name, o.SaleAuthSelector) {
			utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.HintOnCannotSale, map[string]interface{}{
				"[player]": "\"" + chat.Name + "\"",
			}, o.Frame.GetBackendDisplay())
			return
		} else {
			o.askForGoodName(chat.Name)
		}
	}()
	return
}

func (o *PlayerShop) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.fileData)
		}
	}
	return nil
}

func (o *PlayerShop) Activate() {
	ox, oy, oz := o.PackagePlatform[0], o.PackagePlatform[1], o.PackagePlatform[2]
	for h := -1; h < 0; h++ {
		for wx := -1; wx < 2; wx++ {
			for wz := -1; wz < 2; wz++ {
				x, y, z := ox+wx, oy+h, oz+wz
				//fmt.Println(x, y, z)
				o.Frame.GetGameControl().
					SendCmdAndInvokeOnResponse(
						fmt.Sprintf("testforblock %v %v %v air", x, y, z),
						func(output *packet.CommandOutput) {
							//fmt.Println(output)
							if len(output.OutputMessages) > 0 && strings.Contains(output.OutputMessages[0].Message, "outOfWorld") {
								panic(pterm.Error.Sprintf("打包平台 %v 不在常加载区内！请修改打包平台位置或者设为常加载区", o.PackagePlatform))
							}
							if output.SuccessCount != 0 {
								o.Frame.GetGameControl().SendCmd(fmt.Sprintf("setblock %v %v %v sealantern", x, y, z))
							}
						},
					)
			}
		}
	}
}

func (o *PlayerShop) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.fileData)
}
