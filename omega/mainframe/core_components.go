package mainframe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"
)

type BaseCoreComponent struct {
	cfg       *defines.ComponentConfig
	omega     *Omega
	mainFrame defines.MainFrame
}

func (c *BaseCoreComponent) Init(cfg *defines.ComponentConfig) {
	c.cfg = cfg
}

func (c *BaseCoreComponent) Activate() {}

func (c *BaseCoreComponent) Stop() error { return nil }

func (c *BaseCoreComponent) SetSystem(omega interface{}) {
	c.omega = omega.(*Omega)
}

type Menu struct {
	*BaseCoreComponent
	BackendTriggers                []string `json:"backend_triggers" yaml:"backend_triggers"`
	GameTriggers                   []string `json:"game_triggers" yaml:"game_triggers"`
	HintOnUnknownCmd               string   `json:"hint_on_unknown_cmd" yaml:"hint_on_unknown_cmd"`
	MenuHead                       string   `json:"menu_head" yaml:"menu_head"`
	BotTag                         string   `json:"bot_tag" yaml:"bot_tag"`
	MenuFormat                     string   `json:"menu_format" yaml:"menu_format"`
	MenuFormatWithMultipleTriggers string   `json:"multiple_trigger_menu_format" yaml:"menu_format_with_multiple_triggers"`
	WisperHint                     string   `json:"wisper_hint" yaml:"wisper_hint"`
	MenuTail                       string   `json:"menu_tail" yaml:"menu_tail"`
	OpenMenuOnUnknownCmd           bool     `json:"open_menu_on_unknown_cmd" yaml:"open_menu_on_unknown_cmd"`
	ContinueAsking                 bool     `json:"continue_asking"`
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
	triggerWords := m.mainFrame.QueryConfig("TriggerWords").([]string)
	defaultTrigger := m.mainFrame.QueryConfig("DefaultTigger").(string)

	if len(triggerWords) == 0 {
		pterm.Error.Println("没有触发词")
	} else {
		pterm.Info.Println("默认触发词: ", defaultTrigger, " 可用触发词: [", strings.Join(triggerWords, "/ "), "]")
		//if len(triggerWords) > 1 {
		//	pterm.Info.Println("任一触发词都具有同样的效果 ", strings.Join(triggerWords, " / "))
		//}
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

func (m *Menu) popGameMenu(chat *defines.GameChat) bool {
	pk := m.mainFrame.GetGameControl().GetPlayerKit(chat.Name)
	if len(chat.Msg) != 0 {
		pk.Say(m.HintOnUnknownCmd)
		if !m.OpenMenuOnUnknownCmd {
			return true
		}
	}
	pk.Say(fmt.Sprintf(m.MenuHead))
	systemTrigger := m.mainFrame.QueryConfig("DefaultTigger").(string)
	menuFmt := m.MenuFormat
	multipleFmt := m.MenuFormatWithMultipleTriggers
	for _i, e := range m.omega.Reactor.GameMenuEntries {
		i := _i + 1
		tmp := menuFmt
		if len(e.Triggers) > 1 {
			tmp = multipleFmt
		}
		entry := utils.FormateByRepalcment(tmp, map[string]interface{}{
			"[i]":              i,
			"[systemTrigger]":  systemTrigger,
			"[defaultTrigger]": e.Triggers[0],
			"[usage]":          e.Usage,
			"[allTriggers]":    "[" + strings.Join(e.Triggers, "/") + "]",
			"[argumentHint]":   e.ArgumentHint,
		})
		pk.Say(entry)
	}
	pk.Say(fmt.Sprintf(m.WisperHint))
	pk.Say(fmt.Sprintf(m.MenuTail))
	fmt.Println(chat)
	if m.ContinueAsking {
		if player := m.mainFrame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
			available := []string{}
			actions := []func(ctrl *defines.GameChat) bool{}
			for _, e := range m.omega.Reactor.GameMenuEntries {
				actions = append(actions, e.OptionalOnTriggerFn)
				available = append(available, e.Triggers[0])
			}
			hint, resolver := utils.GenStringListHintResolverWithIndex(available)
			if player.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
				if i, err := resolver(chat.Msg); err == nil {
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

func (m *Menu) Activate() {
	for _, e := range m.omega.Reactor.GameMenuEntries {
		if len(e.Triggers) == 0 {
			panic(fmt.Errorf("游戏目录项:%v 缺少触发词", e))
		}
	}
	for _, e := range m.omega.BackendMenuEntries {
		if len(e.Triggers) == 0 {
			panic(fmt.Errorf("后台目录项:%v 缺少触发词", e))
		}
	}
	m.BaseCoreComponent.Activate()
	m.mainFrame.GetGameControl().SendCmd("tag @s add " + m.BotTag)
}

func (m *Menu) Init(cfg *defines.ComponentConfig) {
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
		OptionalOnTriggerFn: m.popGameMenu,
	})
}

type CmdSender struct {
	*BaseCoreComponent
	Trigger string `json:"trigger"`
}

func (c *CmdSender) send(cmds []string) {
	cmd := strings.Join(cmds, " ")
	c.mainFrame.GetGameControl().SendCmdAndInvokeOnResponse(
		cmd,
		func(output *packet.CommandOutput) {
			terMsg := pterm.Info.Sprintf("/%v\n", cmd)
			for _, msg := range output.OutputMessages {
				if msg.Success {
					terMsg += pterm.Success.Sprintf("Msg: %v Params: %v\n", msg.Message, msg.Parameters)
				} else {
					terMsg += pterm.Error.Sprintf("Msg: %v Params: %v\n", msg.Message, msg.Parameters)
				}
			}
			c.mainFrame.GetBackendDisplay().Write(terMsg)
		},
	)
}

func (c *CmdSender) Init(cfg *defines.ComponentConfig) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, c); err != nil {
		panic(err)
	}
}

func (c *CmdSender) Inject(frame defines.MainFrame) {
	c.mainFrame = frame
	frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers: []string{c.Trigger},
			Usage:    fmt.Sprintf("发送指令，如果有可能性，显示结果， 例如 %vlist", c.Trigger),
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			c.send(cmds)
			return true
		},
	})
}

