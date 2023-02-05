package universe_export

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/assembler"
	"phoenixbuilder/mirror/io/mcdb"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"phoenixbuilder/omega/utils/structure"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/pterm/pterm"
)

var paramsNames []string = []string{"[建筑名]", "[起点x]", "[起点y]", "[起点z]", "[终点x]", "[终点y]", "[终点z]"}

type exportTask struct {
	StructureName string
	StartPos      define.CubePos
	EndPos        define.CubePos
}

type ExportData struct {
	CurrentTask *exportTask   `json:"当前正在处理的任务"`
	QueuedTasks []*exportTask `json:"排队中的任务"`
}

type UniverseExport struct {
	*defines.BasicComponent
	Triggers           []string `json:"后台触发词"`
	FileName           string   `json:"断点续导记录文件"`
	AutoContinueImport bool     `json:"Omega启动时是否自动继续导出"`
	ContinueTriggers   []string `json:"继续导出的触发词"`
	CancelTriggers     []string `json:"取消导出的触发词"`
	fileChange         bool
	needDecision       bool
	currentExporter    *Exporter
	data               *ExportData
}

type Exporter struct {
	callStop          bool
	task              *exportTask
	doneWaiter        chan struct{}
	frame             defines.MainFrame
	mu                sync.Mutex
	listening         bool
	allRequiredChunks *structure.ExportedChunksMap
	chunks            map[define.ChunkPos]*mirror.ChunkData
	updateHit         func(pos define.ChunkPos) bool
	feedChan          chan bool
	provider          *mcdb.Provider
}

func (o *Exporter) cancel() {
	o.callStop = true
}

type TeleportFn func(x, z int)

