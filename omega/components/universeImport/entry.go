package universe_import

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/assembler"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"phoenixbuilder/omega/utils/structure"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
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
	BoostRate          float64  `json:"超频加速比"`
	ContinueTriggers   []string `json:"继续导入的触发词"`
	CancelTriggers     []string `json:"取消导入的触发词"`
	TargetOfGetCmd     string   `json:"get指令的目标"`
	PosInferredByGet   *define.CubePos
	fileChange         bool
	needDecision       bool
	data               *UniverseImportData
	currentBuilder     *Importor
}

type Importor struct {
	frontendStopper func()
	middleStopper   func()
	finalFeeder     chan *structure.IOBlockForBuilder
	builder         *structure.Builder
	task            *universeImportTask
	doneWaiter      chan struct{}
	speed           int
	boostSleepTime  time.Duration
	frame           defines.MainFrame
}

func (o *Importor) cancel() {
	o.frontendStopper()
	o.middleStopper()
	o.builder.Stop = true
}

func (o *Importor) Activate() {
	pterm.Info.Printfln("开始处理任务 %v 起点(%v %v %v) 从 %v 方块处开始导入", o.task.Path, o.task.Offset[0], o.task.Offset[1], o.task.Offset[2], o.task.Progress)
	o.frame.GetGameListener().GetChunkAssembler().AdjustSendPeriod(assembler.REQUEST_LAZY)
	o.builder.Build(o.finalFeeder, o.speed, o.boostSleepTime)
	o.frame.GetGameListener().GetChunkAssembler().AdjustSendPeriod(assembler.REQUEST_NORMAL)
	close(o.doneWaiter)
}

func (o *UniverseImport) getFrontEnd(fileName string, data []byte, infoSender func(s string)) (blockFeeder chan *structure.IOBlockForDecoder, stopFn func(), suggestMinCacheChunks int, totalBlocks int, err error) {
	suggestMinCacheChunks = 0

	tryBDX := func() bool {
		if blockFeeder, stopFn, suggestMinCacheChunks, totalBlocks, err = structure.DecodeBDX(data, infoSender); err == nil {
			return true
		} else {
			pterm.Warning.Printfln("文件无法被 bdx 解析器解析，将尝试下一个解析器 (%v)", err)
			return false
		}
	}
	trySchem := func() bool {
		if blockFeeder, stopFn, suggestMinCacheChunks, totalBlocks, err = structure.DecodeSchem(data, infoSender); err == nil {
			return true
		} else {
			pterm.Warning.Printfln("文件无法被 schem 解析器解析，将尝试下一个解析器 (%v)", err)
			return false
		}
	}
	trySchematic := func() bool {
		if blockFeeder, stopFn, suggestMinCacheChunks, totalBlocks, err = structure.DecodeSchematic(data, infoSender); err == nil {
			return true
		} else {
			pterm.Warning.Printfln("文件无法被 schematic 解析器解析，将尝试下一个解析器 (%v)", err)
			return false
		}
	}
	if strings.Contains(fileName, ".bdx") {
		if tryBDX() {
			return
		}
		if trySchem() {
			return
		}
		if trySchematic() {
			return
		}
	} else if strings.Contains(fileName, ".schematic") {
		if trySchematic() {
			return
		}
		if tryBDX() {
			return
		}
		if trySchem() {
			return
		}
	} else {
		if trySchem() {
			return
		}
		if trySchematic() {
			return
		}
		if tryBDX() {
			return
		}
	}
	return nil, nil, 0, 0, fmt.Errorf("无法找到合适的解析器")
}

