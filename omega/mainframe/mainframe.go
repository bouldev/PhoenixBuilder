package mainframe

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/components"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"

	"github.com/pterm/pterm"
)

type Omega struct {
	adaptor  defines.ConnectionAdaptor
	pktsChan chan packet.Packet

	CloseFns     []func() error
	stopC        chan struct{}
	fullyStopped chan struct{}
	closed       bool

	storageRoot string

	uqHolder *uqHolder.UQHolder
	ctx      map[string]interface{}

	backendLogger    defines.LineDst
	redAlertLogger   defines.LineDst
	redAlertHandlers []func(info string)
	fullConfig       *OmegaConfig

	BackendMenuEntries  []*defines.BackendMenuEntry
	BackendInterceptors []func(cmds []string) (stop bool)

	OpenedDBs map[string]*utils.LevelDBWrapper

	GameCtrl *GameCtrl
	Reactor  *Reactor

	Components              []defines.Component
	configStageCompleteFlag bool
}

func NewOmega() *Omega {
	o := &Omega{
		pktsChan:            make(chan packet.Packet, 1024),
		CloseFns:            make([]func() error, 0),
		ctx:                 make(map[string]interface{}),
		BackendMenuEntries:  make([]*defines.BackendMenuEntry, 0),
		BackendInterceptors: make([]func(cmds []string) (stop bool), 0),
		redAlertHandlers:    make([]func(info string), 0),
		OpenedDBs:           make(map[string]*utils.LevelDBWrapper),
		stopC:               make(chan struct{}),
		fullyStopped:        make(chan struct{}),
	}
	o.Reactor = newReactor(o)
	return o
}

func (o *Omega) GetContext() map[string]interface{} {
	return o.ctx
}

func (o *Omega) GetUQHolder() *uqHolder.UQHolder {
	return o.uqHolder
}

func (o *Omega) SetRoot(root string) {
	// config stage
	o.storageRoot = root
}

func (o *Omega) postProcess() {
	if !utils.IsDir(o.storageRoot) {
		fmt.Println("创建数据文件夹 " + o.storageRoot)
		if err := utils.MakeDirP(o.storageRoot); err != nil {
			panic(err)
		}
	}
	o.readConfig()
	dataDir := o.GetPath("data")
	if !utils.IsDir(dataDir) {
		fmt.Println("创建数据文件夹: " + dataDir)
		if err := utils.MakeDirP(dataDir); err != nil {
			panic(err)
		}
	}
	logDir := o.GetPath("logs")
	if !utils.IsDir(logDir) {
		fmt.Println("创建日志文件夹: " + logDir)
		if err := utils.MakeDirP(logDir); err != nil {
			panic(err)
		}
	}
	noSqlDir := o.GetPath("noSQL")
	if !utils.IsDir(noSqlDir) {
		fmt.Println("创建非关系型数据库文件夹: " + noSqlDir)
		if err := utils.MakeDirP(noSqlDir); err != nil {
			panic(err)
		}
	}
}

func (o *Omega) GetAllConfigs() []*defines.ComponentConfig {
	return o.fullConfig.ComponentsConfig
}

func (o *Omega) QueryConfig(topic string) interface{} {
	return o.fullConfig.QueryConfig(topic)
}

func (o *Omega) GetPath(elem ...string) string {
	return path.Join(o.storageRoot, path.Join(elem...))
}

func (o *Omega) GetRelativeFileName(topic string) string {
	return o.GetPath("data", topic)
}

func (o *Omega) GetLogger(topic string) defines.LineDst {
	logger := utils.NewFileNormalLogger(o.GetPath("logs", topic))
	o.CloseFns = append(o.CloseFns, logger.Close)
	return logger
}