type NoSQLDBUtil struct {
	*BaseCoreComponent
}

func (o *NoSQLDBUtil) text2db(cmds []string) {
	if len(cmds) != 2 {
		fmt.Println("db text2db src_text dst_db")
	}
	src_text := cmds[0]
	dst_db := cmds[1]
	db := o.mainFrame.GetNoSqlDB(dst_db)
	if db == nil {
		fmt.Println("cannot open db")
	}
	file, err := os.OpenFile(src_text, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	src := bufio.NewReader(file)
	for {
		if line, _, err := src.ReadLine(); err == nil {
			objs := strings.Split(string(line), "\t")
			key := objs[0]
			value := strings.Join(objs[1:], "\t")
			db.Commit(key, value)
		} else {
			break
		}
	}
	fmt.Println("done")
}

func (o *NoSQLDBUtil) db2text(cmds []string) {
	if len(cmds) != 2 {
		fmt.Println("db db2text src_db dst_text")
	}
	src_db := cmds[0]
	dst_text := cmds[1]
	db := o.mainFrame.GetNoSqlDB(src_db)
	if db == nil {
		fmt.Println("cannot open db")
	}
	var err error
	var buf *bufio.Writer
	var file *os.File
	if dst_text != "screen" {
		file, err = os.OpenFile(dst_text, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
		buf = bufio.NewWriter(file)
	} else {
		buf = bufio.NewWriter(os.Stdout)
	}
	db.IterAll(func(key string, v string) (stop bool) {
		ma, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
			return false
		}
		buf.WriteString(key + "\t" + string(ma))
		return false
	})
	buf.Flush()
	if file != nil {
		file.Close()
	}
	fmt.Println("done")
}

func (o *NoSQLDBUtil) do(cmds []string) {

	if len(cmds) < 1 {
		fmt.Println("Opened dbs")
		for dbName, _ := range o.omega.OpenedDBs {
			fmt.Println(dbName)
		}
		fmt.Println("db text2db/db2text src dst")
		fmt.Println("db src delete/put key <value>")
		return
	}
	if cmds[0] == "text2db" {
		o.text2db(cmds[1:])
	} else if cmds[0] == "db2text" {
		o.db2text(cmds[1:])
	}
	targetDBName := cmds[0]
	availables := []string{}
	flag := false
	for dbName, _ := range o.omega.OpenedDBs {
		if targetDBName == dbName {
			flag = true
			break
		}
		availables = append(availables, dbName)
	}
	if !flag {
		fmt.Println(availables)
		return
	}
	db := o.mainFrame.GetNoSqlDB(targetDBName)
	if len(cmds) < 3 {
		fmt.Println("db put/delete key")
		return
	}
	op := cmds[1]
	key := cmds[2]
	if op == "delete" {
		db.Delete(key)
	} else if op == "put" {
		if len(cmds) != 4 {
			fmt.Println("db put/delete key value")
			return
		}
		db.Commit(key, cmds[3])
	}
}

func (o *NoSQLDBUtil) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers: []string{"db"},
			Usage:    "导出(db2text)和导入(text2db)数据库到可读的文本文件",
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			o.do(cmds)
			return true
		},
	})
}

