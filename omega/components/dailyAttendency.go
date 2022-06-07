package components

import (
	"encoding/json"
	"fmt"
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

type DailyAttendance struct {
	*BasicComponent
	Triggers                []string `json:"触发词"`
	Usage                   string   `json:"菜单提示"`
	CheckPointTimeOffset    string   `json:"签到时间偏移"`
	FileName                string   `json:"玩家签到信息记录文件"`
	HintOnRepeatCheckout    string   `json:"当玩家一天内重复签到时提示"`
	allPlayerAttendanceInfo map[string]*PlayerAttendanceInfo
}

func (o *DailyAttendance) computeLastCheckPointTime() time.Time {
	nowTime := time.Now()
	stratOfThisDay := nowTime.Truncate(24 * time.Hour)
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
	lastCheckPointTime := o.computeLastCheckPointTime()
	toNow := time.Now().Sub(lastCheckPointTime)
	pterm.Info.Printfln("最近一次签到开始时间: %v (-%v小时前)", lastCheckPointTime, toNow.Hours())

}

// func (o *DailyAttendance) doResponse(player string,accumulateAtten)

func (o *DailyAttendance) doUpdate(uidString string, record *PlayerAttendanceInfo) {
	checkPointTime := o.computeLastCheckPointTime()
	accumulateAttendanceDays, continuationAttendanceDays := 0, 0
	updatedTimestamp := utils.TimeToString(time.Now())
	if record.AccumulateAttendanceDays == 0 {
		accumulateAttendanceDays = 1
		continuationAttendanceDays = 1
	} else {
		if lastAttendanceTime, err := utils.StringToTime(record.LastAttendanceTime); err != nil {
			o.Frame.GetBackendDisplay().Write(pterm.Error.Sprintf("日期记录错误 %v 无法解析 %v", record, err))
		} else if lastAttendanceTime.After(checkPointTime) {
			o.Frame.GetGameControl().SayTo(record.PlayerName, o.HintOnRepeatCheckout)
			return
		} else {
			accumulateAttendanceDays++
			if lastAttendanceTime.Before(checkPointTime.Add(-(time.Duration(24) * time.Hour))) {
				continuationAttendanceDays = 1
			} else {
				continuationAttendanceDays++
			}
		}
	}
	record.AccumulateAttendanceDays = accumulateAttendanceDays
	record.ContinuationAttendanceDays = continuationAttendanceDays
	record.LastAttendanceTime = updatedTimestamp
	o.allPlayerAttendanceInfo[uidString] = record

}

func (o *DailyAttendance) onTriggerFn(chat *defines.GameChat) bool {
	playerName := chat.Name
	playerKit := o.Frame.GetGameControl().GetPlayerKit(playerName)
	if playerKit.GetRelatedUQ() == nil {
		return true
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
	return true
}

func (o *DailyAttendance) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.allPlayerAttendanceInfo = make(map[string]*PlayerAttendanceInfo)
	o.Frame.GetJsonData(o.FileName, &o.allPlayerAttendanceInfo)
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

func (o *DailyAttendance) Stop() error {
	fmt.Println("正在保存: " + o.FileName)
	return o.Frame.WriteJsonData(o.FileName, o.allPlayerAttendanceInfo)
}

func (o *DailyAttendance) Activate() {

}
