package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type PortalEntry struct {
	Time string `json:"time"`
	Pos  []int  `json:"pos"`
}

type Portal struct {
	*BasicComponent
	FileName      string   `json:"存档点记录文件名"`
	SaveTrigger   []string `json:"保存存档点触发词"`
	RemoveTrigger []string `json:"删除存档点触发词"`
	LoadTrigger   []string `json:"返回存档点触发词"`
	ListTrigger   []string `json:"列出存档点触发词"`
	Selector      string   `json:"条件选择器"`
	positions     map[string]map[string]*PortalEntry
}

func (o *Portal) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *Portal) getPlayerPositions(name string) map[string]*PortalEntry {
	if ps, hasK := o.positions[name]; hasK {
		return ps
	} else {
		o.positions[name] = map[string]*PortalEntry{}
		return o.positions[name]
	}
}

func (o *Portal) list(chat *defines.GameChat) bool {
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	pk.Say(" §l所有可以前往/加载的地点:")
	for n, p := range o.positions["*"] {
		pk.Say(fmt.Sprintf("  公共: §l§6%v §f§r(位于 %v)", n, p.Pos))
	}
	ps := o.getPlayerPositions(chat.Name)
	for n, p := range ps {
		pk.Say(fmt.Sprintf("  你的: §l§6%v §f§r(位于 %v)", n, p.Pos))
	}

	if pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		if !chat.FrameWorkTriggered {
			chat.FrameWorkTriggered = true
			o.Frame.GetGameListener().Throw(chat)
			return true
		}
		return false
	}) == nil {
		pk.Say(fmt.Sprintf("希望前往地点请输入 %v [地点名]\n希望增加地点请输入 %v [地点名]\n希望移除地点请输入 %v [地点名]", o.LoadTrigger[0], o.SaveTrigger[0], o.RemoveTrigger[0]))
	}
	return true
}

func (o *Portal) doTP(name string, pos string) bool {
	ps := o.getPlayerPositions(name)
	goPS := func(n string, p *PortalEntry) bool {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 前往地点 %v: %v", name, n, p))
		s := utils.FormateByRepalcment(o.Selector, map[string]interface{}{
			"[player]": name,
		})
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("tp %v %v %v %v", s, p.Pos[0], p.Pos[1], p.Pos[2]), func(output *packet.CommandOutput) {
			if output.SuccessCount != 0 {
				o.Frame.GetGameControl().SayTo(name, "§6传送成功")
			} else {
				o.Frame.GetGameControl().SayTo(name, "§4传送失败")
			}
		})
		return true
	}
	for n, _ := range ps {
		if n == pos {
			if goPS(n, ps[n]) {
				return true
			}
		}
	}
	for n, p := range o.positions["*"] {
		if n == pos {
			if goPS(n, p) {
				return true
			}
		}
	}
	o.Frame.GetGameControl().SayTo(name, "前往失败，因为没有那个地点")
	return false
}

func (o *Portal) tp(chat *defines.GameChat) bool {
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	if len(chat.Msg) > 0 {
		pos := chat.Msg[0]
		if o.doTP(chat.Name, pos) {
			return true
		}
	}

	ps := o.getPlayerPositions(chat.Name)
	names := []string{}
	for n, _ := range ps {
		names = append(names, n)
	}
	for n, _ := range o.positions["*"] {
		names = append(names, n)
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(names)
	if pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		i, cancel, err := resolver(chat.Msg)
		if cancel {
			pk.Say(fmt.Sprintf("已取消"))
			return true
		}
		if err != nil {
			pk.Say(fmt.Sprintf("无法前往你说的地点，因为输入%v", err))
			return true
		}
		o.doTP(chat.Name, names[i])
		return true
	}) == nil {
		pk.Say(fmt.Sprintf("可选的地点有 %v 请输入:", hint))
	}
	return true
}

