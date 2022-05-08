package mainframe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"

	"github.com/pterm/pterm"
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
	BackendTriggers                []string `json:"后台菜单触发词" yaml:"后台菜单触发词"`
	GameTriggers                   []string `json:"游戏菜单触发词" yaml:"游戏菜单触发词"`
	HintOnUnknownCmd               string   `json:"无法理解指令时提示" yaml:"无法理解指令时提示"`
	MenuHead                       string   `json:"菜单标题" yaml:"菜单标题"`
	BotTag                         string   `json:"机器人标签" yaml:"机器人标签"`
	MenuFormat                     string   `json:"菜单显示格式" yaml:"菜单显示格式"`
	MenuFormatWithMultipleTriggers string   `json:"多个触发词的菜单显示格式" yaml:"多个触发词的菜单显示格式"`
	WisperHint                     string   `json:"悄悄话菜单提示" yaml:"悄悄话菜单提示"`
	MenuTail                       string   `json:"菜单末尾" yaml:"菜单末尾"`
	OpenMenuOnUnknownCmd           bool     `json:"在遇到未知指令时打开菜单" yaml:"在遇到未知指令时打开菜单"`
	ContinueAsking                 bool     `json:"菜单打开后是否继续询问操作"`
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
	pk.Say("Omega · Async Rental Server Auxiliary · System · Author: §l2401PT")
	pk.Say("基于 PhoenixBuilder, 原型来自 CMA 服务器的 Omega 系统，此处感谢 CMA 的小伙伴们")
	pk.Say(fmt.Sprintf(m.MenuHead))
	systemTrigger := m.omega.OmegaConfig.Trigger.DefaultTigger
	menuFmt := m.MenuFormat
	multipleFmt := m.MenuFormatWithMultipleTriggers
	for _i, e := range m.omega.Reactor.GameMenuEntries {
		i := _i + 1
		tmp := menuFmt
		if len(e.Triggers) > 1 {
			tmp = multipleFmt
		}
		//fmt.Println(tmp)
		entry := utils.FormateByRepalcment(tmp, map[string]interface{}{
			"[i]":              i,
			"[systemTrigger]":  systemTrigger,
			"[defaultTrigger]": e.Triggers[0],
			"[usage]":          e.Usage,
			"[allTriggers]":    "[" + strings.Join(e.Triggers, "/") + "]",
			"[argumentHint]":   e.ArgumentHint,
		})
		//fmt.Println(entry)
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
	PlayerTrigger    string `json:"以玩家身份发送信息的前缀"`
	WebsocketTrigger string `json:"以Websocket身份发送信息的前缀"`
}

func (c *CmdSender) send(cmds []string, ws bool) {
	cmd := strings.Join(cmds, " ")
	onFeedBack := func(output *packet.CommandOutput) {
		terMsg := pterm.Info.Sprintf("/%v\n", cmd)
		for _, msg := range output.OutputMessages {
			if msg.Success {
				terMsg += pterm.Success.Sprintf("Msg: %v Params: %v\n", msg.Message, msg.Parameters)
			} else {
				terMsg += pterm.Error.Sprintf("Msg: %v Params: %v\n", msg.Message, msg.Parameters)
			}
		}
		c.mainFrame.GetBackendDisplay().Write(terMsg)
	}
	if ws {
		c.mainFrame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, onFeedBack)
	} else {
		c.mainFrame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(cmd, onFeedBack)
	}
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
			Triggers: []string{c.WebsocketTrigger},
			Usage:    fmt.Sprintf("以 webscoket 身份发送指令，如果有可能性，显示结果， 例如 %vlist", c.WebsocketTrigger),
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			c.send(cmds, true)
			return true
		},
	})
	frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers: []string{c.PlayerTrigger},
			Usage:    fmt.Sprintf("以玩家身份发送指令，临时打开命令返回以显示结果， 例如 %vlist", c.PlayerTrigger),
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			c.send(cmds, false)
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
		if _line, _, err := src.ReadLine(); err == nil {
			line := strings.TrimSpace(string(_line))
			objs := strings.Split(line, "\t")
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
		buf.WriteString(key + "\t" + string(ma) + "\n")
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
	Records  map[string]*nameEntry
	FileName string `json:"改名历史记录文件"`
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
		o.Records[uuid] = &nameEntry{
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
	o.Records = map[string]*nameEntry{}
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

type KeepAlive struct {
	*BaseCoreComponent
	Schedule int `json:"检测周期"`
	Delay    int `json:"最大延迟"`
	replay   bool
}

func (o *KeepAlive) Init(cfg *defines.ComponentConfig) {
	o.BaseCoreComponent.Init(cfg)
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *KeepAlive) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	o.mainFrame.GetGameListener().SetGameChatInterceptor(func(chat *defines.GameChat) (stop bool) {
		if len(chat.Msg) > 0 && chat.Msg[0] == "alive" {
			o.replay = true
			return true
		}
		return false
	})
}

func (o *KeepAlive) Activate() {
	o.BaseCoreComponent.Activate()
	t := time.NewTicker(time.Second * time.Duration(o.Schedule))
	go func() {
		for {
			<-t.C
			o.replay = false
			o.mainFrame.GetGameControl().SendCmdAndInvokeOnResponse("w @s alive", func(output *packet.CommandOutput) {
				o.replay = true
			})
			<-time.NewTimer(time.Second * time.Duration(o.Delay)).C
			if o.replay == false {
				o.mainFrame.GetBackendDisplay().Write("连接检查失败，疑似Omega假死，准备退出...")
				o.omega.Stop()
				fmt.Println("3秒后退出")
				<-time.NewTimer(time.Second * 3).C
				panic("Omega 假死，已退出...\n" +
					"可以配合如下启动指令使 Omega 系统自动循环重启 (注意，延迟不得小于30秒，否则可能被封号)\n" +
					"对于windows系统： \n" +
					"for /l %i in (0,0,1) do @fastbuilder-windows.exe --omega_system -c 租赁服号 & @TIMEOUT /T 30 /NOBREAK\n" +
					"对于其他系统：\n" +
					"while true; ./fastbuilder -c 租赁服号 --omega_system; sleep 30; done\n")
			}
		}
	}()

}

func getCoreComponentsPool() map[string]func() defines.CoreComponent {
	return map[string]func() defines.CoreComponent{
		"菜单显示":      func() defines.CoreComponent { return &Menu{BaseCoreComponent: &BaseCoreComponent{}} },
		"指令发送":      func() defines.CoreComponent { return &CmdSender{BaseCoreComponent: &BaseCoreComponent{}} },
		"数据库导入导出工具": func() defines.CoreComponent { return &NoSQLDBUtil{&BaseCoreComponent{}} },
		"改名记录":      func() defines.CoreComponent { return &NameRecord{BaseCoreComponent: &BaseCoreComponent{}} },
		"假死检测":      func() defines.CoreComponent { return &KeepAlive{BaseCoreComponent: &BaseCoreComponent{}} },
	}
}
