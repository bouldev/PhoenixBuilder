package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type cdKeyRecord struct {
	Total       int  `json:"可领取次数"`
	AllowRetake bool `json:"可重复领取"`
	TotalTaken  int
	Cmds        []string `json:"指令"`
	PlayerTaken map[string][]*cdKeyTakenRecord
}

type cdKeyTakenRecord struct {
	Time string
	Name string
}

type CDkey struct {
	*BasicComponent
	Triggers        []string                `json:"触发词"`
	Usage           string                  `json:"菜单提示"`
	CDKeys          map[string]*cdKeyRecord `json:"兑换码"`
	HintOnInvalid   string                  `json:"兑换码无效时提示"`
	HintOnRetake    string                  `json:"不可重复兑换时提示"`
	HintOnRateLimit string                  `json:"领取次数到达上限时提示"`
	FileName        string                  `json:"兑换码领取记录文件"`
}

func (o *CDkey) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	o.CDKeys = make(map[string]*cdKeyRecord)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	for _, r := range o.CDKeys {
		r.PlayerTaken = make(map[string][]*cdKeyTakenRecord)
		if r.Total < 0 {
			fmt.Println(r, " 的可领取次数不能为负数，如果希望能无限次领取可以设为0")
		}
	}
}
func (o *CDkey) doRedeem(player string, cmds []string, current, total int) {
	for _, cmd := range cmds {
		res := "无限"
		totalS := fmt.Sprintf("%d", total)
		if total == 0 {
			totalS = "无限"
		} else {
			res = fmt.Sprintf("%d", total-current)
		}
		c := utils.FormatByReplacingOccurrences(cmd, map[string]interface{}{
			"[player]":  player,
			"[current]": current,
			"[total]":   totalS,
			"[res]":     res,
		})
		o.Frame.GetGameControl().SendCmd(c)
	}
}

func (o *CDkey) redeem(chat *defines.GameChat) bool {
	if len(chat.Msg) > 0 {
		key := chat.Msg[0]
		if redeem, hasK := o.CDKeys[key]; hasK {
			if redeem.Total > 0 && redeem.Total == redeem.TotalTaken {
				o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnRateLimit)
			} else {
				uuid := ""
				for _, p := range o.Frame.GetUQHolder().PlayersByEntityID {
					if chat.Name == p.Username {
						uuid = p.UUID.String()
					}
				}
				if uuid == "" {
					fmt.Println("cannot obatin player uuid")
					return true
				}
				if takes, hasK := redeem.PlayerTaken[uuid]; hasK {
					if len(takes) > 0 && !redeem.AllowRetake {
						o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnRetake)
						return true
					}
					redeem.PlayerTaken[uuid] = append(redeem.PlayerTaken[uuid],
						&cdKeyTakenRecord{
							Time: utils.TimeToString(time.Now()),
							Name: chat.Name,
						},
					)
					redeem.TotalTaken++
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("player %v 重复兑换 %v (%v/%v)", chat.Name, key, redeem.TotalTaken, redeem.Total))
					o.doRedeem(chat.Name, redeem.Cmds, redeem.TotalTaken, redeem.Total)
				} else {
					redeem.PlayerTaken[uuid] = []*cdKeyTakenRecord{
						&cdKeyTakenRecord{
							Time: utils.TimeToString(time.Now()),
							Name: chat.Name,
						},
					}
					redeem.TotalTaken++
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("player %v 首次兑换 %v (%v/%v)", chat.Name, key, redeem.TotalTaken, redeem.Total))
					o.doRedeem(chat.Name, redeem.Cmds, redeem.TotalTaken, redeem.Total)
				}
			}
		} else {
			o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnInvalid)
		}
	} else {
		if o.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(c *defines.GameChat) (catch bool) {
			o.redeem(c)
			return true
		}) == nil {
			o.Frame.GetGameControl().SayTo(chat.Name, "请输入兑换码")
		}
	}
	return true
}

func (o *CDkey) Inject(frame defines.MainFrame) {
	o.Frame = frame
	playerTaken := map[string]map[string][]*cdKeyTakenRecord{}
	err := frame.GetJsonData(o.FileName, &playerTaken)
	if err != nil {
		panic(err)
	}
	for cdkey, players := range playerTaken {
		if r, hasK := o.CDKeys[cdkey]; hasK {
			r.PlayerTaken = players
			totalTaken := 0
			for _, p := range players {
				totalTaken += len(p)
			}
			r.TotalTaken = totalTaken
		}
	}
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[兑换码]",
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.redeem,
	})
}

func (o *CDkey) Stop() error {
	fmt.Println("正在保存 " + o.FileName)
	playerTaken := map[string]map[string][]*cdKeyTakenRecord{}
	for cdKey, record := range o.CDKeys {
		playerTaken[cdKey] = record.PlayerTaken
	}
	return o.Frame.WriteJsonData(o.FileName, playerTaken)
}
