package components

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"phoenixbuilder/omega/utils/structure"
	"runtime"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

type universeImportTask struct {
	Path     string         `json:"路径"`
	Progress int            `json:"进度"`
	Offset   define.CubePos `json:"基准点"`
}
type UniverseImportData struct {
	CurrentTask *universeImportTask   `json:"当前正在处理的任务"`
	QueuedTasks []*universeImportTask `json:"排队中的任务"`
}

type UniverseImport struct {
	*defines.BasicComponent
	Triggers           []string `json:"后台触发词"`
	ImportSpeed        int      `json:"每秒导入普通方块数目"`
	FileName           string   `json:"断点续导记录文件"`
	AutoContinueImport bool     `json:"Omega启动时是否自动继续导入"`
	IgnoreBlockNbt     bool     `json:"忽略方块nbt信息"`
	fileChange         bool
	needDecision       bool
	data               *UniverseImportData
	currentBuilder     *Importor
}

type Importor struct {
	frontendStopper func()
	middleStopper   func()
	finalFeeder     chan *structure.IOBlock
	builder         *structure.Builder
	task            *universeImportTask
	doneWaiter      chan struct{}
	speed           int
}

func (o *Importor) cancel() {
	o.frontendStopper()
	o.middleStopper()
	o.builder.Stop = true
}

func (o *Importor) Activate() {
	pterm.Info.Printfln("开始处理任务 %v 起点(%v %v %v) 从 %v 方块处开始导入", o.task.Path, o.task.Offset[0], o.task.Offset[1], o.task.Offset[2], o.task.Progress)
	o.builder.Build(o.finalFeeder, o.speed)
	close(o.doneWaiter)
}

func (o *UniverseImport) getFrontEnd(data []byte, infoSender func(s string)) (blockFeeder chan *structure.IOBlock, stopFn func(), suggestMinCacheChunks int, totalBlocks int, err error) {
	suggestMinCacheChunks = 0
	if blockFeeder, stopFn, _suggestMinCacheChunks, totalBlocks, err := structure.DecodeSchem(data, infoSender); err == nil {
		return blockFeeder, stopFn, _suggestMinCacheChunks, totalBlocks, err
	} else {
		pterm.Warning.Printfln("文件无法被 schem 解析器解析，将尝试下一个解析器 (%v)", err)
	}

	if blockFeeder, stopFn, _suggestMinCacheChunks, totalBlocks, err := structure.DecodeSchematic(data, infoSender); err == nil {
		return blockFeeder, stopFn, _suggestMinCacheChunks, totalBlocks, err
	} else {
		pterm.Warning.Printfln("文件无法被 schematic 解析器解析，将尝试下一个解析器 (%v)", err)
	}
	return nil, nil, 0, 0, fmt.Errorf("无法找到合适的解析器")
}

