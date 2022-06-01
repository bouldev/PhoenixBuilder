package mainframe

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"

	"github.com/pterm/pterm"
)

type MenuVirtualNode struct {
	Hint      string
	Trigger   string
	ChildNode *MenuRenderNode
}

type MenuRenderNode struct {
	RealComponentEntry []*defines.GameMenuEntry
	SubMenus           []*MenuVirtualNode
}

type Menu struct {
	*BaseCoreComponent
	BackendTriggers                []string    `json:"后台菜单触发词" yaml:"后台菜单触发词"`
	GameTriggers                   []string    `json:"游戏菜单触发词" yaml:"游戏菜单触发词"`
	HintOnUnknownCmd               string      `json:"无法理解指令时提示" yaml:"无法理解指令时提示"`
	MenuHead                       string      `json:"菜单标题" yaml:"菜单标题"`
	BotTag                         string      `json:"机器人标签" yaml:"机器人标签"`
	MenuFormat                     string      `json:"菜单显示格式" yaml:"菜单显示格式"`
	MenuFormatWithMultipleTriggers string      `json:"多个触发词的菜单显示格式" yaml:"多个触发词的菜单显示格式"`
	WisperHint                     string      `json:"悄悄话菜单提示" yaml:"悄悄话菜单提示"`
	MenuTail                       string      `json:"菜单末尾" yaml:"菜单末尾"`
	OpenMenuOnUnknownCmd           bool        `json:"在遇到未知指令时打开菜单" yaml:"在遇到未知指令时打开菜单"`
	ContinueAsking                 bool        `json:"菜单打开后是否继续询问操作"`
	MenuStructure                  interface{} `json:"目录结构"`
	menuRootNode                   *MenuRenderNode
	// componentDefaultTriggers       []string
}

func (m *Menu) popup() {
	me := pterm.Prefix{
		Text:  "",
		Style: &pterm.ThemeDefault.SuccessPrefixStyle,
	}
	toWidth := func(s string, w int) string {
		if len(s) > w {
			return s
		}
		h := (w - len(s)) / 2
		e := w - len(s) - h
		return strings.Repeat(" ", h) + s + strings.Repeat(" ", e)
	}
	pterm.NewStyle(pterm.BgDarkGray, pterm.FgLightWhite, pterm.Bold).
		Println(toWidth("后台指令菜单", 126))
	for i, e := range m.omega.BackendMenuEntries {
		//me.Text = toWidth(strings.Join(e.Triggers, " / "), 30)
		me.Text = toWidth(fmt.Sprintf("%d", i+1), 4)
		s := pterm.BgGray.Sprint(pterm.Bold.Sprintf("%v %v", e.Triggers[0], e.ArgumentHint)) + e.Usage
		alters := []string{}
		for _, t := range e.Triggers {
			if t == e.Triggers[0] {
				continue
			}
			alters = append(alters, fmt.Sprintf("%v", t))
		}
		if len(alters) > 1 {
			s += "\n\t- 或者: " + strings.Join(alters, "/")
		}
		(&pterm.PrefixPrinter{Prefix: me}).Println(s)
	}
	me.Text = toWidth("exit", 4)
	(&pterm.PrefixPrinter{Prefix: me}).Println(pterm.BgGray.Sprint(pterm.Bold.Sprintf("exit ")) + "关闭系统")
	pterm.NewStyle(pterm.BgDarkGray, pterm.FgLightWhite, pterm.Bold).
		Println(toWidth("游戏菜单", 124))
	triggerWords := m.omega.OmegaConfig.Trigger.TriggerWords
	defaultTrigger := m.omega.OmegaConfig.Trigger.DefaultTigger

	if len(triggerWords) == 0 {
		pterm.Error.Println("没有触发词")
	} else {
		pterm.Info.Println("默认触发词: ", defaultTrigger, " 可用触发词: [", strings.Join(triggerWords, "/ "), "]")
	}

	for i, e := range m.omega.Reactor.GameMenuEntries {
		me.Text = toWidth(fmt.Sprintf("%d", i+1), 4)
		//me.Text = toWidth(fmt.Sprintf("%v %v", defaultTrigger, e.Triggers[0]), 30)
		head := fmt.Sprintf("%v %v %v", defaultTrigger, e.Triggers[0], e.ArgumentHint)
		s := pterm.Bold.Sprint(pterm.BgGray.Sprint(head)) + " " + e.Usage
		alters := []string{}
		for _, t := range e.Triggers {
			if t == e.Triggers[0] {
				continue
			}
			alters = append(alters, fmt.Sprintf("%v %v", defaultTrigger, t))
		}
		if len(alters) > 1 {
			s += "\n\t- 或者: " + strings.Join(alters, "/")
		}
		(&pterm.PrefixPrinter{Prefix: me}).Println(s)
	}
	if len(m.omega.Reactor.GameMenuEntries) == 0 {
		pterm.Warning.Println("没有可用项")
	}
	pterm.NewStyle(pterm.BgDarkGray, pterm.FgLightWhite, pterm.Bold).
		Println(toWidth("", 120))
}

