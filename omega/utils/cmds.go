package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"reflect"
	"regexp"
	"strconv"
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

func splitScoreFetchGroup(defaultTarget string, targets []string) [][]string {
	pairs := [][]string{}
	occurred := map[string]bool{}
	for _, origTarget := range targets {
		if _, hasK := occurred[origTarget]; hasK {
			continue
		} else {
			occurred[origTarget] = true
		}
		target := origTarget[7 : len(origTarget)-2]
		pair := strings.Split(target, ",")
		if len(pair) == 1 {
			pair = []string{defaultTarget, pair[0]}
		} else if len(pair) > 2 {
			pair = pair[:2]
		}
		pair = append(pair, origTarget)
		pairs = append(pairs, pair)
	}
	return pairs
}

var scoreTester = regexp.MustCompile(`\[score<.*?>\]`)

func LaunchCmdsArray(ctrl defines.GameControl, cmds []defines.Cmd, remapping map[string]interface{}, logger defines.LineDst) {
	needPeekSendingRateReduce := len(cmds) > 8
	scoreboardFetchTarget := ""
	if target, hasK := remapping["[player]"]; hasK {
		if strTarget, success := target.(string); success {
			scoreboardFetchTarget = strTarget
		}
	} else if target, hasK := remapping["[target_player]"]; hasK {
		if strTarget, success := target.(string); success {
			scoreboardFetchTarget = strTarget
		}
	}
	executeSucceed := false
	for i, a := range cmds {
		if a.Conditinal && !executeSucceed {
			continue
		}
		if a.SleepBefore != 0 {
			time.Sleep(time.Duration(a.SleepBefore * float32(time.Second)))
		} else if needPeekSendingRateReduce {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
		}
		conditional := false
		if i+1 != len(cmds) {
			if cmds[i+1].Conditinal {
				conditional = true
			}
		}
		cmd := FormatByReplacingOccurrences(a.Cmd, remapping)
		matches := scoreTester.FindAllString(cmd, -1)
		if len(matches) > 0 {
			pairs := splitScoreFetchGroup(scoreboardFetchTarget, matches)
			// fmt.Println(pairs)
			for _, pair := range pairs {
				_player, _scoreboard, _replacement := pair[0], pair[1], pair[2]
				waitChan := make(chan struct{})
				val := "not_found"
				checkCmd := fmt.Sprintf("scoreboard players add %v %v 0", _player, _scoreboard)
				// fmt.Println(checkCmd)
				ctrl.SendCmdAndInvokeOnResponse(checkCmd, func(output *packet.CommandOutput) {
					if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 4 {
						fmt.Printf("发现缺少记分板%v (%v)->(%v)", _scoreboard, checkCmd, output)
						close(waitChan)
						return
					}
					val = output.OutputMessages[0].Parameters[3]
					close(waitChan)
				})
				<-waitChan
				cmd = strings.ReplaceAll(cmd, _replacement, val)
			}
		}
		if (a.Record == "" || a.Record == "无" || a.Record == "空") && !conditional {
			ctrl.SendCmd(cmd)
		} else {
			waitChan := make(chan bool, 1)
			onResponse := func(output *packet.CommandOutput) {
				if output.SuccessCount == 0 {
					waitChan <- false
				} else {
					waitChan <- true
				}
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
			executeSucceed = <-waitChan
		}
		if a.Sleep != 0 {
			time.Sleep(time.Duration(a.Sleep * float32(time.Second)))
		} else if needPeekSendingRateReduce {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
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

func CheckPlayerMatchSelector(ctrl defines.GameControl, name, selector string) (success chan bool) {
	s := FormatByReplacingOccurrences(selector, map[string]interface{}{
		"[player]": "\"" + name + "\"",
	})
	c := make(chan bool)
	ctrl.SendCmdAndInvokeOnResponse(fmt.Sprintf("testfor %v", s), func(output *packet.CommandOutput) {
		if output.SuccessCount != 0 {
			c <- true
		} else {
			c <- false
		}
	})
	return c
}
func GetPlayerScore(ctrl defines.GameControl, player, scoreboard string, onResult func(val int, err error)) {
	ctrl.SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players add \"%v\" %v 0", player, scoreboard), func(output *packet.CommandOutput) {
		if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 4 {
			onResult(0, fmt.Errorf("没有相关记分板"))
			return
		}
		val, err := strconv.Atoi(output.OutputMessages[0].Parameters[3])
		if err != nil {
			onResult(0, fmt.Errorf("数据解析出错 %v", err))
			return
		} else {
			onResult(val, nil)
		}
	})
}

func GetBlockAt(ctrl defines.GameControl, pos string, result func(outOfWorld bool, isAir bool, name string, pos define.CubePos)) {
	ctrl.SendCmdAndInvokeOnResponse("testforblock "+pos+" air 0", func(output *packet.CommandOutput) {
		if output.SuccessCount != 0 {
			x, _ := strconv.Atoi(output.OutputMessages[0].Parameters[0])
			y, _ := strconv.Atoi(output.OutputMessages[0].Parameters[1])
			z, _ := strconv.Atoi(output.OutputMessages[0].Parameters[2])
			result(false, true, "air", define.CubePos{x, y, z})
			return
		} else {
			if len(output.OutputMessages) > 0 && output.OutputMessages[0].Message == "commands.testforblock.failed.tile" {
				if len(output.OutputMessages[0].Parameters) == 5 {
					if x, err := strconv.Atoi(output.OutputMessages[0].Parameters[0]); err == nil {
						if y, err := strconv.Atoi(output.OutputMessages[0].Parameters[1]); err == nil {
							if z, err := strconv.Atoi(output.OutputMessages[0].Parameters[2]); err == nil {
								frags := strings.Split(output.OutputMessages[0].Parameters[3], ".")
								if len(frags) == 1 {
									result(false, false, strings.ReplaceAll(frags[0], "%", ""), define.CubePos{x, y, z})
								} else {
									result(false, false, frags[1], define.CubePos{x, y, z})
								}
								return
							}
						}
					}
				}
			}
			result(true, false, "unknown", define.CubePos{0, 0, 0})
		}
	})
}
