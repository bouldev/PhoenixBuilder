package mainframe

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"

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
	BackendTriggers                []string          `json:"后台菜单触发词" yaml:"后台菜单触发词"`
	GameTriggers                   []string          `json:"游戏菜单触发词" yaml:"游戏菜单触发词"`
	HideMenuOptions                []string          `json:"不显示这些菜单项"`
	HintOnUnknownCmd               string            `json:"无法理解指令时提示" yaml:"无法理解指令时提示"`
	MenuHead                       string            `json:"菜单标题" yaml:"菜单标题"`
	BotTag                         string            `json:"机器人标签" yaml:"机器人标签"`
	MenuFormat                     string            `json:"菜单显示格式" yaml:"菜单显示格式"`
	MenuFormatWithMultipleTriggers string            `json:"多个触发词的菜单显示格式" yaml:"多个触发词的菜单显示格式"`
	WisperHint                     string            `json:"悄悄话菜单提示" yaml:"悄悄话菜单提示"`
	MenuTail                       string            `json:"菜单末尾" yaml:"菜单末尾"`
	MenuCloseText                  string            `json:"关闭菜单时的提示"`
	ErrorText                      string            `json:"输入有误时提示"`
	SelectText                     string            `json:"等待输入时提示"`
	HasNotMenuOptionText           string            `json:"没有菜单项时提示"`
	OpenMenuOnUnknownCmd           bool              `json:"在遇到未知指令时打开菜单" yaml:"在遇到未知指令时打开菜单"`
	ContinueAsking                 bool              `json:"菜单打开后是否继续询问操作"`
	MenuStructure                  interface{}       `json:"目录结构"`
	ForceOverwriteOptions          map[string]string `json:"强制修改菜单信息"`
	GSLHRSettings                  *GSLHRSettings    `json:"无序号的选项列表提示"`
	GSLHRWISettings                *GSLHRWISettings  `json:"带序号的选项列表提示"`
	GIRRSettings                   *GIRRSettings     `json:"整数范围选择提示"`
	GYNSettings                    *GYNSettings      `json:"要求确认时提示"`
	QFPNSettings                   *QFPNSettings     `json:"搜索玩家时提示"`
	replaceFn                      func(string) string
	menuRootNode                   *MenuRenderNode
	// componentDefaultTriggers       []string
}