func (o *Portal) doRemove(name string, pos string) bool {
	ps := o.getPlayerPositions(name)
	for n, _ := range ps {
		if n == pos {
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 移除了地点 %v: %v", name, n, ps[n]))
			o.Frame.GetGameControl().SayTo(name, "已移除")
			delete(ps, n)
			return true
		}
	}
	o.Frame.GetGameControl().SayTo(name, "移除失败，因为没有那个地点")
	return false
}

func (o *Portal) doAdd(name string, posName string) {
	pk := o.Frame.GetGameControl().GetPlayerKit(name)
	go func() {
		pos := <-pk.GetPos(o.Selector)
		if pos == nil {
			pk.Say("添加失败")
			return
		}
		if ps, hasK := o.positions[name]; hasK {
			ps[posName] = &PortalEntry{
				Time: utils.TimeToString(time.Now()),
				Pos:  pos,
			}
		} else {
			o.positions[name] = map[string]*PortalEntry{
				posName: &PortalEntry{
					Time: utils.TimeToString(time.Now()),
					Pos:  pos,
				},
			}
		}
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v 添加了地点 %v: %v", name, posName, o.positions[name][posName]))
		pk.Say("添加成功")
	}()
}

func (o *Portal) add(chat *defines.GameChat) bool {
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	if len(chat.Msg) > 0 {
		pos := chat.Msg[0]
		o.doAdd(chat.Name, pos)
		return true
	}

	if pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			o.doAdd(chat.Name, chat.Msg[0])
		}
		return true
	}) == nil {
		pk.Say("请输入这个地点的名字:")
	}
	return true
}

func (o *Portal) remove(chat *defines.GameChat) bool {
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	if len(chat.Msg) > 0 {
		pos := chat.Msg[0]
		if o.doRemove(chat.Name, pos) {
			return true
		}
	}

	ps := o.getPlayerPositions(chat.Name)
	names := []string{}
	for n, _ := range ps {
		names = append(names, n)
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(names)
	if pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		i, cancel, err := resolver(chat.Msg)
		if cancel {
			pk.Say(fmt.Sprintf("已取消"))
			return true
		}
		if err != nil {
			pk.Say(fmt.Sprintf("无法移除你说的地点，因为输入%v", err))
			return true
		}
		o.doRemove(chat.Name, names[i])
		return true
	}) == nil {
		pk.Say(fmt.Sprintf("可选的地点有 %v 请输入:", hint))
	}
	return true
}

func (o *Portal) Stop() error {
	fmt.Println("正在保存 " + o.FileName)
	return o.Frame.WriteJsonData(o.FileName, &o.positions)
}

func (o *Portal) Inject(frame defines.MainFrame) {
	o.Frame = frame
	err := frame.GetJsonData(o.FileName, &o.positions)
	if err != nil {
		panic(err)
	}
	if o.positions == nil {
		o.positions = map[string]map[string]*PortalEntry{}
	}
	if _, hasK := o.positions["*"]; !hasK {
		o.positions["*"] = map[string]*PortalEntry{
			"主城": {
				Time: utils.TimeToString(time.Now()),
				Pos:  []int{0, 252, 0},
			},
		}
	}
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.ListTrigger,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "显示所有可以去的地点",
		},
		OptionalOnTriggerFn: o.list,
	})
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.RemoveTrigger,
			ArgumentHint: "[地点]",
			FinalTrigger: false,
			Usage:        "移除一个保存的地点",
		},
		OptionalOnTriggerFn: o.remove,
	})
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.SaveTrigger,
			ArgumentHint: "[地点名]",
			FinalTrigger: false,
			Usage:        "以某个名字保存当前的地点",
		},
		OptionalOnTriggerFn: o.add,
	})
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.LoadTrigger,
			ArgumentHint: "[地点名]",
			FinalTrigger: false,
			Usage:        "前往指定的地点",
		},
		OptionalOnTriggerFn: o.tp,
	})
}