func (o *UniverseImport) StartNewTask() {
	// o.Frame.NoChunkRequestCache()
	// o.Frame.GetGameControl().SendCmd("gamerule commandblocks enabled false ")
	task := o.data.CurrentTask
	filePath := task.Path
	if task.Progress < 0 {
		task.Progress = 0
	}
	pterm.Info.Printfln("尝试处理任务 %v 起点(%v %v %v) 从 %v 方块处开始导入", task.Path, task.Offset[0], task.Offset[1], task.Offset[2], task.Progress)
	data := []byte{}
	fileName := path.Base(task.Path)
	if fp, err := os.OpenFile(filePath, os.O_RDONLY, 0644); err == nil {
		data, err = io.ReadAll(fp)
		if err != nil {
			pterm.Error.Printfln("无法读取文件 %v 的数据 (%v)", filePath, err)
			o.data.CurrentTask = nil
			return
		}
	} else {
		pterm.Error.Printfln("无法读取文件 %v 的数据 (%v)", filePath, err)
		o.data.CurrentTask = nil
		return
	}
	if feeder, stopFn, suggestMinCacheChunks, totalBlocks, err := o.getFrontEnd(fileName, data, func(s string) {
		pterm.Info.Println(s)
	}); err == nil {
		baseProgress := task.Progress
		pterm.Success.Println("文件成功被解析,将开始优化导入顺序")
		boostSleepTime := time.Duration(float64(time.Second) * ((4096.) / (o.BoostRate * float64(o.ImportSpeed))))
		o.currentBuilder = &Importor{
			frontendStopper: stopFn,
			task:            task,
			speed:           o.ImportSpeed,
			boostSleepTime:  boostSleepTime,
			frame:           o.Frame,
		}
		o.currentBuilder.doneWaiter = make(chan struct{})
		progressUpdateInterval := o.ImportSpeed + 1
		if totalBlocks == 0 {
			totalBlocks = 1
		}
		taskName := path.Base(filePath)
		if len(taskName) > 10 {
			taskName = taskName[:10]
		}
		progressBar := pterm.DefaultProgressbar.WithTotal(totalBlocks - 1).WithTitle(taskName)
		lastBlock := 0
		startTime := time.Now()
		updateProgress := func(currBlock int) {
			defer func() {
				r := recover()
				if r != nil {
					pterm.Error.Println("请尝试让一行显示更多的字 (err %v)", r)
				}
			}()
			increasementProgress := currBlock - lastBlock
			lastBlock = currBlock
			if increasementProgress > 0 {
				progressBar.Add(increasementProgress)
			}
			task.Progress = baseProgress + currBlock
			o.fileChange = true
			metricDuration := time.Since(startTime).Seconds()
			realSpeed := float64(currBlock) / metricDuration
			progressBar.Title = taskName + fmt.Sprintf(" 实际速度 %d", int(realSpeed))
		}
		ProgressUpdater := func(currBlock int) {
			if currBlock == 0 {
				pterm.Success.Printfln("可以开始导入了, 速度为 %v", o.ImportSpeed)
				startTime = time.Now()
				progressBar, _ = progressBar.Start()
				if baseProgress > 0 {
					progressBar.Add(baseProgress)
				}
			} else if currBlock-lastBlock > progressUpdateInterval {
				updateProgress(currBlock)
			}
		}
		sender := o.Frame.GetGameControl().SendWOCmd
		builder := &structure.Builder{
			BlockCmdSender: func(cmd string) {
				// fmt.Println(cmd)
				sender(cmd)
			},
			NormalCmdSender: func(cmd string) {
				o.Frame.GetGameControl().SendCmd(cmd)
			},
			ProgressUpdater: ProgressUpdater,
			FinalWaitTime:   3,
			IgnoreNbt:       o.IgnoreBlockNbt,
			InitPosGetter:   o.GetBotPos,
			Ctrl:            o.Frame.GetGameControl(),
		}
		if suggestMinCacheChunks == 0 {
			suggestMinCacheChunks = 256
		}
		pterm.Info.Println("最大缓冲区块数量: ", suggestMinCacheChunks)
		middleFeeder, middleStopFn := structure.AlterImportPosStartAndSpeedWithReArrangeOnce(feeder, task.Offset, task.Progress, suggestMinCacheChunks, 16*16*16*24*3)
		o.currentBuilder.finalFeeder = middleFeeder
		o.currentBuilder.middleStopper = middleStopFn
		o.currentBuilder.builder = builder
		o.Frame.GetBotTaskScheduler().CommitUrgentTask(o.currentBuilder)
		<-o.currentBuilder.doneWaiter
		pterm.Success.Printfln("导入完成 %v ", filePath)
		// o.Frame.AllowChunkRequestCache()
	} else {
		pterm.Error.Printfln("无法解析文件 %v, %v", filePath, err)
	}
	o.data.CurrentTask = nil
}