func (m *Menu) popup() {
	me := pterm.Prefix{
		Text:  "",
		Style: pterm.NewStyle(pterm.BgLightBlue, pterm.FgLightWhite, pterm.Bold),
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
		Println(toWidth("游戏菜单", 80))
	triggerWords := m.omega.OmegaConfig.Trigger.TriggerWords
	defaultTrigger := m.omega.OmegaConfig.Trigger.DefaultTigger

	if len(triggerWords) == 0 {
		pterm.Error.Println("没有触发词")
	} else {
		pterm.Info.Println("默认触发词: ", defaultTrigger, " 可用触发词: [", strings.Join(triggerWords, "/ "), "]")
	}

	primary := pterm.NewStyle(pterm.FgBlue, pterm.BgDefault, pterm.Bold)

	for _, e := range m.omega.Reactor.GameMenuEntries {
		me.Text = toWidth("->", 4)
		//me.Text = toWidth(fmt.Sprintf("%v %v", defaultTrigger, e.Triggers[0]), 30)
		head := fmt.Sprintf("%v %v %v", defaultTrigger, e.Triggers[0], e.ArgumentHint)
		s := primary.Sprint(head) + ": \n    " + e.Usage
		alters := []string{}
		for _, t := range e.Triggers {
			if t == e.Triggers[0] {
				continue
			}
			alters = append(alters, fmt.Sprintf("%v %v", defaultTrigger, t))
		}
		if len(alters) > 1 {
			s += "\n    或者: " + strings.Join(alters, "/")
		}
		(&pterm.PrefixPrinter{Prefix: me}).Println(s)
	}
	if len(m.omega.Reactor.GameMenuEntries) == 0 {
		pterm.Warning.Println("没有可用项")
	}

	pterm.NewStyle(pterm.BgDarkGray, pterm.FgLightWhite, pterm.Bold).
		Println(toWidth("后台指令菜单", 80))
	for _, e := range m.omega.BackendMenuEntries {
		//me.Text = toWidth(strings.Join(e.Triggers, " / "), 30)
		// me.Text = toWidth(fmt.Sprintf("%d", i+1), 4)
		me.Text = toWidth("->", 4)
		s := primary.Sprint(pterm.Bold.Sprintf("%v %v", e.Triggers[0], e.ArgumentHint)) + ": \n    " + e.Usage
		alters := []string{}
		for _, t := range e.Triggers {
			if t == e.Triggers[0] {
				continue
			}
			alters = append(alters, fmt.Sprintf("%v", t))
		}
		if len(alters) > 1 {
			s += "\n    或者: " + strings.Join(alters, "/")
		}
		(&pterm.PrefixPrinter{Prefix: me}).Println(s)
	}
	// me.Text = toWidth("-", 4)
	(&pterm.PrefixPrinter{Prefix: me}).Println(primary.Sprint(pterm.Bold.Sprintf("exit ")) + ": \n    关闭系统")
	pterm.NewStyle(pterm.BgDarkGray, pterm.FgLightWhite, pterm.Bold).
		Println(toWidth("", 120))
}

func (m *Menu) popGameMenu(chat *defines.GameChat, node *MenuRenderNode) bool {
	pk := m.mainFrame.GetGameControl().GetPlayerKit(chat.Name)
	if len(chat.Msg) != 0 {
		msg := chat.Msg[0]
		for _, sm := range node.SubMenus {
			if sm.Trigger == msg {
				chat.Msg = chat.Msg[1:]
				return m.popGameMenu(chat, sm.ChildNode)
			}
		}
		pk.Say(m.HintOnUnknownCmd)
		if !m.OpenMenuOnUnknownCmd {
			return true
		}
	}
	pk.Say("Omega System (Phoenix Builder Embed) by: 2401PT@CMA ")
	pk.Say(m.replaceFn(fmt.Sprintf(m.MenuHead)))
	systemTrigger := m.omega.OmegaConfig.Trigger.DefaultTigger
	menuFmt := m.MenuFormat
	multipleFmt := m.MenuFormatWithMultipleTriggers
	currentI := 0
	available := []string{}
	actions := []func(ctrl *defines.GameChat) bool{}
	hasMenuOption := false
	for _, e := range node.RealComponentEntry {
		// 隐藏菜单项
		if isInHideList := func() bool {
			for _, hideOption := range m.HideMenuOptions {
				for _, trigger := range e.Triggers {
					if hideOption == trigger {
						return true
					}
				}
			}
			return false
		}; isInHideList() {
			continue
		}
		// pterm.Info.Println(e.Verification)
		if e.Verification != nil && e.Verification.Enable {
			if e.Verification.ByNameList != nil && len(e.Verification.ByNameList) > 0 {
				found := false
				for _, n := range e.Verification.ByNameList {
					if n == chat.Name {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			if e.Verification.BySelector != "" {
				select {
				case r := <-utils.CheckPlayerMatchSelector(m.mainFrame.GetGameControl(), chat.Name, e.Verification.BySelector):
					// pterm.Warning.Println(r)
					if !r {
						continue
					}
				case <-time.NewTimer(100 * time.Millisecond).C:
					continue
				}
			}
		}
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
		pk.Say(m.replaceFn(entry))
		hasMenuOption = true
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
		pk.Say(m.replaceFn(entry))
		hasMenuOption = true
		cn := sm.ChildNode
		actions = append(actions, func(newChat *defines.GameChat) bool {
			return m.popGameMenu(newChat, cn)
		})
		available = append(available, sm.Trigger)
	}
	if !hasMenuOption {
		pk.Say(m.replaceFn(m.HasNotMenuOptionText))
	}
	pk.Say(m.replaceFn(m.WisperHint))
	pk.Say(m.replaceFn(m.MenuTail))
	// fmt.Println(chat)
	if m.ContinueAsking && hasMenuOption {
		if player := m.mainFrame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
			hint, resolver := m.GenStringListHintResolverWithIndex(available)
			if player.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
				if i, cancel, err := resolver(chat.Msg); err == nil {
					if cancel {
						player.Say(m.MenuCloseText)
						return true
					}
					chat.Msg = chat.Msg[1:]
					return actions[i](chat)
				} else {
					player.Say(utils.FormatByReplacingOccurrences(m.ErrorText, map[string]interface{}{
						"[error]": err.Error(),
					}))
					return false
				}
			}) == nil {
				player.Say(utils.FormatByReplacingOccurrences(m.SelectText, map[string]interface{}{
					"[hint]": hint,
				}))
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

// func (m *Menu) debugDisplayMenuStructure(node *MenuRenderNode, prefix string) {
// 	for _, e := range node.RealComponentEntry {
// 		fmt.Println(prefix, e.Triggers, e.Usage)
// 	}
// 	for _, sm := range node.SubMenus {
// 		fmt.Println(prefix, ">", sm.Trigger, sm.Hint)
// 		m.debugDisplayMenuStructure(sm.ChildNode, prefix+"\t")
// 	}
// }

func (m *Menu) fresh() {
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
}

func (m *Menu) Activate() {
	// m.componentDefaultTriggers = make([]string, len(m.omega.Reactor.GameMenuEntries))
	m.fresh()
	m.omega.Reactor.freshMenu = m.fresh
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
	if cfg.Version == "0.0.1" {
		cfg.Configs["没有菜单项时提示"] = "没有可选项"
		cfg.Configs["关闭菜单时的提示"] = "已取消"
		cfg.Configs["输入有误时提示"] = "抱歉，我没明白你的意思,因为输入[error]"
		cfg.Configs["等待输入时提示"] = "可选项有[hint],请在下方输入:"
		cfg.Configs["不显示这些菜单项"] = []string{
			"需要被隐藏的菜单项1",
			"需要被隐藏的菜单项2..",
		}
		cfg.Configs["无序号的选项列表提示"] = map[string]string{
			"提示样式":       "[ [options]] 之一",
			"选项间隔符":      ", ",
			"取消选择触发词":    "取消",
			"选项为空时的提示":   "为空",
			"不在选项范围内的提示": "不在选项范围内",
		}
		cfg.Configs["带序号的选项列表提示"] = map[string]string{
			"提示样式":       "[ [options]] 之一,或者[intHint]",
			"选项样式":       "[i].[option]",
			"选项间隔符":      ", ",
			"取消选择触发词":    "取消",
			"选项为空时的提示":   "[没有可选项]",
			"不在选项范围内的提示": "不在可选范围内",
		}
		cfg.Configs["整数范围选择提示"] = map[string]string{
			"提示样式":       "[min]~[max]之间的整数",
			"取消选择触发词":    "取消",
			"选项为空时的提示":   "为空",
			"不在选项范围内的提示": "不在范围中",
			"输入无效时的提示":   "不是一个有效的数",
		}
		cfg.Configs["要求确认时提示"] = map[string]string{
			"提示样式":     "输入 是/y 同意; 否/n 拒绝",
			"选项为空时的提示": "为空",
			"输入无效时的提示": "不是有效的回答",
		}
		cfg.Configs["搜索玩家时提示"] = map[string]string{
			"选项样式":       "[i].[currentName] [historyName]",
			"取消选择触发词":    "取消",
			"要求输入玩家名时提示": "请输入目标玩家名,或者目标玩家名的一部分(或输入: [cancel] )",
			"要求选择玩家名时提示": "请选择下方的序号，或者目标玩家名(或输入: [cancel] )",
			"找不到玩家时提示":   "没有搜索到匹配的玩家，请输入目标玩家名,或者目标玩家名的一部分(或输入: [cancel] )",
		}
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	m.MenuStructure = []interface{}{}
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, m); err != nil {
		panic(err)
	}
	replaceElems := []string{}
	for k, v := range m.ForceOverwriteOptions {
		replaceElems = append(replaceElems, k)
		replaceElems = append(replaceElems, v)
	}
	replacer := strings.NewReplacer(replaceElems...)
	m.replaceFn = func(s string) string {
		return replacer.Replace(s)
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
	botName := ""
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     m.GameTriggers,
			Usage:        "打开菜单",
			FinalTrigger: true,
		},
		OptionalOnTriggerFn: func(chat *defines.GameChat) (stop bool) {
			if botName == "" {
				botName = m.mainFrame.GetUQHolder().GetBotName()
			}
			// fmt.Println(botName)
			if botName == chat.Name {
				return true
			}
			return m.popGameMenu(chat, m.menuRootNode)
		},
	})
	m.SetCollaborateFunc()
}
