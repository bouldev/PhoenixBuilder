package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strconv"
	"time"
)

type Option struct {
	Name             string      `json:"选项名"`
	Description      string      `json:"附加说明"`
	CurrencyName     string      `json:"货币名"`
	Price            int         `json:"价格"`
	MaxRecyclePerDay int         `json:"每天最多回收"`
	ClearCmd         string      `json:"清除模版"`
	RewardCmdsIn     interface{} `json:"收益指令"`
	RewardCmds       []defines.Cmd
}

type LimitRecord struct {
	StrTime    string `json:"时间"`
	cached     bool
	time       time.Time
	CountsLeft int `json:"剩余次数"`
}

type Recycle struct {
	*BasicComponent
	RecordFileName      string   `json:"最后回收记录文件"`
	Triggers            []string `json:"触发词"`
	Format              string   `json:"展示模版"`
	Options             []Option `json:"回收清单文件"`
	PlayerRecycleRecord map[string]map[string]LimitRecord
}

func (o *Recycle) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	var err error
	for _, o := range o.Options {
		if o.RewardCmds, err = utils.ParseAdaptiveCmd(o.RewardCmdsIn); err != nil {
			panic(err)
		}
	}
}

// func (o *Recycle) putDefaultFile() {
// 	if !utils.IsFile(o.Frame.GetRelativeFileName(o.FileName)) {
// 		pterm.Warning.Printf("没有检测到回收清单文件,将在 %v 下展开默认回收清单\n", o.Frame.GetRelativeFileName(o.FileName))
// 		o.Frame.WriteFileData(o.FileName, defaultRecycleFile)
// 	}
// }

func (o *Recycle) getCountsLeft(name string, option string) (time.Time, int) {
	if _, ok := o.PlayerRecycleRecord[name]; !ok {
		// o.PlayerRecycleRecord[name] = make(map[string]LimitRecord)
		return time.Now().Add(-time.Hour * time.Duration(24*2)), 0
	}
	if _, ok := o.PlayerRecycleRecord[name][option]; !ok {
		return time.Now().Add(-time.Hour * time.Duration(24*2)), 0
	}
	if o.PlayerRecycleRecord[name][option].cached {
		return o.PlayerRecycleRecord[name][option].time, o.PlayerRecycleRecord[name][option].CountsLeft
	}
	t, err := time.Parse("2006-01-02 15:04:05", o.PlayerRecycleRecord[name][option].StrTime)
	if err != nil {
		o.PlayerRecycleRecord[name][option] = LimitRecord{
			StrTime:    time.Now().Add(-time.Hour * time.Duration(24*2)).Format("2006-01-02 15:04:05"),
			CountsLeft: o.PlayerRecycleRecord[name][option].CountsLeft,
		}
		return time.Now().Add(-time.Hour * time.Duration(24*2)), o.PlayerRecycleRecord[name][option].CountsLeft
	}
	return t, o.PlayerRecycleRecord[name][option].CountsLeft
}

func (o *Recycle) computeMaxRecycleCount(name string, option string, maxCount int) int {
	if maxCount == 0 {
		return 999
	}
	t, count := o.getCountsLeft(name, option)
	if count < 0 {
		count = 0
	}
	factor := time.Now().Sub(t).Seconds() / float64(time.Hour.Seconds()*24)
	recCount := float64(maxCount) * factor
	allowCount := int(recCount) + count
	if allowCount > maxCount {
		allowCount = maxCount
	}
	if allowCount < 0 {
		return 0
	}
	return allowCount
}

func (o *Recycle) setCountsLeft(name string, option string, counts int) {
	if _, ok := o.PlayerRecycleRecord[name]; !ok {
		o.PlayerRecycleRecord[name] = make(map[string]LimitRecord)
	}
	o.PlayerRecycleRecord[name][option] = LimitRecord{
		StrTime:    time.Now().Format("2006-01-02 15:04:05"),
		time:       time.Now(),
		cached:     true,
		CountsLeft: counts,
	}
}

func (o *Recycle) Stop() error {
	fmt.Println("正在保存" + o.RecordFileName)
	return o.Frame.WriteJsonData(o.RecordFileName, o.PlayerRecycleRecord)
}

func (o *Recycle) popMenu(name string) {
	pk := o.Frame.GetGameControl().GetPlayerKit(name)
	availableOptions := []string{}
	for i, e := range o.Options {
		leftStr := "无限制"
		if e.MaxRecyclePerDay != 0 {
			leftStr = fmt.Sprintf("%v", o.computeMaxRecycleCount(name, e.Name, e.MaxRecyclePerDay))
		}
		I := i + 1
		m := utils.FormatByReplacingOccurrences(o.Format, map[string]interface{}{
			"[i]":             I,
			"[name]":          "\"" + e.Name + "\"",
			"[description]":   e.Description,
			"[price]":         e.Price,
			"[currency_name]": e.CurrencyName,
			"[maxRecycle]":    leftStr,
		})
		availableOptions = append(availableOptions, e.Name)
		pk.Say(m)
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(availableOptions)
	if pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		i, cancel, err := resolver(chat.Msg)
		if cancel {
			pk.Say("已取消")
			return true
		}
		if err != nil {
			pk.Say(fmt.Sprintf("无法理解你的选择，因为 %v", err))
			return true
		}
		option := o.Options[i]
		if len(chat.Msg) == 1 {
			o.askForAmount(chat.Name, option)
		} else {
			o.tryHandleAmount(chat.Name, option, chat.Msg[1])
		}
		//chat.Msg = utils.InsertHead(o.Triggers[0], chat.Msg)
		//o.Frame.GetGameListener().Throw(chat)
		return true
	}) == nil {
		pk.Say(fmt.Sprintf("你想回收什么呢？ 请输入 %v, 以及你想回收的 [数量] 喔!", hint))
	}
}

