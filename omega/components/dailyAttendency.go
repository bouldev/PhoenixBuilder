package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"

	"github.com/pterm/pterm"
)

type PlayerAttendanceInfo struct {
	LastAttendanceTime         string `json:"最后签到日期"`
	AccumulateAttendanceDays   int    `json:"累计签到天数"`
	ContinuationAttendanceDays int    `json:"连续签到天数"`
	PlayerName                 string `json:"玩家名"`
}

type RewardPolicy struct {
	OnAccumulateDay     int         `json:"当累计签到天数为_时"`
	OnAccumulateDayGE   int         `json:"当累计签到天数大于等于_时"`
	OnAccumulateDayLE   int         `json:"当累计签到天数小于等于_时"`
	OnContinuationDay   int         `json:"当连续签到天数为_时"`
	OnContinuationDayGE int         `json:"当连续签到天数大于等于_时"`
	OnContinuationDayLE int         `json:"当连续签到天数小于等于_时"`
	OnWeek              int         `json:"当今天是周_时"`
	Describe            string      `json:"规则描述"`
	RewardCmdsIn        interface{} `json:"当符合上述所有不为零的规则时执的奖励指令"`
	rewardCmds          []defines.Cmd
}

func (rp *RewardPolicy) isMatch(accumulateDays, continuationDays int) (match bool) {
	match = false
	week := int(time.Now().Weekday())
	if week == 0 {
		// sunday =0 by default -> sunday =7
		week = 7
	}
	// fmt.Println(week)
	if rp.OnWeek != 0 && rp.OnWeek != int(week) {
		return
	}
	if rp.OnAccumulateDay != 0 && rp.OnAccumulateDay != accumulateDays {
		return
	}
	if rp.OnAccumulateDayGE != 0 && rp.OnAccumulateDayGE > accumulateDays {
		return
	}
	if rp.OnAccumulateDayLE != 0 && rp.OnAccumulateDayLE < accumulateDays {
		return
	}
	if rp.OnContinuationDay != 0 && rp.OnContinuationDay != continuationDays {
		return
	}
	if rp.OnContinuationDayGE != 0 && rp.OnContinuationDayGE > continuationDays {
		return
	}
	if rp.OnContinuationDayLE != 0 && rp.OnContinuationDayLE < continuationDays {
		return
	}
	return true
}

type DailyAttendance struct {
	*BasicComponent
	Triggers                []string        `json:"触发词"`
	Usage                   string          `json:"菜单提示"`
	CheckPointTimeOffset    string          `json:"签到时间偏移"`
	FileName                string          `json:"玩家签到信息记录文件"`
	HintOnRepeatCheckout    string          `json:"当玩家一天内重复签到时提示"`
	StopOnFirstMatch        bool            `json:"为true只匹配第一个符合的签到规则false则使用所有匹配的签到规则"`
	RewardPolicies          []*RewardPolicy `json:"签到规则"`
	PlayerLoginDelay        int             `json:"玩家登录延迟"`
	HintOnSuggestCheckOut   string          `json:"提醒玩家签到的消息"`
	PassiveCheckOut         bool            `json:"为true时被动签到false时主动签到"`
	allPlayerAttendanceInfo map[string]*PlayerAttendanceInfo
}

func (o *DailyAttendance) computeLastCheckPointTime() time.Time {
	nowTime := time.Now()
	stratOfThisDay := nowTime.Truncate(24 * time.Hour).Add(-time.Duration(8) * time.Hour)
	if timeOffset, err := time.ParseDuration(o.CheckPointTimeOffset); err != nil {
		panic(fmt.Errorf("时间偏移无效: %v", err))
	} else {
		baselineTime := stratOfThisDay.Add(timeOffset)
		if baselineTime.After(nowTime) {
			if baselineTime.Add(-(time.Duration(24) * time.Hour)).After(nowTime) {
				panic("签到时间偏移应该控制在 24 小时内")
			}
		} else {
			if baselineTime.Add(time.Duration(24) * time.Hour).Before(nowTime) {
				panic("签到时间偏移应该控制在 24 小时内")
			}
		}
		return baselineTime
	}
}