func (o *Importor) onBlockUpdate(pos define.CubePos, origRTID, currentRTID uint32) {
	o.builder.OnBlockUpdate(pos, origRTID, currentRTID)
}

func (o *Importor) onLevelChunk(cd *mirror.ChunkData) {
	o.builder.OnLevelChunk(cd)
}

func (o *UniverseImport) Init(cfg *defines.ComponentConfig) {
	if cfg.Version == "0.0.1" {
		cfg.Configs["超频加速比"] = 10
		cfg.Configs["忽略方块nbt信息"] = false
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	if cfg.Version == "0.0.2" {
		cfg.Configs["继续导入的触发词"] = []string{"继续导入"}
		cfg.Configs["取消导入的触发词"] = []string{"取消导入"}
		cfg.Version = "0.0.3"
		cfg.Upgrade()
	}
	if cfg.Version == "0.0.3" {
		cfg.Configs["get指令的目标"] = "2401PT"
		cfg.Version = "0.0.4"
		cfg.Upgrade()
	}
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
		if o.PosInferredByGet == nil {
			pterm.Error.Printfln("导入指令格式不正确，应该为 %v [路径] [x] [y] [z]", o.Triggers[0])
			return true
		} else {
			_cmds := []string{cmds[0], fmt.Sprintf("%v", o.PosInferredByGet[0]), fmt.Sprintf("%v", o.PosInferredByGet[1]), fmt.Sprintf("%v", o.PosInferredByGet[2])}
			_cmds = append(_cmds, cmds[1:]...)
			cmds = _cmds
		}
	}
	filePath := cmds[0]
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(filePath, "\"") || strings.HasSuffix(filePath, "\"") {
			filePath = strings.Trim(filePath, "\"")
		}
		if (!utils.IsDir(filePath)) && (!utils.IsFile(filePath)) {
			pathAlter := strings.ReplaceAll(filePath, "/", "\\")
			if (!utils.IsDir(pathAlter)) && (!utils.IsFile(pathAlter)) {
				pathAlter := strings.ReplaceAll(filePath, "//", "\\")
				if (!utils.IsDir(pathAlter)) && (!utils.IsFile(pathAlter)) {
					pathAlter := strings.ReplaceAll(filePath, "//", "/")
					if (!utils.IsDir(pathAlter)) && (!utils.IsFile(pathAlter)) {
						pathAlter = strings.ReplaceAll(filePath, "/", "//")
						if (!utils.IsDir(pathAlter)) && (!utils.IsFile(pathAlter)) {
							pathAlter = strings.ReplaceAll(filePath, "\\", "/")
							if (!utils.IsDir(pathAlter)) && (!utils.IsFile(pathAlter)) {
								pathAlter = strings.ReplaceAll(filePath, "\\", "//")
								if (!utils.IsDir(pathAlter)) && (!utils.IsFile(pathAlter)) {
									// 这总不能还是斜杠的问题了吧？！
									filePath = o.Frame.GetRelativeFileName(filePath)
								} else {
									filePath = pathAlter
								}
							} else {
								filePath = pathAlter
							}
						} else {
							filePath = pathAlter
						}
					} else {
						filePath = pathAlter
					}
				} else {
					filePath = pathAlter
				}
			} else {
				filePath = pathAlter
			}
		}
	} else if !strings.HasPrefix(filePath, "/") {
		if (!utils.IsDir(filePath)) && (!utils.IsFile(filePath)) {
			filePath = o.Frame.GetRelativeFileName(filePath)
			if (!utils.IsDir(filePath)) && (!utils.IsFile(filePath)) {
				filePath = path.Join(o.Frame.GetStorageRoot(), filePath)
			}
		}
	}
	find, _, errStack := utils.GetFileNotFindStack(filePath)
	if !find {
		pterm.Error.Println("文件 %v 无法找到，具体问题如下：", filePath)
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
	fp, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		pterm.Error.Println("文件 %v 无法打开，具体问题为: %v", filePath, err)
	} else {
		defer fp.Close()
		img, err := imaging.Decode(fp)
		if err != nil {
			o.data.QueuedTasks = append(o.data.QueuedTasks, &universeImportTask{
				Path:     filePath,
				Offset:   start,
				Progress: 0,
			})
		} else {
			// is an image
			go func() {
				dir := o.Frame.GetOmegaNormalCacheDir("image_import", path.Base(filePath))
				os.MkdirAll(dir, 0755)
				staructureFile, err := PreProcessImage(img, dir, cmds[4:])
				if err != nil {
					pterm.Error.Println(err)
				} else {
					suggestX := int(math.Round(float64(start.X())/64.)) * 64
					suggestZ := int(math.Round(float64(start.Z())/64)) * 64
					if (suggestX/64)%2 == 0 {
						suggestX += 64
					}
					if (suggestZ/64)%2 == 0 {
						suggestZ += 64
					}
					if suggestX != start.X() || suggestZ != start.Z() {
						pterm.Warning.Printfln("对于地图画，建议将导入点设为 %v %v %v", suggestX, start.Y(), suggestZ)
					}
					o.data.QueuedTasks = append(o.data.QueuedTasks, &universeImportTask{
						Path:     staructureFile,
						Offset:   start,
						Progress: 0,
					})
				}
			}()
		}
	}
	return true
}