func (m *Menu) popGameMenu(chat *defines.GameChat, node *MenuRenderNode) bool {
	pk := m.mainFrame.GetGameControl().GetPlayerKit(chat.Name)
	if len(chat.Msg) != 0 {
		pk.Say(m.HintOnUnknownCmd)
		if !m.OpenMenuOnUnknownCmd {
			return true
		}
	}
	pk.Say("Omega · Async Rental Server Auxiliary · System · Author: §l2401PT")
	pk.Say("基于 PhoenixBuilder, 原型来自 CMA 服务器的 Omega 系统，此处感谢 CMA 的小伙伴们")
	pk.Say(fmt.Sprintf(m.MenuHead))
	systemTrigger := m.omega.OmegaConfig.Trigger.DefaultTigger
	menuFmt := m.MenuFormat
	multipleFmt := m.MenuFormatWithMultipleTriggers
	currentI := 0
	available := []string{}
	actions := []func(ctrl *defines.GameChat) bool{}
	for _, e := range node.RealComponentEntry {
		currentI++
		i := currentI
		tmp := menuFmt
		if len(e.Triggers) > 1 {
			tmp = multipleFmt
		}
		//fmt.Println(tmp)
		entry := utils.FormatByReplacingOccurrences(tmp, map[string]interface{}{
			"[i]":              i,
			"[systemTrigger]":  systemTrigger,
			"[defaultTrigger]": e.Triggers[0],
			"[usage]":          e.Usage,
			"[allTriggers]":    "[" + strings.Join(e.Triggers, "/") + "]",
			"[argumentHint]":   e.ArgumentHint,
		})
		//fmt.Println(entry)
		pk.Say(entry)
		actions = append(actions, e.OptionalOnTriggerFn)
		available = append(available, e.Triggers[0])
	}
	for _, sm := range node.SubMenus {
		currentI++
		i := currentI
		entry := utils.FormatByReplacingOccurrences(sm.Hint, map[string]interface{}{
			"[i]":              i,
			"[systemTrigger]":  systemTrigger,
			"[defaultTrigger]": sm.Trigger,
		})
		pk.Say(entry)
		cn := sm.ChildNode
		actions = append(actions, func(newChat *defines.GameChat) bool {
			return m.popGameMenu(newChat, cn)
		})
		available = append(available, sm.Trigger)
	}
	pk.Say(fmt.Sprintf(m.WisperHint))
	pk.Say(fmt.Sprintf(m.MenuTail))
	fmt.Println(chat)
	if m.ContinueAsking {
		if player := m.mainFrame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
			hint, resolver := utils.GenStringListHintResolverWithIndex(available)
			if player.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
				if i, cancel, err := resolver(chat.Msg); err == nil {
					if cancel {
						player.Say("已取消")
						return true
					}
					chat.Msg = chat.Msg[1:]
					return actions[i](chat)
				} else {
					player.Say("抱歉，我没明白你的意思,因为输入" + err.Error())
					return false
				}
			}) == nil {
				player.Say("可选项有" + hint + ",请在下方输入:")
			}
		}
	}
	return true
}

