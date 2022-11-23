package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"
	"strconv"
	"strings"
)

type MoneyTransfer struct {
	*defines.BasicComponent
	Triggers        []string            `json:"触发词"`
	DefaultCurrency *defines.Currency   `json:"默认货币"`
	AllCurrency     []*defines.Currency `json:"可转账货币"`
	Usage           string              `json:"提示信息"`
}

func (o *MoneyTransfer) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	if o.DefaultCurrency.CurrencyName == "" {
		panic("必须设置默认货币名")
	}
	currencys := []*defines.Currency{o.DefaultCurrency}
	allNames := map[string]bool{}
	allNames[o.DefaultCurrency.CurrencyName] = true
	for _, c := range o.AllCurrency {
		if _, ok := allNames[c.CurrencyName]; ok {
			continue
		}
		currencys = append(currencys, c)
	}
	o.AllCurrency = currencys
}

func (o *MoneyTransfer) getCurrencyName() string {
	allNames := map[string]bool{}
	nameStrs := []string{o.DefaultCurrency.CurrencyName + "(默认)"}
	allNames[o.DefaultCurrency.CurrencyName] = true
	for _, c := range o.AllCurrency {
		if _, ok := allNames[c.CurrencyName]; ok {
			continue
		}
		nameStrs = append(nameStrs, c.CurrencyName)
	}
	return strings.Join(nameStrs, "/")
}

func (o *MoneyTransfer) tryGetAmountAndCurrentInStr(in string) (amount int, currency *defines.Currency, err error) {
	re := regexp.MustCompile("^[-]?[0-9]+")
	val := re.FindAllString(in, 1)
	if len(val) == 0 {
		return 0, nil, fmt.Errorf("不是一个数字")
	}
	v, _ := strconv.Atoi(val[0])
	if v < 0 {
		return 0, nil, fmt.Errorf("不可以是负数")
	}
	for _, currency = range o.AllCurrency {
		if strings.Contains(in, currency.CurrencyName) {
			return v, currency, nil
		}
	}
	return v, o.DefaultCurrency, nil
}

func (o *MoneyTransfer) doTransfer(src, dst string, c *defines.Currency, amount int) {
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(
		fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v 0", src, c.ScoreboardName), func(output *packet.CommandOutput) {
			if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 4 {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("购买时发现玩家没有记分板%v %v", c.ScoreboardName, output))
				o.Frame.GetGameControl().SayTo(src, "似乎没有相关记分板，或者你这个记分板没有分数")
				return
			}
			hasMoney, err := strconv.Atoi(output.OutputMessages[0].Parameters[3])
			if err != nil {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("转账时解析出错 %v", err))
				o.Frame.GetGameControl().SayTo(src, "出于未知原因，无法完成购买")
				return
			}
			if hasMoney > amount {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("移除玩家 %v 数额 %v / %v (%v)", src, amount, hasMoney, c.ScoreboardName))
				o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", src, c.ScoreboardName, amount),
					func(output *packet.CommandOutput) {
						if output.SuccessCount > 0 {
							o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", dst, c.ScoreboardName, amount),
								func(output *packet.CommandOutput) {
									rec, success := "", false
									if output.SuccessCount > 0 {
										rec = fmt.Sprintf("%v -> %v: 转账了 %v %v", src, dst, amount, c.CurrencyName)
										success = true
									} else {
										rec = fmt.Sprintf("%v -> %v: 转账失败了(将尝试退回)", src, dst)
										success = false
									}
									o.Frame.GetBackendDisplay().Write(rec)
									if !success {
										o.Frame.GetGameControl().SayTo(src, "转账失败，尝试退回")
										o.Frame.GetGameControl().SayTo(dst, "转账失败，尝试退回")
										o.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", src, c.ScoreboardName, amount))
									} else {
										o.Frame.GetGameControl().SayTo(src, "转账完成: 给予"+dst)
										o.Frame.GetGameControl().SayTo(dst, "转账完成: 来自"+src)
										o.Frame.GetGameControl().SayTo(src, fmt.Sprintf("%v %v", amount, c.CurrencyName))
										o.Frame.GetGameControl().SayTo(dst, fmt.Sprintf("%v %v", amount, c.CurrencyName))
									}
								})
						} else {
							o.Frame.GetGameControl().SayTo(src, "很遗憾,转账失败了")
							o.Frame.GetBackendDisplay().Write(fmt.Sprintf("转账时扣费出现意料之外的错误 %v", output))
						}
					})
			} else {
				o.Frame.GetGameControl().SayTo(src, "很遗憾,你钱不够")
			}
		},
	)
}

func (o *MoneyTransfer) tryTransfer(src, dst string, args []string) {
	hint := ""
	if len(args) == 1 {
		if strings.Contains(args[0], "取消") {
			return
		}
		amount, c, err := o.tryGetAmountAndCurrentInStr(args[0])
		if err == nil {
			o.doTransfer(src, dst, c, amount)
			return
		}
		hint = "不行," + err.Error() + ","
	} else if len(args) > 1 {
		in := strings.Join(args, " ")
		amount, c, err := o.tryGetAmountAndCurrentInStr(in)
		if err == nil {
			o.doTransfer(src, dst, c, amount)
			return
		}
		hint = "不行," + err.Error() + ","
	}
	if o.Frame.GetGameControl().SetOnParamMsg(src,
		func(chat *defines.GameChat) (catch bool) {
			o.tryTransfer(src, dst, chat.Msg)
			return true
		}) == nil {
		o.Frame.GetGameControl().SayTo(src, hint+"你要转多少? 或者，输入 取消 :")
	}
}

func (o *MoneyTransfer) check(chat *defines.GameChat) bool {
	noTarget := len(chat.Msg) == 0
	var availablePlayers []string
	flag := false
	for _, p := range o.Frame.GetUQHolder().PlayersByEntityID {
		availablePlayers = append(availablePlayers, p.Username)
		if !noTarget && chat.Msg[0] == p.Username {
			flag = true
		}
	}
	if flag {
		o.tryTransfer(chat.Name, chat.Msg[0], chat.Msg[1:])
		return true
	}
	if noTarget {
		o.Frame.GetGameControl().SayTo(chat.Name, "转账给谁?")
	} else {
		o.Frame.GetGameControl().SayTo(chat.Name, "没有你说的那个玩家...")
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(availablePlayers)
	if o.Frame.GetGameControl().SetOnParamMsg(chat.Name,
		func(chat *defines.GameChat) (catch bool) {
			i, cancel, err := resolver(chat.Msg)
			if cancel {
				o.Frame.GetGameControl().SayTo(chat.Name, "已取消")
				return true
			}
			if err != nil {
				o.Frame.GetGameControl().SayTo(chat.Name, "无法传送，因为输入"+err.Error())
				return true
			}
			o.tryTransfer(chat.Name, availablePlayers[i], chat.Msg[1:])
			return true
		}) == nil {
		o.Frame.GetGameControl().SayTo(chat.Name, "可选项有: "+hint+" 请输入喔:")
	}
	return true
}

func (o *MoneyTransfer) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if o.Usage == "" {
		o.Usage = "给目标玩家转账，可以被转账的货币包括"
	}
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[玩家名] [数量]",
			FinalTrigger: false,
			Usage:        o.Usage + o.getCurrencyName(),
		},
		OptionalOnTriggerFn: o.check,
	})
}