func (o *UniverseImport) onGetCalled(cmds []string) (stop bool) {
	target := o.TargetOfGetCmd
	if len(cmds) > 0 {
		target = cmds[0]
	}
	utils.GetPos(o.Frame.GetGameControl(), target, func(results []utils.QueryPosResult, err error) {
		if err != nil {
			pterm.Error.Printfln("无法获取名为 %v 目标的坐标, 请检查 %v 是否在服务器或者考虑调整设置, %v", target, target, err)
		} else {
			result := results[0]
			result.Position.Y -= 1.62001001834869
			o.PosInferredByGet = &define.CubePos{int(math.Floor(result.Position.X)), int(math.Floor(result.Position.Y)), int(math.Floor(result.Position.Z))}
			pterm.Info.Printfln("已经获得 %v 所在坐标 %v, 后续使用 load 指令时可以省略 [x] [y] [z]", target, o.PosInferredByGet)
		}
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
			ArgumentHint: "[路径] [x] [y] [z]  (对于图片而言，应该为 [路径] [x] [y] [z] [x方向地图数] [z方向地图数])",
			FinalTrigger: false,
			Usage:        "导入建筑，支持 bdx schem schmatic",
		},
		OptionalOnTriggerFn: o.onImport,
	})
	o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     []string{"get", "Get", "GET"},
			ArgumentHint: "[目标]",
			FinalTrigger: false,
			Usage:        "获取起始点",
		},
		OptionalOnTriggerFn: o.onGetCalled,
	})
	if !o.AutoContinueImport && (o.data.CurrentTask != nil || len(o.data.QueuedTasks) > 0) {
		o.needDecision = true
		o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
			MenuEntry: defines.MenuEntry{
				Triggers:     o.ContinueTriggers,
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
			Triggers:     o.CancelTriggers,
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
	// o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.OnLevelChunk)
}

func (o *UniverseImport) OnLevelChunk(cd *mirror.ChunkData) {
	if o.currentBuilder != nil {
		o.currentBuilder.onLevelChunk(cd)
	}
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