func (o *UniverseImport) StartNewTask() {
	task := o.data.CurrentTask
	path := task.Path
	if task.Progress < 0 {
		task.Progress = 0
	}
	pterm.Info.Printfln("尝试处理任务 %v 起点(%v %v %v) 从 %v 方块处开始导入", task.Path, task.Offset[0], task.Offset[1], task.Offset[2], task.Progress)
	data := []byte{}
	if fp, err := os.OpenFile(path, os.O_RDONLY, 0644); err == nil {
		data, err = ioutil.ReadAll(fp)
		if err != nil {
			pterm.Error.Printfln("无法读取文件 %v 的数据 (%v)", path, err)
			o.data.CurrentTask = nil
			return
		}
	} else {
		pterm.Error.Printfln("无法读取文件 %v 的数据 (%v)", path, err)
		o.data.CurrentTask = nil
		return
	}
	if feeder, stopFn, suggestMinCacheChunks, totalBlocks, err := o.getFrontEnd(data, func(s string) {
		pterm.Info.Println(s)
	}); err == nil {
		baseProgress := task.Progress
		pterm.Success.Println("文件成功被解析,将开始优化导入顺序")
		if runtime.GOOS == "windows" && o.ImportSpeed > 100 {
			pterm.Error.Println("受限于windows计时器精度, 导入系统无法达到你指定的导入速度（和fb一样的问题），请考虑使用任意非windows系统（linux/macos/安卓/ios/使用linux的面板）实现导入")
		}
		o.currentBuilder = &Importor{
			frontendStopper: stopFn,
			task:            task,
			speed:           o.ImportSpeed,
		}
		o.currentBuilder.doneWaiter = make(chan struct{})
		progressUpdateInterval := o.ImportSpeed + 1
		progressBar := pterm.DefaultProgressbar.WithTotal(totalBlocks).WithTitle("Task: " + task.Path)
		lastProgress := 0
		updateProgress := func(currBlock int) {
			currProgress := 1 + baseProgress + currBlock
			increasementProgress := currProgress - lastProgress
			if increasementProgress > 0 {
				progressBar.Add(increasementProgress)
			} else {
				pterm.Error.Println("Negative increasementProgress: %v=(1+%v+%v)-%v ", increasementProgress, baseProgress, currBlock, lastProgress)
			}

			lastProgress = currProgress
			task.Progress = currProgress
			o.fileChange = true

			// 因为 omega 启动器设计失误（每次读一行）我不得不这么做
			fmt.Println()        // 下移一行（打出）
			fmt.Print("\033[1A") // 回到上一行
			fmt.Print("\033[K")  // 清除该行
		}
		ProgressUpdater := func(currBlock int) {
			if currBlock == 0 {
				pterm.Success.Printfln("可以开始导入了, 速度为 %v", o.ImportSpeed)
				progressBar, _ = progressBar.Start()
				updateProgress(currBlock)
			}
			if currBlock%progressUpdateInterval == 0 {
				updateProgress(currBlock)
			}
		}
		sender := o.Frame.GetGameControl().SendWOCmd
		builder := &structure.Builder{
			BlockCmdSender: func(cmd string) {
				// fmt.Println(cmd)
				sender(cmd)
			},
			TpCmdSender: func(cmd string) {
				o.Frame.GetGameControl().SendCmd(cmd)
			},
			ProgressUpdater: ProgressUpdater,
			FinalWaitTime:   3,
			IgnoreNbt:       o.IgnoreBlockNbt,
			InitPosGetter:   o.GetBotPos,
		}
		if suggestMinCacheChunks < 256 {
			suggestMinCacheChunks = 256
		}
		pterm.Info.Println("最大缓冲区块数量: ", suggestMinCacheChunks)
		middleFeeder, middleStopFn := structure.AlterImportPosStartAndSpeedWithReArrangeOnce(feeder, task.Offset, task.Progress, suggestMinCacheChunks, 16*16*16*24*3)
		o.currentBuilder.finalFeeder = middleFeeder
		o.currentBuilder.middleStopper = middleStopFn
		o.currentBuilder.builder = builder
		o.Frame.GetBotTaskScheduler().CommitUrgentTask(o.currentBuilder)
		<-o.currentBuilder.doneWaiter
		pterm.Success.Printfln("\n导入完成 %v ", path)
	} else {
		pterm.Error.Println("无法解析文件 %v ", path)
	}
	o.data.CurrentTask = nil
}

func (o *Importor) onBlockUpdate(pos define.CubePos, origRTID, currentRTID uint32) {
	o.builder.OnBlockUpdate(pos, origRTID, currentRTID)
}

func (o *UniverseImport) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
}

func (o *UniverseImport) onBlockUpdate(pos define.CubePos, origRTID, currentRTID uint32) {
	if o.currentBuilder != nil {
		o.currentBuilder.onBlockUpdate(pos, origRTID, currentRTID)
	}
}

func (o *UniverseImport) GetBotPos() define.CubePos {
	p := o.Frame.GetUQHolder().BotPos.Position
	return define.CubePos{int(p.X()), int(p.Y()), int(p.Z())}
}

