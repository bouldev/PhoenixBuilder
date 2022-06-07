package mainframe

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/collaborate"
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

//
//func (o *NoSQLDBUtil) text2db(cmds []string) {
//	if len(cmds) != 2 {
//		fmt.Println("db text2db src_text dst_db")
//	}
//	src_text := cmds[0]
//	dst_db := cmds[1]
//	//db := o.mainFrame.GetNoSqlDB(dst_db)
//	if db == nil {
//		fmt.Println("cannot open db")
//	}
//	file, err := os.OpenFile(src_text, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
//	defer file.Close()
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	src := bufio.NewReader(file)
//	for {
//		if _line, _, err := src.ReadLine(); err == nil {
//			line := strings.TrimSpace(string(_line))
//			objs := strings.Split(line, "\t")
//			key := objs[0]
//			value := strings.Join(objs[1:], "\t")
//			db.Commit(key, value)
//		} else {
//			break
//		}
//	}
//	fmt.Println("done")
//}
//
//func (o *NoSQLDBUtil) db2text(cmds []string) {
//	if len(cmds) != 2 {
//		fmt.Println("db db2text src_db dst_text")
//	}
//	src_db := cmds[0]
//	dst_text := cmds[1]
//	db := o.mainFrame.GetNoSqlDB(src_db)
//	if db == nil {
//		fmt.Println("cannot open db")
//	}
//	var err error
//	var buf *bufio.Writer
//	var file *os.File
//	if dst_text != "screen" {
//		file, err = os.OpenFile(dst_text, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		buf = bufio.NewWriter(file)
//	} else {
//		buf = bufio.NewWriter(os.Stdout)
//	}
//	db.IterAll(func(key string, v string) (stop bool) {
//		ma, err := json.Marshal(v)
//		if err != nil {
//			fmt.Println(err)
//			return false
//		}
//		buf.WriteString(key + "\t" + string(ma) + "\n")
//		return false
//	})
//	buf.Flush()
//	if file != nil {
//		file.Close()
//	}
//	fmt.Println("done")
//}
//
//func (o *NoSQLDBUtil) do(cmds []string) {
//
//	if len(cmds) < 1 {
//		fmt.Println("Opened dbs")
//		for dbName, _ := range o.omega.OpenedDBs {
//			fmt.Println(dbName)
//		}
//		fmt.Println("db text2db/db2text src dst")
//		fmt.Println("db src delete/put key <value>")
//		return
//	}
//	if cmds[0] == "text2db" {
//		o.text2db(cmds[1:])
//	} else if cmds[0] == "db2text" {
//		o.db2text(cmds[1:])
//	}
//	targetDBName := cmds[0]
//	availables := []string{}
//	flag := false
//	for dbName, _ := range o.omega.OpenedDBs {
//		if targetDBName == dbName {
//			flag = true
//			break
//		}
//		availables = append(availables, dbName)
//	}
//	if !flag {
//		fmt.Println(availables)
//		return
//	}
//	db := o.mainFrame.GetNoSqlDB(targetDBName)
//	if len(cmds) < 3 {
//		fmt.Println("db put/delete key")
//		return
//	}
//	op := cmds[1]
//	key := cmds[2]
//	if op == "delete" {
//		db.Delete(key)
//	} else if op == "put" {
//		if len(cmds) != 4 {
//			fmt.Println("db put/delete key value")
//			return
//		}
//		db.Commit(key, cmds[3])
//	}
//}
//
func (o *NoSQLDBUtil) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	//frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
	//	MenuEntry: defines.MenuEntry{
	//		Triggers: []string{"db"},
	//		Usage:    "导出(db2text)和导入(text2db)数据库到可读的文本文件",
	//	},
	//	OptionalOnTriggerFn: func(cmds []string) (stop bool) {
	//		o.do(cmds)
	//		return true
	//	},
	//})
}

type NameRecord struct {
	*BaseCoreComponent
	Records           map[string]*collaborate.TYPE_NameEntry
	searchableByName  map[string]*collaborate.TYPE_PossibleNames
	searchableEntries map[string]*collaborate.TYPE_PossibleNames
	FileName          string `json:"改名历史记录文件"`
}

func (o *NameRecord) Init(cfg *defines.ComponentConfig) {
	o.BaseCoreComponent.Init(cfg)
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	o.searchableEntries = make(map[string]*collaborate.TYPE_PossibleNames)
	o.searchableByName = make(map[string]*collaborate.TYPE_PossibleNames)
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
		o.Records[uuid] = &collaborate.TYPE_NameEntry{
			CurrentName:    name,
			LastUpdateTime: newTime,
			NameRecord: []string{
				updateString,
			},
		}
	}
	e := o.Records[uuid]
	pn := &collaborate.TYPE_PossibleNames{Entry: e}
	pn.GenSearchAbleString()
	o.searchableEntries[uuid] = pn
	o.searchableByName[name] = pn
}

func (o *NameRecord) Stop() error {
	fmt.Println("正在保存 " + o.FileName)
	return o.mainFrame.WriteJsonData(o.FileName, o.Records)
}

func (o *NameRecord) GetPossibleName(name string, maxC int) (names []*collaborate.TYPE_PossibleNames) {
	names = make([]*collaborate.TYPE_PossibleNames, 0, maxC)
	var exactName *collaborate.TYPE_PossibleNames
	exactName = nil
	if entry, hasK := o.searchableByName[name]; hasK {
		exactName = entry
		names = append(names, entry)
	}
	//fmt.Println("exactly match ", names)
	for _, p := range o.searchableEntries {
		//fmt.Println(p.SearchableString)
		if exactName != nil && exactName.Entry.CurrentName == p.Entry.CurrentName {
			continue
		}
		if strings.Contains(p.SearchableString, name) {
			names = append(names, p)
			if len(names) == maxC {
				return
			}
		}
	}
	return
}

func (o *NameRecord) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	o.Records = map[string]*collaborate.TYPE_NameEntry{}
	err := frame.GetJsonData(o.FileName, &o.Records)
	if err != nil {
		panic(err)
	}
	for k, e := range o.Records {
		pn := &collaborate.TYPE_PossibleNames{Entry: e}
		pn.GenSearchAbleString()
		o.searchableEntries[k] = pn
		o.searchableByName[pn.Entry.CurrentName] = pn
	}
	frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		name, ud := entry.Username, entry.UUID
		name = utils.ToPlainName(name)
		o.update(name, ud.String())
	})
	var fn collaborate.FUNC_GetPossibleName
	fn = o.GetPossibleName
	(*o.mainFrame.GetContext())[collaborate.INTERFACE_POSSIBLE_NAME] = fn
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
		o.replay = true
		// if len(chat.Msg) > 0 && chat.Msg[0] == "alive" {

		// }
		if chat.Name == o.mainFrame.GetUQHolder().GetBotName() {
			// fmt.Println("popOut")
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
					"可以使用启动器保持自动重连，或者： \n" +
					"可以配合如下启动指令使 Omega 系统自动循环重启 (注意，延迟不得小于30秒，否则可能被封号)\n" +
					"对于windows系统： \n" +
					"for /l %i in (0,0,1) do @fastbuilder-windows.exe --omega_system -c 租赁服号 & @TIMEOUT /T 30 /NOBREAK\n" +
					"对于其他系统：\n" +
					"while true; ./fastbuilder -c 租赁服号 --omega_system; sleep 30; done\n",
				)
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