func (o *Exporter) Activate() {
	defer func() {
		close(o.doneWaiter)
		o.frame.GetGameListener().GetChunkAssembler().AdjustSendPeriod(assembler.REQUEST_NORMAL)
	}()
	o.frame.GetGameListener().GetChunkAssembler().AdjustSendPeriod(assembler.REQUEST_AGGRESSIVE)
	startPos := o.task.StartPos
	endPos := o.task.EndPos
	structureName := o.task.StructureName
	hopPath, allRequiredChunks := structure.PlanHopSwapPath(startPos.X(), startPos.Z(), endPos.X(), endPos.Z(), 11)
	o.allRequiredChunks = allRequiredChunks
	o.chunks = make(map[define.ChunkPos]*mirror.ChunkData)

	progressBar, _ := pterm.DefaultProgressbar.WithTotal(len(*allRequiredChunks)).WithTitle(structureName).Start()
	startTime := time.Now()
	hitCount := 0
	o.updateHit = func(pos define.ChunkPos) bool {
		isHit := allRequiredChunks.Hit(pos)
		if isHit {
			hitCount += 1
			metricDuration := time.Since(startTime).Seconds()
			realSpeed := float64(hitCount) / metricDuration
			progressBar.Title = structureName + fmt.Sprintf(" 实际速度 %.2f", realSpeed)
			progressBar.Increment()
			return true
		} else {
			return false
		}
	}

	overallCacheDir := o.frame.GetOmegaCacheDir("structure_export", structureName)
	mcWorldDir := path.Join(overallCacheDir, structureName)
	_provider, err := mcdb.New(mcWorldDir, opt.FlateCompression)
	if err != nil {
		pterm.Error.Println(err)
	}
	o.provider = _provider
	o.provider.D.LevelName = structureName

	for pos, ct := range *allRequiredChunks {
		// fmt.Println("WANT 2", pos)
		if ct.CachedMark {
			// fmt.Println("Cached")
			continue
		}
		cd := o.provider.GetWithNoFallBack(pos)
		if cd != nil {
			// fmt.Println("HIT")
			o.chunks[pos] = cd
			o.updateHit(pos)
		} else {
			// fmt.Println("MISS")
		}
	}

	for pos, ct := range *o.allRequiredChunks {
		// fmt.Println("WANT 1", pos)
		if ct.CachedMark {
			cd := o.frame.GetWorldProvider().GetWithNoFallBack(pos)
			if cd != nil {
				o.provider.Write(cd)
				o.chunks[pos] = cd
			}
			// fmt.Println("Cached")
			continue
		}
		cd := o.frame.GetWorldProvider().GetWithNoFallBack(pos)
		if cd != nil {
			// fmt.Println("HIT")
			o.chunks[pos] = cd
			o.updateHit(pos)
			o.provider.Write(cd)
		} else {
			// fmt.Println("MISS")
		}
	}

	PrintSuccessAndWriteFile := func() {
		pterm.Success.Printfln("已成功获取 %v 需要的所有区块 %v", structureName, overallCacheDir)
		fp, err := os.OpenFile(path.Join(mcWorldDir, "Omega导出建筑记录.txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
		if err == nil {
			pterm.Success.Printfln("基岩版存档生成成功，建筑信息位于 Omega导出建筑记录.txt")
			fmt.Fprintf(fp, "@STRUCTURE: %v @START: %v %v %v @END: %v %v %v @TIME: %v\n", structureName, startPos.X(), startPos.Y(), startPos.Z(), endPos.X(), endPos.Y(), endPos.Z(), utils.TimeToString(time.Now()))
			fp.Close()
		} else {
			pterm.Error.Printfln("基岩版存档生成失败")
		}
		o.provider.Close()
		structureNameWithPos := fmt.Sprintf("%v@%v,%v,%v.mcworld", structureName, startPos.X(), startPos.Y(), startPos.Z())
		fp, err = os.OpenFile(path.Join(overallCacheDir, structureNameWithPos), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err == nil {
			err = utils.Zip(mcWorldDir, fp, func(filePath string, info os.FileInfo) (discard bool) { return false })
			if err == nil {
				pterm.Success.Printfln("mcworld 文件生成成功")
			} else {
				pterm.Error.Printfln("mcworld 文件生成失败 %v", err)
			}
			fp.Close()
		} else {
			pterm.Error.Printfln("mcworld 文件生成失败 %v", err)
		}
		err = structure.EncodeSchem(o.chunks, startPos, endPos, structureName, overallCacheDir)
		if err != nil {
			pterm.Error.Printfln("schem 文件生成失败 %v", err)
		} else {
			pterm.Success.Printfln("schem 文件生成成功")
		}
		targetDir := path.Join(o.frame.GetStorageRoot(), "Omega导出", structureName)
		if utils.IsDir(targetDir) {
			tmpDir := targetDir
			i := 1
			for utils.IsDir(tmpDir) {
				tmpDir = targetDir + fmt.Sprintf("(%v)", i)
				i++
			}
			targetDir = tmpDir
		}
		os.MkdirAll(path.Join(o.frame.GetStorageRoot(), "Omega导出"), 0755)
		if err := utils.CopyDirectory(overallCacheDir, targetDir); err == nil {
			pterm.Success.Printfln("导出已经成功，文件位于 %v", targetDir)
			os.RemoveAll(overallCacheDir)
		} else {
			pterm.Success.Printfln("导出已经成功，文件位于 %v (%v)", overallCacheDir, err)
		}
	}

	if len(*hopPath) == 0 {
		PrintSuccessAndWriteFile()
		return
	}

	if len(*hopPath) == 0 {
		PrintSuccessAndWriteFile()
		return
	}

	o.listening = true
	o.feedChan = make(chan bool, 10240)
	for i := 1; i <= 3; i++ {
		for _, hp := range *hopPath {
			pterm.Warning.Println("now hop to: ", hp.Pos)
			o.doHop(func(x, z int) {
				o.frame.GetGameControl().SendCmd(fmt.Sprintf("tp @s %v 320 %v", x, z))
			}, hp, 10*float32(i), 0.5*float32(i), 5*float32(i))
		}
		if len(*hopPath) == 0 || i == 3 {
			break
		}
		pterm.Error.Printfln("区块没有完全获取成功，重试 %v/%v", i, 2)
	}

	if len(*hopPath) == 0 {
		PrintSuccessAndWriteFile()
		return
	} else {
		for pos, m := range *allRequiredChunks {
			if !m.CachedMark {
				pterm.Error.Printfln("未能获得区块 %v", pos)
			}
		}
		pterm.Error.Printfln("导出失败,您可以使用完全相同的指令再次尝试导出，会使用这次的缓存加速")
		if o.provider != nil {
			o.provider.Close()
		}
		return
	}
}

func (o *Exporter) doHop(
	teleportFn TeleportFn, hopPoint *structure.ExportHopPos,
	initWaitTime, minWaitTime, maxWaitTime float32,
) {
	teleportFn(hopPoint.Pos[0], hopPoint.Pos[2])
	minTimer := time.NewTimer(time.Duration(int(float32(time.Second) * minWaitTime)))
	maxTimer := time.NewTimer(time.Duration(int(float32(time.Second) * maxWaitTime)))
	time.Sleep(time.Duration(int(float32(time.Second) * initWaitTime)))
	allChunksHit := false
	for {
		select {
		case <-minTimer.C:
			if len(*o.allRequiredChunks) == 0 {
				return
			}
		case <-maxTimer.C:
			pterm.Info.Println("no new chunk arrived in max hop time after last chunk arrived, quit hop point")
			return
		case <-o.feedChan:
			maxTimer = time.NewTimer(time.Duration(int(float32(time.Second) * minWaitTime)))
			if !allChunksHit {
				_allHit := true
				for _, c := range hopPoint.LinkedChunk {
					if !c.CachedMark {
						_allHit = false
						break
					}
				}
				if _allHit {
					allChunksHit = true
				}
			}
		}
	}
}

func (o *Exporter) onLevelChunk(cd *mirror.ChunkData) {
	// fmt.Println("chunk arrived", cd.ChunkPos)
	if !o.listening {
		return
	}

	if tc := (*o.allRequiredChunks)[cd.ChunkPos]; tc != nil && !tc.CachedMark {
		// fmt.Println("HIT")
		o.updateHit(cd.ChunkPos)
		o.chunks[cd.ChunkPos] = cd
		o.provider.Write(cd)
		o.feedChan <- true
	} else {
		// fmt.Println("MISS")
		o.feedChan <- false
	}
}

func (o *UniverseExport) StartNewTask() {
	task := o.data.CurrentTask
	o.currentExporter = &Exporter{
		callStop:   false,
		task:       task,
		doneWaiter: make(chan struct{}),
		frame:      o.Frame,
		mu:         sync.Mutex{},
		listening:  false,
	}
	o.Frame.GetBotTaskScheduler().CommitUrgentTask(o.currentExporter)
	<-o.currentExporter.doneWaiter
	o.data.CurrentTask = nil
	o.currentExporter = nil
}

func parseSaveCmd(cmds []string) (startPos, endPos define.CubePos, structureName string, err error) {
	paramsNames := []string{"[建筑名]", "[起点x]", "[起点y]", "[起点z]", "[终点x]", "[终点y]", "[终点z]"}
	err = fmt.Errorf("保存指令错误, 应该为:\n %v "+strings.Join(paramsNames, " "), cmds[0])
	values := [6]int{}
	if len(cmds) < 8 {
		return
	}
	structureName = cmds[1]
	for i := 0; i < 6; i++ {
		if v, _err := strconv.Atoi(cmds[i+2]); _err != nil {
			err = fmt.Errorf(err.Error() + fmt.Sprintf("\n参数 %v 不正确", paramsNames[i+1]))
		} else {
			values[i] = v
		}
	}
	sortStart := func(i int) {
		startPos[i] = values[i]
		endPos[i] = values[i+3]
		if values[i] > values[i+3] {
			startPos[i] = values[i+3]
			endPos[i] = values[i]
		}
	}
	for i := 0; i < 3; i++ {
		sortStart(i)
	}
	err = nil
	return
}

func (o *UniverseExport) onImport(cmds []string) (stop bool) {
	if o.needDecision {
		o.needDecision = false
		o.cancelAll()
	}
	_cmds := []string{o.Triggers[0]}
	_cmds = append(_cmds, cmds...)
	startPos, endPos, structureName, err := parseSaveCmd(_cmds)
	if err != nil {
		pterm.Error.Println(err)
		return true
	}
	o.data.QueuedTasks = append(o.data.QueuedTasks, &exportTask{
		StructureName: structureName,
		StartPos:      startPos,
		EndPos:        endPos,
	})
	return true
}

func (o *UniverseExport) initFileData() {
	o.data = &ExportData{QueuedTasks: make([]*exportTask, 0)}
}

func (o *UniverseExport) cancelAll() {
	if o.currentExporter != nil {
		o.currentExporter.cancel()
	}
	o.initFileData()
}

func (o *UniverseExport) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.initFileData()
	if err := o.Frame.GetJsonData(o.FileName, o.data); err != nil {
		panic(err)
	}
	if o.data == nil {
		o.initFileData()
	}
	o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: strings.Join(paramsNames, " "),
			FinalTrigger: false,
			Usage:        "导出建筑 (同时导出为 schem 基岩版存档 mcworld，适合 国际服、amulet mcedit 直接打开)",
		},
		OptionalOnTriggerFn: o.onImport,
	})
	if !o.AutoContinueImport && (o.data.CurrentTask != nil || len(o.data.QueuedTasks) > 0) {
		o.needDecision = true
		o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
			MenuEntry: defines.MenuEntry{
				Triggers:     o.ContinueTriggers,
				ArgumentHint: "",
				FinalTrigger: false,
				Usage:        "从之前的断点继续导出",
			},
			OptionalOnTriggerFn: func(cmds []string) (stop bool) {
				o.needDecision = false
				return true
			},
		})
	}
	o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.CancelTriggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "取消所有导出任务",
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			o.cancelAll()
			return true
		},
	})
	// o.Frame.GetGameListener().AppendOnBlockUpdateInfoCallBack(o.onBlockUpdate)
	o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.OnLevelChunk)
}

func (o *UniverseExport) OnLevelChunk(cd *mirror.ChunkData) {
	if o.currentExporter != nil {
		o.currentExporter.onLevelChunk(cd)
	}
}

func (o *UniverseExport) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	if cfg.Version == "0.0.1" {
		cfg.Configs["继续导出的触发词"] = []string{"继续导出"}
		cfg.Configs["取消导出的触发词"] = []string{"取消导出"}
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
}

func (o *UniverseExport) Activate() {
	t := time.NewTicker(time.Second)
	for range t.C {
		if o.needDecision {
			continue
		}
		if o.data.CurrentTask == nil {
			if len(o.data.QueuedTasks) > 0 {
				o.data.CurrentTask = o.data.QueuedTasks[0]
				o.data.QueuedTasks = o.data.QueuedTasks[1:]
				o.fileChange = true
			}
		}
		if o.data.CurrentTask != nil {
			o.StartNewTask()
			o.currentExporter = nil
			o.data.CurrentTask = nil
			o.fileChange = true
		}
	}
}

func (o *UniverseExport) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.data)
		}
	}
	return nil
}

func (o *UniverseExport) Stop() error {
	fmt.Println("正在保存: " + o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.data)
}