func (o *Recycle) tryHandleAmount(name string, option Option, amount string) {
	maxC := 0
	if option.MaxRecyclePerDay != 0 {
		maxC = o.computeMaxRecycleCount(name, option.Name, option.MaxRecyclePerDay)
	} else {
		maxC = 999
	}
	if maxC == 0 {
		o.Frame.GetGameControl().GetPlayerKit(name).Say("已经不能再回收了，明天再来吧")
		return
	}
	set, err := strconv.Atoi(amount)
	if err != nil {
		o.Frame.GetGameControl().GetPlayerKit(name).Say("输入的数量[" + amount + "] 不是一个正整数")
		return
	}
	if set < 1 {
		o.Frame.GetGameControl().GetPlayerKit(name).Say("输入的数量[" + amount + "] 必须大于0")
		return
	}
	if set > maxC {
		o.Frame.GetGameControl().GetPlayerKit(name).Say("你最多只能回收" + strconv.Itoa(maxC) + "个")
		set = maxC
	}
	o.startRecycle(name, option, set)
}

func (o *Recycle) askForAmount(name string, option Option) {
	maxC := 0
	if option.MaxRecyclePerDay != 0 {
		maxC = o.computeMaxRecycleCount(name, option.Name, option.MaxRecyclePerDay)
		if maxC == 0 {
			o.Frame.GetGameControl().GetPlayerKit(name).Say("已经不能再回收了，明天再来吧")
			return
		}
	} else {
		maxC = 999
	}
	hint, resolver := utils.GenIntRangeResolver(1, maxC)
	if o.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		amount, cancel, err := resolver(chat.Msg)
		if cancel {
			o.Frame.GetGameControl().GetPlayerKit(name).Say("已取消")
			return true
		}
		if err != nil {
			o.Frame.GetGameControl().GetPlayerKit(name).Say("输入的数量无效，因为 " + err.Error())
			return true
		}
		o.startRecycle(name, option, amount)
		return true
	}) == nil {
		o.Frame.GetGameControl().GetPlayerKit(name).Say("你想回收多少个" + option.Name + "呢？ 请输入 " + hint)
	}
}

func (o *Recycle) startRecycle(name string, option Option, amount int) {
	// totalPrice := amount * option.Price
	cmd := utils.FormatByReplacingOccurrences(option.ClearCmd, map[string]interface{}{
		"[player]": "\"" + name + "\"",
		"[count]":  amount,
	})
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		//fmt.Println(output)
		if output.SuccessCount == 0 {
			o.Frame.GetGameControl().SayTo(name, "没有找到相关物品")
			return
		}
		realAmount := output.OutputMessages[0].Parameters[1]
		atoi, err := strconv.Atoi(realAmount)
		if err != nil {
			o.Frame.GetGameControl().SayTo(name, "发生了未知错误"+err.Error())
			return
		}
		//fmt.Println(atoi)
		o.onRecycleSuccess(name, option, atoi)
	})
}

func (o *Recycle) onRecycleSuccess(name string, option Option, realCount int) {
	//o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 回收 %v * %v 收益 %v (%v->%v)"))
	leftStr := "无限制"
	if option.MaxRecyclePerDay != 0 {
		maxC := o.computeMaxRecycleCount(name, option.Name, option.MaxRecyclePerDay)
		Cleft := maxC - realCount
		if Cleft < 0 {
			Cleft = 0
		}
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 回收 %v * %v 收益 %v (%v->%v)", name, realCount, option.Name, realCount*option.Price, maxC, Cleft))
		o.setCountsLeft(name, option.Name, Cleft)
		leftStr = fmt.Sprintf("%v", Cleft)
	} else {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 回收 %v * %v 收益 %v", name, realCount, option.Name, realCount*option.Price))
	}

	o.Frame.GetGameControl().SayTo(name, "回收成功")
	mapping := map[string]interface{}{
		"[player]":     "\"" + name + "\"",
		"[realCount]":  realCount,
		"[totalPrice]": realCount * option.Price,
		"[leftCount]":  leftStr,
		"[price]":      option.Price,
	}
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), option.RewardCmds, mapping, o.Frame.GetBackendDisplay())
}

func (o *Recycle) dispatch(chat *defines.GameChat) bool {
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	if len(chat.Msg) == 0 {
		o.popMenu(chat.Name)
		return true
	}
	item := chat.Msg[0]
	for _, e := range o.Options {
		if e.Name == item {
			if len(chat.Msg) > 1 {
				o.tryHandleAmount(chat.Name, e, chat.Msg[1])
			} else {
				o.askForAmount(chat.Name, e)
			}
			return true
		}
	}
	pk.Say(fmt.Sprintf("%v 不是有效的回收选项", item))
	o.popMenu(chat.Name)
	return true
}

func (o *Recycle) Inject(frame defines.MainFrame) {
	o.Frame = frame
	// o.putDefaultFile()
	// o.Options = []Option{}
	// err := o.Frame.GetJsonData(o.FileName, &o.Options)
	// if err != nil {
	// 	panic(err)
	// }
	o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 回收项目已加载", len(o.Options)))
	o.PlayerRecycleRecord = map[string]map[string]LimitRecord{}
	o.Frame.GetJsonData(o.RecordFileName, &o.PlayerRecycleRecord)
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[名称] [数量]",
			Usage:        "回收物品",
			FinalTrigger: false,
		},
		OptionalOnTriggerFn: o.dispatch,
	})
}