func (o *DailyAttendance) Init(cfg *defines.ComponentConfig) {
	if m, err := json.Marshal(cfg.Configs); err != nil {
		panic(err)
	} else {
		if err := json.Unmarshal(m, o); err != nil {
			panic(err)
		}
	}
	for _, policy := range o.RewardPolicies {
		var err error
		if policy.rewardCmds, err = utils.ParseAdaptiveCmd(policy.RewardCmdsIn); err != nil {
			panic(fmt.Errorf("无法理解的签到奖励指令序列 %v (%v)", policy.RewardCmdsIn, err))
		}
	}
	lastCheckPointTime := o.computeLastCheckPointTime()
	toNow := time.Now().Sub(lastCheckPointTime)
	pterm.Info.Printfln("最近一次签到开始时间: %v (-%.1f小时前)", lastCheckPointTime, toNow.Hours())

}

func (o *DailyAttendance) doSendReward(player string, accumulateDays, continuationDays int, policy *RewardPolicy) {
	remapping := map[string]interface{}{
		"[player]": "\"" + player + "\"",
		"[连续签到]":   continuationDays,
		"[累计签到]":   accumulateDays,
	}
	utils.LaunchCmdsArray(o.Frame.GetGameControl(), policy.rewardCmds, remapping, o.Frame.GetBackendDisplay())
}

func (o *DailyAttendance) doResponse(player string, accumulateDays, continuationDays int) {
	hint := fmt.Sprintf("玩家 %v 累计签到 %v 连续签到 %v", player, accumulateDays, continuationDays)
	matchFirst := o.StopOnFirstMatch
	matched := false
	for _, policy := range o.RewardPolicies {
		if match := policy.isMatch(accumulateDays, continuationDays); match {
			matched = true
			hint = hint + fmt.Sprintf(" [匹配奖励]: %v", policy.Describe)
			o.doSendReward(player, accumulateDays, continuationDays, policy)
			if matchFirst {
				break
			}
		}
	}
	if !matched {
		hint = hint + fmt.Sprintf(" 没有匹配到任何奖励项")
	}
	o.Frame.GetBackendDisplay().Write(hint)
}

func (o *DailyAttendance) doUpdate(uidString string, record *PlayerAttendanceInfo) {
	checkPointTime := o.computeLastCheckPointTime()
	// accumulateAttendanceDays, continuationAttendanceDays := 0, 0
	updatedTimestamp := utils.TimeToString(time.Now())
	if record.AccumulateAttendanceDays == 0 {
		record.AccumulateAttendanceDays = 1
		record.ContinuationAttendanceDays = 1
	} else {
		// lastAttendanceTime, _ := utils.StringToTimeWithLocal(record.LastAttendanceTime + " +0800 CST")
		// fmt.Println(lastAttendanceTime.Local())
		// fmt.Println(checkPointTime)
		if lastAttendanceTime, err := utils.StringToTime(record.LastAttendanceTime + " +0800 CST"); err != nil {
			o.Frame.GetBackendDisplay().Write(pterm.Error.Sprintf("日期记录错误 %v 无法解析 %v", record, err))
		} else if lastAttendanceTime.After(checkPointTime) {
			o.Frame.GetGameControl().SayTo(record.PlayerName, o.HintOnRepeatCheckout)
			return
		} else {
			record.AccumulateAttendanceDays++
			if lastAttendanceTime.Before(checkPointTime.Add(-(time.Duration(24) * time.Hour))) {
				record.ContinuationAttendanceDays = 1
			} else {
				record.ContinuationAttendanceDays++
			}
		}
	}
	record.LastAttendanceTime = updatedTimestamp
	// fmt.Println(record)
	o.allPlayerAttendanceInfo[uidString] = record
	o.doResponse(record.PlayerName, record.AccumulateAttendanceDays, record.ContinuationAttendanceDays)
}

