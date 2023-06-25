package mainframe

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	lf "phoenixbuilder/omega/mainframe/lang_support/lua_frame"
	omgApi "phoenixbuilder/omega/mainframe/lang_support/lua_frame/omgcomponentapi"
	"phoenixbuilder/omega/utils"

	// "runtime/pprof"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type BaseCoreComponent struct {
	cfg       *defines.ComponentConfig
	omega     *Omega
	mainFrame defines.MainFrame
}

func (c *BaseCoreComponent) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	c.cfg = cfg
}

func (c *BaseCoreComponent) Signal(signal int) error {
	return nil
}

func (bc *BaseCoreComponent) BeforeActivate() error {
	return nil
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
	WriteOnlyTrigger string `json:"发送WriteOnly指令的前缀"`
}

func (c *CmdSender) send(cmds []string, typ string) {
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
	if typ == "WS" {
		c.mainFrame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, onFeedBack)
	} else if typ == "Player" {
		c.mainFrame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(cmd, onFeedBack)
	} else {
		c.mainFrame.GetGameControl().SendWOCmd(cmd)
	}
}

func (c *CmdSender) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	if cfg.Version == "0.0.1" {
		cfg.Version = "0.0.2"
		cfg.Configs["发送WriteOnly指令的前缀"] = "#"
		cfg.Upgrade()
	}
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
			c.send(cmds, "WS")
			return true
		},
	})
	frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers: []string{c.PlayerTrigger},
			Usage:    fmt.Sprintf("以玩家身份发送指令，临时打开命令返回以显示结果， 例如 %vlist", c.PlayerTrigger),
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			c.send(cmds, "Player")
			return true
		},
	})
	frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers: []string{c.WriteOnlyTrigger},
			Usage:    fmt.Sprintf("发送 WriteOnly 指令，不会返回结果， 例如 %vsay Hello", c.WriteOnlyTrigger),
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			c.send(cmds, "WO")
			return true
		},
	})
}

type NoSQLDBUtil struct {
	*BaseCoreComponent
}

//	func (o *NoSQLDBUtil) text2db(cmds []string) {
//		if len(cmds) != 2 {
//			fmt.Println("db text2db src_text dst_db")
//		}
//		src_text := cmds[0]
//		dst_db := cmds[1]
//		//db := o.mainFrame.GetNoSqlDB(dst_db)
//		if db == nil {
//			fmt.Println("cannot open db")
//		}
//		file, err := os.OpenFile(src_text, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
//		defer file.Close()
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		src := bufio.NewReader(file)
//		for {
//			if _line, _, err := src.ReadLine(); err == nil {
//				line := strings.TrimSpace(string(_line))
//				objs := strings.Split(line, "\t")
//				key := objs[0]
//				value := strings.Join(objs[1:], "\t")
//				db.Commit(key, value)
//			} else {
//				break
//			}
//		}
//		fmt.Println("done")
//	}
//
//	func (o *NoSQLDBUtil) db2text(cmds []string) {
//		if len(cmds) != 2 {
//			fmt.Println("db db2text src_db dst_text")
//		}
//		src_db := cmds[0]
//		dst_text := cmds[1]
//		db := o.mainFrame.GetNoSqlDB(src_db)
//		if db == nil {
//			fmt.Println("cannot open db")
//		}
//		var err error
//		var buf *bufio.Writer
//		var file *os.File
//		if dst_text != "screen" {
//			file, err = os.OpenFile(dst_text, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			buf = bufio.NewWriter(file)
//		} else {
//			buf = bufio.NewWriter(os.Stdout)
//		}
//		db.IterAll(func(key string, v string) (stop bool) {
//			ma, err := json.Marshal(v)
//			if err != nil {
//				fmt.Println(err)
//				return false
//			}
//			buf.WriteString(key + "\t" + string(ma) + "\n")
//			return false
//		})
//		buf.Flush()
//		if file != nil {
//			file.Close()
//		}
//		fmt.Println("done")
//	}
//
// func (o *NoSQLDBUtil) do(cmds []string) {
//
//		if len(cmds) < 1 {
//			fmt.Println("Opened dbs")
//			for dbName, _ := range o.omega.OpenedDBs {
//				fmt.Println(dbName)
//			}
//			fmt.Println("db text2db/db2text src dst")
//			fmt.Println("db src delete/put key <value>")
//			return
//		}
//		if cmds[0] == "text2db" {
//			o.text2db(cmds[1:])
//		} else if cmds[0] == "db2text" {
//			o.db2text(cmds[1:])
//		}
//		targetDBName := cmds[0]
//		availables := []string{}
//		flag := false
//		for dbName, _ := range o.omega.OpenedDBs {
//			if targetDBName == dbName {
//				flag = true
//				break
//			}
//			availables = append(availables, dbName)
//		}
//		if !flag {
//			fmt.Println(availables)
//			return
//		}
//		db := o.mainFrame.GetNoSqlDB(targetDBName)
//		if len(cmds) < 3 {
//			fmt.Println("db put/delete key")
//			return
//		}
//		op := cmds[1]
//		key := cmds[2]
//		if op == "delete" {
//			db.Delete(key)
//		} else if op == "put" {
//			if len(cmds) != 4 {
//				fmt.Println("db put/delete key value")
//				return
//			}
//			db.Commit(key, cmds[3])
//		}
//	}
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

