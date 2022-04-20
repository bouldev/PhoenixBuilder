package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type GameTimeSyncConfig struct {
	Enable        bool `json:"启用现实时间同步"`
	SyncDuration  int  `json:"同步指令发送周期"`
	MoonSync      bool `json:"月相同步"`
	DontTouchMoon bool `json:"不调整月相"`
	LockMoon      int  `json:"锁定月相为"`
}

type ScoreboardCalibrateConfig struct {
	Enable       bool     `json:"启用记分板校准"`
	SyncDuration int      `json:"校准指令发送周期"`
	Cmds         []string `json:"校准指令"`
}

type TimeSync struct {
	*BasicComponent
	TimeOffset                int `json:"时区偏移(秒)"`
	GameTimeSyncConfig        `json:"现实时间同步"`
	ScoreboardCalibrateConfig `json:"记分板校准"`
}

func (o *TimeSync) Init(cfg *defines.ComponentConfig) {
	marshal, err := json.Marshal(cfg.Configs)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(marshal, o)
	if err != nil {
		panic(err)
	}
	if o.GameTimeSyncConfig.Enable {
		if o.MoonSync && o.DontTouchMoon {
			panic(fmt.Sprintf("月相同步/不调整月相 只能启用其一"))
		}
		if o.LockMoon != 0 && (o.DontTouchMoon || o.MoonSync) {
			panic(fmt.Sprintf("锁定月相 不为0时，月相同步/不调整月相 都不可被启用"))
		}
	}
}

func (o *TimeSync) getRealTimeInMCTick() int {
	nowTime := time.Now()
	nowTime = nowTime.Add(time.Duration(o.TimeOffset) * time.Second)
	absSecond := (((nowTime.Hour()-6)*60)+nowTime.Minute())*60 + nowTime.Second()
	if o.GameTimeSyncConfig.MoonSync {
		absSecond = (nowTime.Day()*8/30)*(24*60*60) + absSecond
	}
	mcTick := (absSecond * 24000) / (24 * 60 * 60)
	if o.GameTimeSyncConfig.DontTouchMoon {
		basicTick := (o.Frame.GetUQHolder().Time / 24000) * 24000
		mcTick += int(basicTick)
	}
	if o.LockMoon != 0 {
		mcTick += (o.LockMoon - 1) * 24000
	}
	return mcTick
}

func (o *TimeSync) calibrateTime() {
	nowTime := time.Now()
	nowTime = nowTime.Add(time.Duration(o.TimeOffset) * time.Second)
	year := nowTime.Year()
	month := nowTime.Month()
	day := nowTime.Day()
	h24 := nowTime.Hour()
	h12 := h24 % 12
	min := nowTime.Minute()
	sec := nowTime.Second()
	replacement := map[string]interface{}{
		"[year]":  year,
		"[month]": month,
		"[day]":   day,
		"[h24]":   h24,
		"[h12]":   h12,
		"[min]":   min,
		"[sec]":   sec,
	}
	for _, t := range o.Cmds {
		rc := utils.FormateByRepalcment(t, replacement)
		//fmt.Println(rc)
		o.Frame.GetGameControl().SendCmd(rc)
	}
}

func (o *TimeSync) Inject(frame defines.MainFrame) {
	o.Frame = frame
}

func (o *TimeSync) Activate() {
	if o.GameTimeSyncConfig.Enable {
		t := time.NewTicker(time.Duration(o.GameTimeSyncConfig.SyncDuration) * time.Second)
		go func() {
			for {
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("/time set %v", o.getRealTimeInMCTick()))
				<-t.C
			}
		}()
	}
	if o.ScoreboardCalibrateConfig.Enable {
		t := time.NewTicker(time.Duration(o.GameTimeSyncConfig.SyncDuration) * time.Second)
		go func() {
			for {
				o.calibrateTime()
				<-t.C
			}
		}()
	}
}
