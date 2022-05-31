package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strconv"
)

type Good struct {
	Name  string   `json:"商品显示名"`
	Once  bool     `json:"一次只能买一件"`
	Price int      `json:"价格"`
	Cmds  []string `json:"给予模版"`
}

type GoodsGroup struct {
	CurrencyName string `json:"货币显示名"`
	CurrencyCmd  string `json:"货币记分板指令名"`
	Goods        []Good `json:"商品"`
}

type PlainGood struct {
	Good
	CurrencyName string
	CurrencyCmd  string
}

type Shop struct {
	*BasicComponent
	Goods      map[string]GoodsGroup `json:"商品清单文件"`
	Triggers   []string              `json:"触发词"`
	Format     string                `json:"展示模版"`
	FormatOnce string                `json:"一次只能购买一个时的展示模版"`
	PlainItems map[string]PlainGood
}

func (o *Shop) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *Shop) askForItemList(chat *defines.GameChat) {
	groupNames := []string{}
	for k, _ := range o.Goods {
		groupNames = append(groupNames, k)
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(groupNames)

	if o.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(newChat *defines.GameChat) (catch bool) {
		i, cancel, err := resolver(newChat.Msg)
		if cancel {
			o.Frame.GetGameControl().GetPlayerKit(chat.Name).Say("已取消")
			return true
		}
		if err != nil {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("无法处理你的要求，因为"+err.Error()))
			return true
		}
		groupName := groupNames[i]
		cn := o.Goods[groupName].CurrencyName
		availableGoods := []string{}
		for i, g := range o.Goods[groupName].Goods {
			_i := i + 1
			f := o.Format
			if g.Once {
				f = o.FormatOnce
			}
			availableGoods = append(availableGoods, g.Name)
			cmd := utils.FormatByReplacingOccurrences(f, map[string]interface{}{
				"[i]":             _i,
				"[price]":         g.Price,
				"[currency_name]": cn,
				"[name]":          "\"" + g.Name + "\"",
			})
			o.Frame.GetGameControl().SayTo(chat.Name, cmd)
		}
		itemHint, itemResolver := utils.GenStringListHintResolverWithIndex(availableGoods)
		if o.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(itemChat *defines.GameChat) (catch bool) {
			itemI, cancel, err := itemResolver(itemChat.Msg)
			if cancel {
				o.Frame.GetGameControl().SayTo(chat.Name, "已取消")
				return true
			}
			if err != nil {
				o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("无法处理你的要求，因为"+err.Error()))
				return true
			}
			goodName := availableGoods[itemI]
			itemChat.Msg[0] = goodName
			o.tryBuy(itemChat)
			return true
		}) == nil {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("想购买的话，请输入 %v %v [物品名或序号] [数量] 喔！\n物品信息可选有"+itemHint,
				o.Frame.GetGameListener().GetTriggerWord(), o.Triggers[0]))
		}

		return true
	}) == nil {
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§6§l这是已有的商品分类，你想确认一下哪一类呢?\n"+hint+", 请输入喔:"))
	}
}

func (o *Shop) startBuy(player string, count int, good PlainGood) {
	fmt.Println(player, count, good)
	if good.Once {
		count = 1
	}
	totalPrice := count * good.Price
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(
		fmt.Sprintf("scoreboard players add \"%v\" %v 0", player, good.CurrencyCmd), func(output *packet.CommandOutput) {
			//fmt.Println(output)
			if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 4 {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("购买时发现玩家没有记分板%v %v", good.CurrencyName, output))
				o.Frame.GetGameControl().SayTo(player, "似乎没有相关记分板，或者你这个记分板没有分数")
				return
			}
			hasMoney, err := strconv.Atoi(output.OutputMessages[0].Parameters[3])
			if err != nil {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("购买时解析出错 %v", err))
				o.Frame.GetGameControl().SayTo(player, "出于未知原因，无法完成购买")
				return
			}
			if hasMoney > totalPrice {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("玩家 %v 花费 %v / %v 购买了 %v * %v", player, totalPrice, hasMoney, good.Name, count))
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreboard players remove \"%v\" %v %v", player, good.CurrencyCmd, totalPrice))
				for _, t := range good.Cmds {
					c := utils.FormatByReplacingOccurrences(t, map[string]interface{}{
						"[player]":      "\"" + player + "\"",
						"[totalPrice]":  totalPrice,
						"[moneyHas]":    hasMoney,
						"[moneyLeft]":   hasMoney - totalPrice,
						"[count]":       count,
						"[currency]":    good.CurrencyName,
						"[currencyCMD]": good.CurrencyCmd,
					})
					o.Frame.GetGameControl().SendCmd(c)
				}
			} else {
				o.Frame.GetGameControl().SayTo(player, "很遗憾,你钱不够")
			}
		},
	)
}

func (o *Shop) tryBuy(chat *defines.GameChat) {
	item := chat.Msg[0]
	if len(chat.Msg) == 1 {
		chat.Msg = append(chat.Msg, "1")
	}
	count := chat.Msg[1]
	good, hasK := o.PlainItems[item]
	if !hasK {
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§4§l似乎没有这个商品"))
		o.askForItemList(chat)
		return
	}
	atoi, err := strconv.Atoi(count)
	if err != nil || atoi <= 0 {
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§4§l输入的数量无效，必须是一个正整数"))
		return
	}
	o.startBuy(chat.Name, atoi, good)
}

func (o *Shop) guide(chat *defines.GameChat) bool {
	msg := chat.Msg
	if len(msg) == 0 {
		o.askForItemList(chat)
		return true
	}
	if len(msg) == 1 {
		chat.Msg = append(msg, "1")
		o.tryBuy(chat)
	} else {
		o.tryBuy(chat)
	}
	return true
}

func (o *Shop) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[物品] [数量]",
			FinalTrigger: false,
			Usage:        "购买东西，留空查看清单",
		},
		OptionalOnTriggerFn: o.guide,
	})
	o.PlainItems = map[string]PlainGood{}
	for _, gs := range o.Goods {
		c := gs.CurrencyCmd
		cn := gs.CurrencyName
		for _, e := range gs.Goods {
			o.PlainItems[e.Name] = PlainGood{
				Good:         e,
				CurrencyName: cn,
				CurrencyCmd:  c,
			}
		}
	}
	o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 商品已加载", len(o.PlainItems)))
}

func (o *Shop) Activate() {

}