func (o *Omega) GetFileData(topic string) ([]byte, error) {
	fp, err := os.OpenFile(o.GetRelativeFileName(topic), os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	return buf, err
}
func (o *Omega) WriteFileData(topic string, data []byte) error {
	fp, err := os.OpenFile(o.GetRelativeFileName(topic), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.Write(data)
	return err
}

func (o *Omega) WriteJsonData(topic string, data interface{}) error {
	fname := o.GetRelativeFileName(topic)
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	err = enc.Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func (o *Omega) GetJsonData(topic string, ptr interface{}) error {
	data, err := o.GetFileData(topic)
	if err != nil {
		return err
	}
	if data == nil || len(data) == 0 {
		return nil
	}
	err = json.Unmarshal(data, ptr)
	if err != nil {
		return err
	}
	return nil
}

func (o *Omega) GetNoSqlDB(topic string) defines.NoSqlDB {
	if db, hasK := o.OpenedDBs[topic]; hasK {
		return db
	}
	db := utils.GetLevelDB(o.GetPath("noSQL", topic))
	o.OpenedDBs[topic] = db
	o.CloseFns = append(o.CloseFns, db.Close)
	return db
}

func (o *Omega) FullyStopped() chan struct{} {
	return o.fullyStopped
}

func (o *Omega) Stop() error {
	if o.closed {
		<-o.fullyStopped
		return nil
	}
	o.closed = true
	o.backendLogger.Write("正在保存数据并关闭系统...")
	close(o.stopC)
	errS := ""
	//fmt.Println(o.CloseFns)
	for _, fn := range o.CloseFns {
		//fmt.Println(fn)
		if e := fn(); e != nil {
			errS += "\t" + e.Error() + "\n"
		}
	}
	if errS != "" {
		return fmt.Errorf("关闭系统各部件中，发生了以下错误:\n" + errS)
	}
	fmt.Println("Omega 系统已安全退出")
	close(o.fullyStopped)
	return nil
}

type BackEndLogger struct {
	loggers []defines.LineDst
}

func (bl *BackEndLogger) Write(line string) {
	for _, logger := range bl.loggers {
		logger.Write(line)
	}
}

func (o *Omega) GetBackendDisplay() defines.LineDst {
	return o.backendLogger
}

func (o *Omega) backendMenuEntryToStdInterceptor(entry *defines.BackendMenuEntry) func(cmds []string) (stop bool) {
	return func(cmds []string) (stop bool) {
		if trig, reducedCmds := utils.CanTrigger(cmds, entry.Triggers, true, false); trig {
			return entry.OptionalOnTriggerFn(reducedCmds)
		}
		return false
	}
}

func (o *Omega) SetBackendCmdInterceptor(fn func(cmds []string) (stop bool)) {
	o.BackendInterceptors = append(o.BackendInterceptors, fn)
}

func (o *Omega) SetBackendMenuEntry(entry *defines.BackendMenuEntry) {
	o.BackendMenuEntries = append(o.BackendMenuEntries, entry)
	interceptor := o.backendMenuEntryToStdInterceptor(entry)
	o.SetBackendCmdInterceptor(interceptor)
}

type FuncsToLogger struct {
	GetFns func() []func(info string)
}

func (ftl *FuncsToLogger) Write(info string) {
	for _, fn := range ftl.GetFns() {
		fn(info)
	}
}

func (o *Omega) configStageComplete() {
	o.configStageCompleteFlag = true
}

func (o *Omega) RedAlert(info string) {
	o.redAlertLogger.Write(info)
}

func (o *Omega) RegOnAlertHandler(cb func(info string)) {
	o.redAlertHandlers = append(o.redAlertHandlers, cb)
}

func (o *Omega) loadComponents() (success bool) {
	success = false
	defer func() {
		r := recover()
		if r != nil {
			success = false
			pterm.Error.Printf("组件配置文件不正确，因此 Omega 系统拒绝启动，具体错误如下:\n%v\n", r)
		}
	}()
	total := len(o.fullConfig.ComponentsConfig)
	coreComponentsLoaded := map[string]bool{}
	corePool := getCoreComponentsPool()
	builtInPool := components.GetComponentsPool()
	for n, _ := range corePool {
		coreComponentsLoaded[n] = false
	}
	for i, cfg := range o.fullConfig.ComponentsConfig {
		I := i + 1
		Name := cfg.Name
		Version := cfg.Version
		Source := cfg.Source
		if cfg.Disabled {
			o.backendLogger.Write(fmt.Sprintf("\t跳过加载组件 %3d/%3d [%v] %v@%v", I, total, Source, Name, Version))
			continue
		}
		o.backendLogger.Write(fmt.Sprintf("\t正在加载组件 %3d/%3d [%v] %v@%v", I, total, Source, Name, Version))
		var component defines.Component
		if Source == "Core" {
			if componentFn, hasK := corePool[Name]; !hasK {
				o.backendLogger.Write("没有找到核心组件: " + Name)
				panic("没有找到核心组件: " + Name)
			} else {
				coreComponentsLoaded[Name] = true
				_component := componentFn()
				_component.SetSystem(o)
				component = _component
			}
		} else if Source == "Built-In" {
			if componentFn, hasK := builtInPool[Name]; !hasK {
				o.backendLogger.Write("没有找到内置组件: " + Name)
				panic("没有找到内置组件: " + Name)
			} else {
				component = componentFn()
			}
		}
		component.Init(cfg)
		component.Inject(NewBox(o, Name))
		o.Components = append(o.Components, component)
	}
	for n, l := range coreComponentsLoaded {
		if !l {
			panic(fmt.Errorf("核心组件 (Core) 必须被加载, 但是 %v 被配置为不加载", n))
		}
	}
	return true
}

func (o *Omega) Bootstrap(adaptor defines.ConnectionAdaptor) {
	fmt.Println("开始预处理任务")
	o.postProcess()
	o.adaptor = adaptor
	o.uqHolder = adaptor.GetInitUQHolderCopy()
	o.backendLogger = &BackEndLogger{
		loggers: []defines.LineDst{
			o.GetLogger("后台信息.log"),
			utils.NewIONormalLogger(os.Stdout),
		},
	}
	o.redAlertLogger = &BackEndLogger{
		loggers: []defines.LineDst{
			o.backendLogger,
			o.GetLogger("security_event.log"),
			&FuncsToLogger{GetFns: func() []func(info string) {
				return o.redAlertHandlers
			}},
		},
	}
	o.backendLogger.Write("日志系统已可用,正在激活主框架...")
	o.backendLogger.Write("加载组件中...")
	if o.loadComponents() == false {
		o.Stop()
		return
	}
	o.backendLogger.Write("组件全部加载&配置完成, 正在将更新后的配置写回配置文件...")
	o.writeBackConfig()
	o.configStageComplete()
	o.backendLogger.Write("启用 Game Ctrl 模块")
	o.GameCtrl = newGameCtrl(o)

	o.backendLogger.Write("开始激活组件并挂载后执行任务...")
	for _, component := range o.Components {
		c := component
		o.CloseFns = append(o.CloseFns, func() error {
			return c.Stop()
		})
		go component.Activate()
	}
	//fmt.Println(o.CloseFns)
	o.backendLogger.Write("全部完成，系统启动")
	for _, p := range o.uqHolder.PlayersByEntityID {
		for _, cb := range o.Reactor.OnFirstSeePlayerCallback {
			cb(p.Username)
		}
	}
	{
		logo := GetLogo(LOGO_BOTH)
		//banner := []string{
		//	"┌───────────────────────────────────────────────────────────────────────┐",
		//	"|   ██████  ███    ███ ███████  ██████   █████      ███    ██  ██████   |",
		//	"|  ██    ██ ████  ████ ██      ██       ██   ██     ████   ██ ██        |",
		//	"|  ██    ██ ██ ████ ██ █████   ██   ███ ███████     ██ ██  ██ ██   ███  |",
		//	"|  ██    ██ ██  ██  ██ ██      ██    ██ ██   ██     ██  ██ ██ ██    ██  |",
		//	"|   ██████  ██      ██ ███████  ██████  ██   ██     ██   ████  ██████   |",
		//	"└───────────────────────────────────────────────────────────────────────┘",
		//}
		fmt.Println(strings.Join(logo, "\n"))
	}
	pterm.Success.Println("OMEGA_ng 等待指令")
	pterm.Success.Println("输入 ? 以获得帮助")
}
func (o *Omega) Activate() {
	defer func(o *Omega) {
		err := o.Stop()
		if err != nil {

		}
	}(o)
	go func() {
		for {
			pkt := o.adaptor.Read()
			if pkt == nil {
				continue
			}
			//fmt.Println(pkt)
			if o.closed {
				o.backendLogger.Write(pterm.Warning.Sprintln("Game Packet Feeder & Reactor & UQHoder 已退出"))
				return
			}
			uqHolderDelayUpdate := false
			if pkt.ID() == packet.IDPlayerList {
				pk := pkt.(*packet.PlayerList)
				if pk.ActionType == packet.PlayerListActionRemove {
					uqHolderDelayUpdate = true
				}
			}
			if !uqHolderDelayUpdate {
				o.uqHolder.Update(pkt)
				o.Reactor.React(pkt)
			} else {
				o.Reactor.React(pkt)
				o.uqHolder.Update(pkt)
			}
		}
	}()
	for {
		backendInputChan := o.adaptor.GetBackendCommandFeeder()
		select {
		case cmd := <-backendInputChan:
			if cmd == "exit" {
				o.Stop()
				return
			}
			cmds := utils.GetStringContents(cmd)
			catched := false
			for _, interceptor := range o.BackendInterceptors {
				stop := interceptor(cmds)
				if stop {
					catched = true
					break
				}
			}
			if catched {
				continue
			}
			o.backendLogger.Write(pterm.Warning.Sprintf("没有组件可以处理该指令: %v (%v), 输入?获得帮助", cmd, cmds))
			o.backendLogger.Write(pterm.Warning.Sprintf("尝试调用 FB 指令"))
			go func() {
				o.adaptor.FBEval(cmd)
			}()
		case <-o.stopC:
			o.backendLogger.Write(pterm.Warning.Sprintln("后台指令分派器已退出"))
			return
		}
	}
}