func (m *Menu) buildMenuStructure(currentNode *MenuRenderNode, currentStructure interface{}, attachPoints map[string]func(*defines.GameMenuEntry)) {
	sl, success := currentStructure.([]interface{})
	if !success {
		panic(fmt.Errorf("%v必须为列表形式", currentStructure))

	}
	for _, s := range sl {
		switch st := s.(type) {
		case string:
			if currentNode.RealComponentEntry == nil {
				currentNode.RealComponentEntry = make([]*defines.GameMenuEntry, 0)
			}
			attachPoints[st] = func(e *defines.GameMenuEntry) {
				currentNode.RealComponentEntry = append(currentNode.RealComponentEntry, e)
			}
		case map[string]interface{}:
			if currentNode.SubMenus == nil {
				currentNode.SubMenus = make([]*MenuVirtualNode, 0)
			}
			child := &MenuVirtualNode{
				Hint:      st["菜单项"].(string),
				Trigger:   st["触发词"].(string),
				ChildNode: &MenuRenderNode{},
			}
			currentNode.SubMenus = append(currentNode.SubMenus, child)
			m.buildMenuStructure(child.ChildNode, st["子节点"], attachPoints)
		default:
			panic(fmt.Errorf("%v", st))
		}
	}
}

func (m *Menu) debugDisplayMenuStructure(node *MenuRenderNode, prefix string) {
	for _, e := range node.RealComponentEntry {
		fmt.Println(prefix, e.Triggers, e.Usage)
	}
	for _, sm := range node.SubMenus {
		fmt.Println(prefix, ">", sm.Trigger, sm.Hint)
		m.debugDisplayMenuStructure(sm.ChildNode, prefix+"\t")
	}
}

func (m *Menu) Activate() {
	// m.componentDefaultTriggers = make([]string, len(m.omega.Reactor.GameMenuEntries))
	currentAllTriggers := make([]string, 0)
	m.menuRootNode = &MenuRenderNode{}
	attachPoints := make(map[string]func(*defines.GameMenuEntry))
	m.buildMenuStructure(m.menuRootNode, m.MenuStructure, attachPoints)
	for _, e := range m.omega.Reactor.GameMenuEntries {
		if len(e.Triggers) == 0 {
			panic(fmt.Errorf("游戏目录项:%v 缺少触发词", e))
		} else {
			i := len(currentAllTriggers)
			for _, t := range e.Triggers {
				for _, ct := range currentAllTriggers[:i] {
					if strings.HasPrefix(t, ct) {
						if ct == t {
							pterm.Error.Printfln("触发词冲突 %v 出现了两次或更多次", e)
						} else {
							pterm.Error.Printfln("触发词冲突:触发词 %v 会被曲解为触发词 %v", t, ct)
						}
					}
				}
				currentAllTriggers = append(currentAllTriggers, t)
			}
			defaultTrigger := e.Triggers[0]
			if attachFn, hasK := attachPoints[defaultTrigger]; hasK {
				attachFn(e)
			} else {
				m.menuRootNode.RealComponentEntry = append(m.menuRootNode.RealComponentEntry, e)
			}
			// m.componentDefaultTriggers = append(m.componentDefaultTriggers, defaultTrigger)
		}
	}

	// m.debugDisplayMenuStructure(m.menuRootNode, "")

	for _, e := range m.omega.BackendMenuEntries {
		if len(e.Triggers) == 0 {
			panic(fmt.Errorf("后台目录项:%v 缺少触发词", e))
		}
	}

	m.BaseCoreComponent.Activate()
	m.mainFrame.GetGameControl().SendCmd("tag @s add " + m.BotTag)

}

func (m *Menu) Init(cfg *defines.ComponentConfig) {
	m.MenuStructure = []interface{}{}
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, m); err != nil {
		panic(err)
	}
}

func (m *Menu) Inject(frame defines.MainFrame) {
	m.mainFrame = frame
	frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     m.BackendTriggers,
			Usage:        "打开菜单",
			FinalTrigger: true,
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			m.popup()
			return true
		},
	})
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     m.GameTriggers,
			Usage:        "打开菜单",
			FinalTrigger: true,
		},
		OptionalOnTriggerFn: func(chat *defines.GameChat) (stop bool) {
			return m.popGameMenu(chat, m.menuRootNode)
		},
	})
}
