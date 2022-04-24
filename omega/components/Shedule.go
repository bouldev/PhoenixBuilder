package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type Schedule struct {
	*BasicComponent
	Name     string        `json:"任务"`
	Duration int           `json:"周期"`
	actions  []defines.Cmd `json:"动作"`
	LogFile  string        `json:"结果记录文件"`
	logger   defines.LineDst
	stopC    chan struct{}
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
	utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.actions, map[string]interface{}{}, o.logger)
	//for _, _a := range o.Actions {
	//	a := _a
	//	cmd := utils.FormateByRepalcment(a.Cmd, map[string]interface{}{})
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

func (o *Schedule) Activate() {
	o.launchTask()
	if o.Duration == 0 {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("计划任务 %v 已退出, 因为周期为 0", o.Name))
		return
	}
	ticker := time.NewTicker(time.Duration(o.Duration) * time.Second)
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