type PerformaceAnalysis struct {
	*BaseCoreComponent
	startSuccess bool
	pprofFp      *os.File
}

func (o *PerformaceAnalysis) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	// pprofFp, err := os.OpenFile(o.omega.GetPath("cpu.prof"), os.O_RDWR|os.O_CREATE, 0644)
	// o.pprofFp = pprofFp
	// if err == nil {
	// 	if err := pprof.StartCPUProfile(pprofFp); err == nil {
	// 		o.startSuccess = true
	// 	}
	// }
}

func (o *PerformaceAnalysis) Stop() error {
	// if o.startSuccess {
	// 	fmt.Println("正在保存性能分析文件")
	// 	pprof.StopCPUProfile()
	// 	o.pprofFp.Close()
	// }
	return nil
}

type OPCheck struct {
	*BaseCoreComponent
	checkDone bool
}

func (o *OPCheck) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	botUniqueID := frame.GetUQHolder().BotUniqueID
	o.mainFrame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAdventureSettings, func(p packet.Packet) {
		if o.checkDone {
			return
		}
		pk := p.(*packet.AdventureSettings)
		if botUniqueID == pk.PlayerUniqueID {
			if pk.PermissionLevel != packet.PermissionLevelOperator {
				pterm.Warning.Println("警告：机器人不具备 OP 权限")
			}
		}
	})
}

type NameRecord struct {
	*BaseCoreComponent
	Records           map[string]*collaborate.TYPE_NameEntry
	searchableByName  map[string]*collaborate.TYPE_PossibleNames
	searchableEntries map[string]*collaborate.TYPE_PossibleNames
	FileName          string `json:"改名历史记录文件"`
}

func (o *NameRecord) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	o.BaseCoreComponent.Init(cfg, storage)
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
			if maxC != 0 && len(names) == maxC {
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
	var fn collaborate.FUNCTYPE_GET_POSSIBLE_NAME = o.GetPossibleName
	o.mainFrame.SetContext(collaborate.INTERFACE_POSSIBLE_NAME, fn)
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
	Schedule   int `json:"检测周期"`
	Delay      int `json:"最大延迟"`
	replay     bool
	lastPacket packet.Packet
	lastTime   time.Time
}

func (o *KeepAlive) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	o.BaseCoreComponent.Init(cfg, storage)
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *KeepAlive) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	o.mainFrame.GetGameListener().SetOnAnyPacketCallBack(func(p packet.Packet) {
		o.replay = true
		o.lastPacket = p
		o.lastTime = time.Now()
	})
	o.mainFrame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     []string{"packet_analysis"},
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "显示最近发包情况",
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			pterm.Info.Println(o.omega.Reactor.analyzer.PrintAnalysis())
			pterm.Info.Println(o.omega.GameCtrl.analyzer.PrintAnalysis())
			fname := path.Join(o.omega.GetStorageRoot(), "最后发送的指令记录.txt")
			if fp, err := os.OpenFile(fname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755); err == nil {
				fp.WriteString(o.omega.GameCtrl.analyzer.GenSendedCmdList())
				o.mainFrame.GetBackendDisplay().Write("最后发送的指令记录保存在 " + fname)
			}
			return true
		},
	})
}

