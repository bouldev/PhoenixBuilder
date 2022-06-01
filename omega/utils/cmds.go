package utils

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"reflect"
	"strings"
	"time"
)

func ParseAdaptiveCmd(c interface{}) (cmds []defines.Cmd, err error) {
	switch tc := c.(type) {
	case []string:
		cmds = []defines.Cmd{}
		for _, _c := range tc {
			cmds = append(cmds, defines.Cmd{
				Cmd:    _c,
				Record: "",
				As:     "WS",
			})
		}
		return cmds, nil
	case []interface{}:
		cmds = []defines.Cmd{}
		for _, _c := range tc {
			if sc, ok := _c.(string); ok {
				cmds = append(cmds, defines.Cmd{
					Cmd:    sc,
					Record: "",
					As:     "WS",
				})
				continue
			}
			marshal, err := json.Marshal(_c)
			if err != nil {
				return nil, err
			}
			nc := defines.Cmd{}
			err = json.Unmarshal(marshal, &nc)
			if err != nil {
				return nil, err
			}
			if nc.Record != "" && nc.Record != "无" && nc.Record != "空" &&
				nc.Record != "成功次数" &&
				nc.Record != "完整结果" {
				return nil, fmt.Errorf("结果记录 仅 可为\"空\"/\"成功次数\"/\"完整结果\"之一，你的设置是: %v", nc)
			}
			switch nc.As {
			case "":
				nc.As = "WS"
			case "WS":
			case "Websocket":
				nc.As = "WS"
			case "WebSocket":
				nc.As = "WS"
			case "websocket":
				nc.As = "WS"
			case "Player":
			case "player":
				nc.As = "Player"
			case "玩家":
				nc.As = "Player"
			default:
				return nil, fmt.Errorf("身份 仅 可为\"WS\"/\"Player\"/\"玩家\"之一，你的设置是: %v", nc)
			}
			cmds = append(cmds, nc)
		}
		return cmds, nil
	default:
		return nil, fmt.Errorf("无法理解的指令序列格式 %v, 期望: []string 或 []interface{}", reflect.TypeOf(c))
	}
}

func ParseAdaptiveJsonCmd(cfg map[string]interface{}, p []string) (cmds []defines.Cmd, err error) {
	var c interface{}
	_p := p
	c = cfg
	for len(p) != 0 {
		p0 := p[0]
		p = p[1:]
		if _c, ok := cfg[p0]; ok {
			c = _c
		} else {
			return nil, fmt.Errorf("需要的配置项路径完整路径为: %v, 但是无法找到路径 %v", _p, p0)
		}
	}
	return ParseAdaptiveCmd(c)
}

func LaunchCmdsArray(ctrl defines.GameControl, cmds []defines.Cmd, remapping map[string]interface{}, logger defines.LineDst) {
	for _, a := range cmds {
		if a.SleepBefore != 0 {
			time.Sleep(time.Duration(a.SleepBefore * float32(time.Second)))
		}
		time.Sleep(time.Duration(a.SleepBefore * float32(time.Second)))
		cmd := FormatByReplacingOccurrences(a.Cmd, remapping)
		if a.Record == "" || a.Record == "无" || a.Record == "空" {
			ctrl.SendCmd(cmd)
		} else {
			onResponse := func(output *packet.CommandOutput) {
				if a.Record == "成功次数" {
					logger.Write(fmt.Sprintf("[%v]=>success:[%v]", cmd, output.SuccessCount))
				} else {
					logger.Write(fmt.Sprintf("[%v]=>output:[%v]", cmd, output.OutputMessages))
				}
			}
			if a.As == "WS" {
				ctrl.SendCmdAndInvokeOnResponse(cmd, onResponse)
			} else {
				ctrl.SendCmdAndInvokeOnResponseWithFeedback(cmd, onResponse)
			}
		}
		if a.Sleep != 0 {
			time.Sleep(time.Duration(a.Sleep * float32(time.Second)))
		}
	}
}

func GetPlayerList(ctrl defines.GameControl, selector string, onResult func([]string)) {
	ctrl.SendCmdAndInvokeOnResponse("testfor "+selector, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 && len(output.OutputMessages) > 0 && len(output.OutputMessages[0].Parameters) > 0 {
			players := strings.Split(output.OutputMessages[0].Parameters[0], ", ")
			onResult(players)
			return
		}
		onResult([]string{})
	})
}