func (o *DailyAttendance) onTriggerFn(chat *defines.GameChat) bool {
	playerName := chat.Name
	o.checkOut(playerName)
	return true
}

func (o *DailyAttendance) checkOut(playerName string) {
	playerKit := o.Frame.GetGameControl().GetPlayerKit(playerName)
	if playerKit.GetRelatedUQ() == nil {
		return
	}
	uid := playerKit.GetRelatedUQ().UUID
	uidStr := uid.String()
	var record *PlayerAttendanceInfo
	if _record, hasK := o.allPlayerAttendanceInfo[uidStr]; hasK {
		_record.PlayerName = playerName
		record = _record
	} else {
		record = &PlayerAttendanceInfo{
			PlayerName:               playerName,
			AccumulateAttendanceDays: 0,
		}
	}
	o.doUpdate(uidStr, record)
	return
}

func (o *DailyAttendance) isPlayerChecked(player *uqHolder.Player) bool {
	uidStr := player.UUID.String()
	if record, hasK := o.allPlayerAttendanceInfo[uidStr]; hasK {
		// 玩家有签到记录
		// 计算最近一次签到开始时间
		checkPointTime := o.computeLastCheckPointTime()
		// 换算玩家最近签到时间
		if lastAttendanceTime, err := utils.StringToTime(record.LastAttendanceTime + " +0800 CST"); err != nil {
			// 记录文件损坏，在后台显示警告
			o.Frame.GetBackendDisplay().Write(pterm.Error.Sprintf("日期记录错误 %v 无法解析 %v", record, err))
			return true
		} else if lastAttendanceTime.After(checkPointTime) {
			// 玩家最近签到时间比最近一次签到时间还晚，玩家已经签到了
			return true
		}
		// 玩家最近签到时间比最近一次签到时间还晚，玩家还没有签到
		return false
	} else {
		// 玩家没有签到记录
		return false
	}
}

func (o *DailyAttendance) onPlayerLogin(entry protocol.PlayerListEntry) {
	go func() {
		t := time.NewTimer(time.Duration(o.PlayerLoginDelay) * time.Second)
		<-t.C
		// 此时检测玩家是否还在线
		uid := entry.EntityUniqueID
		if player, hasK := o.Frame.GetUQHolder().PlayersByEntityID[uid]; hasK {
			if !o.isPlayerChecked(player) {
				// 假如这里已经判断玩家还没有签到,我们要向玩家发送指定提示
				text := utils.FormatByReplacingOccurrences(o.HintOnSuggestCheckOut, map[string]interface{}{
					"[player]": entry.Username,
				})
				o.Frame.GetGameControl().SayTo(entry.Username, text)
				if o.PassiveCheckOut {
					// 被动签到
					o.checkOut(entry.Username)
				}
			}
		}
	}()
}

func (o *DailyAttendance) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.allPlayerAttendanceInfo = make(map[string]*PlayerAttendanceInfo)
	o.Frame.GetJsonData(o.FileName, &o.allPlayerAttendanceInfo)
	if !o.PassiveCheckOut {
		// 以下为菜单项渲染
		o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
			MenuEntry: defines.MenuEntry{
				Triggers:     o.Triggers,
				ArgumentHint: "",
				Usage:        o.Usage,
				FinalTrigger: false,
			},
			OptionalOnTriggerFn: o.onTriggerFn,
		})
	}
	// 一个组件的生命周期包括 初始化-> 注入主框架 -> 激活 -> 停止四个阶段
	if o.HintOnSuggestCheckOut != "" || o.PassiveCheckOut {
		o.Frame.GetGameListener().AppendLoginInfoCallback(o.onPlayerLogin)
	}
}

func (o *DailyAttendance) Stop() error {
	fmt.Println("正在保存: " + o.FileName)
	return o.Frame.WriteJsonData(o.FileName, o.allPlayerAttendanceInfo)
}