func (o *KeepAlive) Activate() {
	o.BaseCoreComponent.Activate()
	t := time.NewTicker(time.Second * time.Duration(o.Schedule))
	go func() {
		for {
			<-t.C
			if o.replay {
				o.replay = false
				continue
			} else {
				o.mainFrame.GetGameControl().SendCmdAndInvokeOnResponse("list", func(output *packet.CommandOutput) {
					o.replay = true
				})
				<-time.NewTimer(time.Second * time.Duration(o.Delay)).C
				if !o.replay {
					o.mainFrame.GetBackendDisplay().Write("连接检查失败，疑似Omega假死，准备退出...")
					if o.lastPacket == nil {
						o.mainFrame.GetBackendDisplay().Write("自启动以来没有收到任何数据包")
					} else {
						o.mainFrame.GetBackendDisplay().Write(fmt.Sprintf("最后收到的数据包距现在 %v, 类型 %v 数据 %v", time.Since(o.lastTime), utils.PktIDInvMapping[int(o.lastPacket.ID())], o.lastPacket))
					}
					o.mainFrame.GetBackendDisplay().Write("以下为最后发送数据包的统计信息:")
					o.mainFrame.GetBackendDisplay().Write(o.omega.GameCtrl.analyzer.PrintAnalysis())
					fname := path.Join(o.omega.GetStorageRoot(), "最后发送的指令记录.txt")
					if fp, err := os.OpenFile(fname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755); err == nil {
						fp.WriteString(o.omega.GameCtrl.analyzer.GenSendedCmdList())
						o.mainFrame.GetBackendDisplay().Write("最后发送的指令记录保存在 " + fname)
					}
					o.omega.Stop()
					fmt.Println("3秒后退出")
					<-time.NewTimer(time.Second * 3).C
					// restartHint := "while true; do ./fastbuilder -c 租赁服号 --omega_system; sleep 30; done\n"
					// if runtime.GOOS == "windows" {
					// 	restartHint = "for /l %i in (0,0,1) do @fastbuilder-windows.exe --omega_system -c 租赁服号 & @TIMEOUT /T 30 /NOBREAK\n"
					// }
					panic("Omega 假死，已退出...\n" +
						"建议使用 Omega 启动器以实现自动断线重连\n")
				}
			}
		}
	}()
}

type Partol struct {
	*BaseCoreComponent
	EnableInvisibility     bool      `json:"启用隐身"`
	EnablePartol           bool      `json:"启用随机巡逻"`
	Patrol                 int       `json:"随机巡逻间隔(秒)"`
	TeleportWhenPlayerJoin bool      `json:"是否在玩家上线时传送至其位置"`
	AlwaysInOverworld      bool      `json:"是否将机器人固定在主世界维度"`
	OverworldAnchor        []float32 `json:"主世界锚点"`
}

func (o *Partol) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	o.BaseCoreComponent.Init(cfg, storage)
	if cfg.Version == "0.0.1" {
		cfg.Version = "0.0.2"
		cfg.Configs["启用随机巡逻"] = true
		cfg.Configs["主世界锚点"] = []float32{0, 320, 0}
		cfg.Upgrade()
	}
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	if o.Patrol < 90 {
		panic("巡逻时间太短，至少应该设为 90")
	}
}

func (o *Partol) Inject(frame defines.MainFrame) {
	o.mainFrame = frame
	if o.TeleportWhenPlayerJoin {
		o.mainFrame.GetGameListener().SetOnTypedPacketCallBack(packet.IDText, func(p packet.Packet) {
			pk := p.(*packet.Text)
			if pk.TextType == 2 && pk.Message == "§e%multiplayer.player.joined" {
				o.mainFrame.GetGameControl().SendWOCmd(fmt.Sprintf("execute \"%s\" ~ 320 ~ tp \"%s\" ~ ~ ~", pk.Parameters[0], o.mainFrame.GetUQHolder().GetBotName()))
			}
		})
	}
	if o.AlwaysInOverworld {
		o.mainFrame.GetGameListener().SetOnTypedPacketCallBack(packet.IDChangeDimension, func(p packet.Packet) {
			pk := p.(*packet.ChangeDimension)
			if pk.Dimension != 0 {
				go func() {
					// 延迟3秒是为了给予机器人接收数据包的时间
					<-time.NewTimer(time.Second * time.Duration(3)).C
					cmd := fmt.Sprintf("tp \"%s\" %f %f %f", o.mainFrame.GetUQHolder().GetBotName(), o.OverworldAnchor[0], o.OverworldAnchor[1], o.OverworldAnchor[2])
					o.mainFrame.GetGameControl().SendWOCmd(cmd)
				}()
			}
		})
	}
}

func (o *Partol) Activate() {
	//o.mainFrame.GetGameControl().SendCmd("/say test=======")
	if !o.EnablePartol {
		return
	}
	count := 0
	for {
		count++
		o.mainFrame.GetBotTaskScheduler().CommitBackgroundTask(&defines.BasicBotTaskPauseAble{
			BasicBotTask: defines.BasicBotTask{
				Name: fmt.Sprintf("Portal %v", count),
				ActivateFn: func() {
					utils.GetPlayerList(o.mainFrame.GetGameControl(), "@r[rm=100]", func(players []string) {
						if o.EnableInvisibility {
							o.mainFrame.GetGameControl().SendCmd(fmt.Sprintf("effect @s invisibility %v 1 true", o.Patrol*2))
						}
						if len(players) > 0 {
							o.mainFrame.GetGameControl().SendWOCmd(fmt.Sprintf("execute \"%s\" ~ 320 ~ tp \"%s\" ~ ~ ~", players[0], o.mainFrame.GetUQHolder().GetBotName()))
						}
					})
				},
			},
		})
		<-time.NewTimer(time.Second * time.Duration(o.Patrol)).C
	}
}