func (o *UniverseImport) onImport(cmds []string) (stop bool) {
	if o.needDecision {
		o.needDecision = false
		o.cancelAll()
	}
	if len(cmds) < 4 {
		pterm.Error.Printfln("导入指令格式不正确，应该为 %v [路径] [x] [y] [z]", o.Triggers[0])
		return true
	}
	path := cmds[0]
	path = o.Frame.GetRelativeFileName(path)
	find, _, errStack := utils.GetFileNotFindStack(path)
	if !find {
		pterm.Error.Println("文件 %v 无法找到，具体问题如下：", path)
		for _, l := range errStack {
			pterm.Error.Println(l)
		}
		return true
	}
	start := define.CubePos{0, 0, 0}
	if i, err := strconv.Atoi(cmds[1]); err != nil {
		pterm.Error.Println("参数 [x] 不正确 ", err.Error())
		return true
	} else {
		start[0] = i
	}
	if i, err := strconv.Atoi(cmds[2]); err != nil {
		pterm.Error.Println("参数 [y] 不正确 ", err.Error())
		return true
	} else {
		start[1] = i
	}
	if i, err := strconv.Atoi(cmds[3]); err != nil {
		pterm.Error.Println("参数 [z] 不正确 ", err.Error())
		return true
	} else {
		start[2] = i
	}
	o.data.QueuedTasks = append(o.data.QueuedTasks, &universeImportTask{
		Path:     path,
		Offset:   start,
		Progress: 0,
	})
	return true
}

func (o *UniverseImport) cancelAll() {
	if o.currentBuilder != nil {
		o.currentBuilder.cancel()
	}
	o.initFileData()
}

func (o *UniverseImport) initFileData() {
	o.data = &UniverseImportData{QueuedTasks: make([]*universeImportTask, 0)}
}

func (o *UniverseImport) Inject(frame defines.MainFrame) {
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
			ArgumentHint: "[路径] [x] [y] [z]",
			FinalTrigger: false,
			Usage:        "导入建筑(目前仅支持 schem，其他文件类型将在后续加入)",
		},
		OptionalOnTriggerFn: o.onImport,
	})
	if !o.AutoContinueImport && (o.data.CurrentTask != nil || len(o.data.QueuedTasks) > 0) {
		o.needDecision = true
		o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
			MenuEntry: defines.MenuEntry{
				Triggers:     []string{"断点续导"},
				ArgumentHint: "",
				FinalTrigger: false,
				Usage:        "从之前的断点继续导入",
			},
			OptionalOnTriggerFn: func(cmds []string) (stop bool) {
				o.needDecision = false
				return true
			},
		})
	}
	o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     []string{"取消导入"},
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "取消所有导入任务",
		},
		OptionalOnTriggerFn: func(cmds []string) (stop bool) {
			o.cancelAll()
			return true
		},
	})
	// o.Frame.GetGameListener().AppendOnBlockUpdateInfoCallBack(o.onBlockUpdate)
}

func (o *UniverseImport) Activate() {
	t := time.NewTicker(time.Second)
	for range t.C {
		if o.needDecision {
			continue
		}
		// if o.currentBuilder != nil {
		// 	if o.currentBuilder.done() {
		// 		o.currentBuilder = nil
		// 		o.data.CurrentTask = nil
		// 		o.fileChange = true
		// 	} else {
		// 		continue
		// 	}
		// }
		if o.data.CurrentTask == nil {
			if len(o.data.QueuedTasks) > 0 {
				o.data.CurrentTask = o.data.QueuedTasks[0]
				o.data.QueuedTasks = o.data.QueuedTasks[1:]
				o.fileChange = true
			}
		}
		if o.data.CurrentTask != nil {
			o.StartNewTask()
			o.currentBuilder = nil
			o.data.CurrentTask = nil
			o.fileChange = true
		}
	}
}

func (o *UniverseImport) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.data)
		}
	}
	return nil
}

func (o *UniverseImport) Stop() error {
	fmt.Println("正在保存: " + o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.data)
}
