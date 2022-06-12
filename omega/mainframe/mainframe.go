package mainframe

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"runtime"
	"strings"
	"time"

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
	ctx      *map[string]interface{}

	backendLogger    defines.LineDst
	redAlertLogger   defines.LineDst
	redAlertHandlers []func(info string)
	ComponentConfigs []*defines.ComponentConfig
	OmegaConfig      *defines.OmegaConfig

	BackendMenuEntries  []*defines.BackendMenuEntry
	BackendInterceptors []func(cmds []string) (stop bool)

	//OpenedDBs map[string]*utils.LevelDBWrapper

	GameCtrl *GameCtrl
	Reactor  *Reactor

	Components              []defines.Component
	configStageCompleteFlag bool
}

func NewOmega() *Omega {
	o := &Omega{
		pktsChan: make(chan packet.Packet, 1024),
		CloseFns: make([]func() error, 0),
		// ctx:                 make(map[string]interface{}),
		BackendMenuEntries:  make([]*defines.BackendMenuEntry, 0),
		BackendInterceptors: make([]func(cmds []string) (stop bool), 0),
		redAlertHandlers:    make([]func(info string), 0),
		//OpenedDBs:           make(map[string]*utils.LevelDBWrapper),
		stopC:        make(chan struct{}),
		fullyStopped: make(chan struct{}),
	}
	o.Reactor = newReactor(o)
	o.ctx = &map[string]interface{}{}
	return o
}

func (o *Omega) GetContext() *map[string]interface{} {
	return o.ctx
}

func (o *Omega) GetUQHolder() *uqHolder.UQHolder {
	return o.uqHolder
}

func (o *Omega) GetWorldsDir() string {
	return path.Join(o.storageRoot, "worlds")
}

func (o *Omega) GetAllConfigs() []*defines.ComponentConfig {
	return o.ComponentConfigs
}
func (o *Omega) GetOmegaConfig() *defines.OmegaConfig {
	return o.OmegaConfig
}
func (o *Omega) GetPath(elem ...string) string {
	for _, ele := range elem {
		if strings.HasPrefix(ele, "/") || strings.Contains(ele, "..") {
			panic(fmt.Errorf("为了安全考虑，路径开头不能为 / 且不能包含 .."))
		}
	}
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
	return utils.GetFileData(o.GetRelativeFileName(topic))
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
	return utils.WriteJsonData(fname, data)
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

//func (o *Omega) GetNoSqlDB(topic string) defines.NoSqlDB {
//	if db, hasK := o.OpenedDBs[topic]; hasK {
//		return db
//	}
//	db := utils.GetLevelDB(o.GetPath("noSQL", topic))
//	o.OpenedDBs[topic] = db
//	o.CloseFns = append(o.CloseFns, db.Close)
//	return db
//}

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

func GetMemUsageByMB() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return (m.HeapIdle - m.HeapReleased + m.StackSys) / 1024 / 1024
}

func GetMemUsageByMBInDetailedString() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	toMB := func(v uint64) float32 {
		return float32(v) / 1024 / 1024
	}
	return fmt.Sprintf("系统分配[包括备用]内存 %.1f MB, ([空闲堆]%.1fMB / [释放堆]%.1fMB / [堆]%.1fMB / [分配栈]%.1fMB)", toMB(m.Sys), toMB(m.HeapIdle), toMB(m.HeapReleased), toMB(m.HeapInuse), toMB(m.StackSys))
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
				// o.backendLogger.Write(pterm.Warning.Sprintln("Game Packet Feeder & Reactor & UQHoder 已退出"))
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
	go func() {
		if o.OmegaConfig.ShowMemUsagePeriod == 0 && o.OmegaConfig.MemLimit == 0 {
			return
		}
		usage := GetMemUsageByMB()
		if o.OmegaConfig.ShowMemUsagePeriod != 0 {
			go func() {
				for {
					pterm.Info.Printfln("[内存] %v", GetMemUsageByMBInDetailedString())
					<-time.NewTimer(time.Duration(o.OmegaConfig.ShowMemUsagePeriod) * time.Second).C
				}
			}()
		}
		for {
			usage = GetMemUsageByMB()
			if usage > uint64(o.OmegaConfig.MemLimit) {
				hint := fmt.Sprintf("使用内存 %v MB 超出安全上限 %v MB, 为保证数据安全，Omega 将立刻保存数据并重启以释放内存(您可以在 配置/主系统中调整)", usage, o.OmegaConfig.MemLimit)
				pterm.Warning.Println(hint)
				o.Stop()
				panic(hint)
			}
			<-time.NewTimer(3 * time.Second).C
		}
	}()
	for {
		backendInputChan := o.adaptor.GetBackendCommandFeeder()
		select {
		case cmd := <-backendInputChan:
			if cmd == "stop" {
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
			// o.backendLogger.Write(pterm.Warning.Sprintln("后台指令分派器已退出"))
			return
		}
	}
}