// 插件
type LuaComponenter struct {
	*BaseCoreComponent
	Monitor   *lf.Monitor
	LuaFrame  *lf.BuiltlnFn
	mainFrame defines.MainFrame
}

func (b *LuaComponenter) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	//读取lua框架
	b.Monitor = lf.NewMonitor(omgApi.NewOmgCoreComponent(b.omega, b.mainFrame))
	//读取一次已经产生的文件
	b.Monitor.InintComponents()
	i := 0
	for k, v := range b.Monitor.LuaComponentData {
		i++
		if v.JsonConfig.Disabled {

			b.Monitor.OmgFrame.Omega.GetBackendDisplay().Write(pterm.Warning.Sprintf("\t跳过加载组件 %3d/%3d [%v] %v@%v", i, len(b.Monitor.LuaComponentData), v.JsonConfig.Source, k, v.JsonConfig.Version))
			//b.omega.backendLogger.Write()
		} else {
			b.Monitor.OmgFrame.Omega.GetBackendDisplay().Write(pterm.Success.Sprintf("\t正在加载组件 %3d/%3d [%v] %v@%v", i, len(b.Monitor.LuaComponentData), v.JsonConfig.Source, k, v.JsonConfig.Version))
			//b.omega.backendLogger.Write()
		}

	}
}
func (b *LuaComponenter) Inject(frame defines.MainFrame) {
	b.mainFrame = frame
	//注入函数 并且开启插件

	b.Monitor.InjectComponents()

}

func (o *LuaComponenter) Activate() {
	time.Sleep(time.Second * 3)
	//开启组件
	o.Monitor.OmgFrame.MainFrame = o.mainFrame
	for k, _ := range o.Monitor.ComponentPoll {
		err := o.Monitor.StartComponent(k, o.Monitor.LuaComponentData[k].LuaFile)
		if err != nil {
			lf.PrintInfo(lf.NewPrintMsg("警告", err))
		}
	}
	//现在开始监听后台
	func() {
		o.Monitor.OmgFrame.Omega.SetBackendCmdInterceptor(func(cmds []string) (stop bool) {
			is := false
			if cmds[0] == "lua" {
				is = true
			}
			cmd := ""
			for _, v := range cmds {
				cmd += v + " "
			}
			if err := o.Monitor.CmdCenter(cmd); err != nil {
				lf.PrintInfo(lf.NewPrintMsg("警告", err))
			}
			return is
		})
		//o.omega.SetBackendCmdInterceptor()
		o.mainFrame.GetGameListener().SetGameChatInterceptor(o.MsgDistributionCenter)
	}()

}

// 每次消息传输过来则分发处理
func (b *LuaComponenter) MsgDistributionCenter(chat *defines.GameChat) bool {
	b.Monitor.BuiltlnFner.Listener.Range(func(key, value interface{}) bool { // 遍历所有监听器
		msg := ""
		for _, v := range chat.Msg {
			msg += v + " "
		}
		message := lf.Message{
			Type:    chat.Name,
			Content: msg,
		}
		listener := key.(*lf.Listener) // 获取监听器实例
		select {
		case listener.MsgChannel <- message: // 尝试将消息发送到监听器的消息通道
		default: // 如果监听器的消息通道已满
			<-listener.MsgChannel          // 从通道中读取并丢弃一条最旧的消息
			listener.MsgChannel <- message // 将新消息发送到监听器的消息通道
		}
		return true
	})
	return false
}

func getCoreComponentsPool() map[string]func() defines.CoreComponent {
	return map[string]func() defines.CoreComponent{
		"菜单显示":      func() defines.CoreComponent { return &Menu{BaseCoreComponent: &BaseCoreComponent{}} },
		"指令发送":      func() defines.CoreComponent { return &CmdSender{BaseCoreComponent: &BaseCoreComponent{}} },
		"数据库导入导出工具": func() defines.CoreComponent { return &NoSQLDBUtil{&BaseCoreComponent{}} },
		"改名记录":      func() defines.CoreComponent { return &NameRecord{BaseCoreComponent: &BaseCoreComponent{}} },
		"假死检测":      func() defines.CoreComponent { return &KeepAlive{BaseCoreComponent: &BaseCoreComponent{}} },
		"性能分析":      func() defines.CoreComponent { return &PerformaceAnalysis{BaseCoreComponent: &BaseCoreComponent{}} },
		"OP权限自检":    func() defines.CoreComponent { return &OPCheck{BaseCoreComponent: &BaseCoreComponent{}} },
		"随机巡逻":      func() defines.CoreComponent { return &Partol{BaseCoreComponent: &BaseCoreComponent{}} },
		"lua插件支持":   func() defines.CoreComponent { return &LuaComponenter{BaseCoreComponent: &BaseCoreComponent{}} },
	}
}