type nameEntry struct {
	CurrentName    string   `json:"current_name"`
	LastUpdateTime string   `json:"last_update_time"`
	NameRecord     []string `json:"history"`
}

type NameRecord struct {
	*BaseCoreComponent
	Records  map[string]nameEntry
	FileName string `json:"file_name"`
}

func (o *NameRecord) Init(cfg *defines.ComponentConfig) {
	o.BaseCoreComponent.Init(cfg)
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *NameRecord) update(name, uuid string) {
	newTime := utils.TimeToString(time.Now())
	updateString := fmt.Sprintf("%v;%v", name, newTime)
	if player, hasK := o.Records[uuid]; hasK {
		if player.CurrentName != name {
			player.CurrentName = name
			player.LastUpdateTime = newTime
			o.mainFrame.GetBackendDisplay().Write(
				fmt.Sprintf("玩家%v改名了，曾用名为:%v", name, player.NameRecord),
			)
			player.NameRecord = append(player.NameRecord, updateString)
		}
	} else {
		o.Records[uuid] = nameEntry{
			CurrentName:    name,
			LastUpdateTime: newTime,
			NameRecord: []string{
				updateString,
			},
		}
	}
}

func (o *NameRecord) Stop() error {
	fmt.Println("正在保存 " + o.FileName)
	return o.mainFrame.WriteJsonData(o.FileName, o.Records)
}

func (o *NameRecord) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	o.Records = map[string]nameEntry{}
	err := frame.GetJsonData(o.FileName, &o.Records)
	if err != nil {
		panic(err)
	}
	frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		name, ud := entry.Username, entry.UUID
		name = utils.ToPlainName(name)
		o.update(name, ud.String())
	})
}

func (o *NameRecord) Activate() {
	o.BaseCoreComponent.Activate()
	for _, p := range o.mainFrame.GetUQHolder().PlayersByEntityID {
		name := utils.ToPlainName(p.Username)
		ud := p.UUID
		o.update(name, ud.String())
	}
}

func getCoreComponentsPool() map[string]func() defines.CoreComponent {
	return map[string]func() defines.CoreComponent{
		"Menu":        func() defines.CoreComponent { return &Menu{BaseCoreComponent: &BaseCoreComponent{}} },
		"CmdSender":   func() defines.CoreComponent { return &CmdSender{&BaseCoreComponent{}, "/"} },
		"NoSQLDBUtil": func() defines.CoreComponent { return &NoSQLDBUtil{&BaseCoreComponent{}} },
		"NameRecord":  func() defines.CoreComponent { return &NameRecord{BaseCoreComponent: &BaseCoreComponent{}} },
	}
}
