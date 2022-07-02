package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"

	"github.com/pterm/pterm"
)

type Schedule struct {
	*defines.BasicComponent
	Name            string        `json:"任务"`
	Duration        float32       `json:"周期"`
	StartTimeInReal string        `json:"第一次启动的现实时间"`
	actions         []defines.Cmd `json:"动作"`
	LogFile         string        `json:"结果记录文件"`
	logger          defines.LineDst
	stopC           chan struct{}
}

func (o *Schedule) Init(cfg *defines.ComponentConfig) {
	marshal, err := json.Marshal(cfg.Configs)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(marshal, o)
	if err != nil {
		panic(err)
	}
	//if o.Actions == nil {
	//	panic(fmt.Sprintf("nil Actions in %v is not allowed", o.Name))
	//}
	//for _, a := range o.Actions {
	//	if a.Record != "" && a.Record != "无" && a.Record != "空" &&
	//		a.Record != "成功次数" &&
	//		a.Record != "完整结果" {
	//		panic(fmt.Errorf("结果记录 仅 可为\"空\"/\"成功次数\"/\"完整结果\"之一，你的设置是: %v", a))
	//	}
	//}
	if o.actions, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"动作"}); err != nil {
		panic(err)
	}
}

func (o *Schedule) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if o.LogFile == "" {
		o.logger = frame.GetBackendDisplay()
	} else {
		o.logger = &utils.MultipleLogger{Loggers: []defines.LineDst{
			frame.GetBackendDisplay(),
			frame.GetLogger(o.LogFile),
		}}
	}
	o.stopC = make(chan struct{})
}

func (o *Schedule) Stop() error {
	close(o.stopC)
	return nil
}

func (o *Schedule) launchTask() {
	nowTime := time.Now()
	year := nowTime.Year()
	month := int(nowTime.Month())
	day := nowTime.Day()
	h24 := nowTime.Hour()
	min := nowTime.Minute()
	sec := nowTime.Second()
	replacement := map[string]interface{}{
		"[year]":  year,
		"[month]": fmt.Sprintf("%02d", month),
		"[day]":   fmt.Sprintf("%02d", day),
		"[h24]":   fmt.Sprintf("%02d", h24),
		"[min]":   fmt.Sprintf("%02d", min),
		"[sec]":   fmt.Sprintf("%02d", sec),
	}
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.actions, replacement, o.logger)
	//for _, _a := range o.Actions {
	//	a := _a
	//	cmd := utils.FormatByReplacingOccurrences(a.Cmd, map[string]interface{}{})
	//	if a.Record == "" || a.Record == "无" || a.Record == "空" {
	//		o.Frame.GetGameControl().SendCmd(cmd)
	//	} else {
	//		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
	//			if a.Record == "成功次数" {
	//				o.logger.Write(fmt.Sprintf("[%v]:[%v]=>success:[%v]", o.Name, cmd, output.SuccessCount))
	//			} else {
	//				o.logger.Write(fmt.Sprintf("[%v]:[%v]=>output:[%v]", o.Name, cmd, output.OutputMessages))
	//			}
	//		})
	//	}
	//	time.Sleep(time.Duration(a.Sleep) * time.Second)
	//}
}

func (o *Schedule) TryOffset(offset time.Duration) (time.Duration, bool) {
	nt := time.Now()
	nowTime := nt.Add(offset)
	year := nowTime.Year()
	month := int(nowTime.Month())
	day := nowTime.Day()
	h24 := nowTime.Hour()
	min := nowTime.Minute()
	sec := nowTime.Second()
	replacement := map[string]interface{}{
		"[year]":  year,
		"[month]": fmt.Sprintf("%02d", month),
		"[day]":   fmt.Sprintf("%02d", day),
		"[h24]":   fmt.Sprintf("%02d", h24),
		"[min]":   fmt.Sprintf("%02d", min),
		"[sec]":   fmt.Sprintf("%02d", sec),
	}
	// fmt.Println(replacement)
	// fmt.Println(o.StartTimeInReal)
	markedTime := utils.FormatByReplacingOccurrences(o.StartTimeInReal, replacement)
	// fmt.Println(markedTime)
	// fmt.Println(markedTime)
	// fmt.Println("------")
	if baseT, err := time.Parse("2006-01-02 15:04:05 -0700 MST", markedTime+" +0800 CST"); err != nil {
		panic(fmt.Sprintf("第一次启动的现实时间 %v 格式不正确，\n"+
			"应该类似 [year]-[month]-[day] [h24]:04:05   (在最近一小时的 4分5秒第一次启动)\n"+
			"或者类似 [year]-[month]-[day] 00:04:05 (在最近一天的 凌晨4分5秒第一次启动),%v\n", o.StartTimeInReal, err))
	} else {
		baseT := baseT.Local()
		if baseT.After(time.Now()) {
			return baseT.Sub(time.Now()), true
		}
	}

	return 0, false
}

func (o *Schedule) GetRealStartTime() time.Duration {
	offset := []time.Duration{
		0, time.Second, time.Minute, time.Hour,
		time.Now().AddDate(0, 0, 1).Sub(time.Now()),
		time.Now().AddDate(0, 1, 1).Sub(time.Now()),
		time.Now().AddDate(1, 0, 0).Sub(time.Now()),
	}
	for _, off := range offset {
		if d, success := o.TryOffset(off); success {
			return d
		}
	}
	panic(fmt.Sprintf("第一次启动的现实时间 %v 格式不正确，\n"+
		"应该类似 [year]-[month]-[day] [h24]:04:05   (在最近一小时的 4分5秒第一次启动)\n"+
		"或者类似 [year]-[month]-[day] 00:04:05 (在最近一天的 凌晨4分5秒第一次启动)\n", o.StartTimeInReal))

}

func (o *Schedule) doTick() {
	o.launchTask()
	if o.Duration == 0 {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("计划任务 %v 已退出, 因为周期为 0", o.Name))
		return
	}
	ticker := time.NewTicker(time.Duration(o.Duration * float32(time.Second)))
	for {
		select {
		case <-ticker.C:
			go o.launchTask()
		case <-o.stopC:
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("计划任务 %v 已退出", o.Name))
			return
		}
	}
}

func (o *Schedule) Activate() {
	if o.StartTimeInReal != "" {
		go func() {
			d := o.GetRealStartTime()
			// fmt.Println(d)
			startTimeStr := time.Now().Add(d).Format("2006-01-02 15:04:05")
			o.Frame.GetBackendDisplay().Write(pterm.Info.Sprintf(
				"计划任务 %v 将于 %d 秒后 (%v) 第一次执行，随后每隔 %v 秒执行一次",
				o.Name, int(d.Seconds()), startTimeStr, o.Duration,
			))
			time.Sleep(d)
			o.Frame.GetBackendDisplay().Write(pterm.Info.Sprintf(
				"计划任务 %v 于 %v 第一次执行，随后每隔 %v 秒执行一次 (下一次时间为 %v)",
				o.Name, time.Now().Format("2006-01-02 15:04:05"), o.Duration, time.Now().Add(time.Duration(o.Duration*float32(time.Second))).Format("2006-01-02 15:04:05"),
			))
			o.doTick()
		}()
	} else {
		o.doTick()
	}
}
