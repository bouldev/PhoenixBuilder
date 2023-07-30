package task

import (
	"fmt"
	NBTAssigner "phoenixbuilder/fastbuilder/bdump/nbt_assigner"
	"phoenixbuilder/fastbuilder/builder"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/environment"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/types"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"go.uber.org/atomic"
)

const (
	TaskStateUnknown     = 0
	TaskStateRunning     = 1
	TaskStatePaused      = 2
	TaskStateDied        = 3
	TaskStateCalculating = 4
	TaskStateSpecialBrk  = 5
)

type Task struct {
	TaskId        int64
	CommandLine   string
	OutputChannel chan *types.Module
	ContinueLock  sync.Mutex
	State         byte
	Type          byte
	AsyncInfo
	Config *configuration.FullConfig
	holder *TaskHolder
}

type AsyncInfo struct {
	Built     int
	Total     int
	BeginTime time.Time
}

type TaskHolder struct {
	TaskIdCounter       *atomic.Int64
	TaskMap             sync.Map
	BrokSender          chan string
	ExtraDisplayStrings []string
}

func NewTaskHolder() *TaskHolder {
	return &TaskHolder{
		TaskIdCounter:       atomic.NewInt64(0),
		TaskMap:             sync.Map{},
		BrokSender:          make(chan string),
		ExtraDisplayStrings: []string{},
	}
}

func GetStateDesc(st byte) string {
	if st == 0 {
		return I18n.T(I18n.TaskTypeUnknown)
	} else if st == 1 {
		return I18n.T(I18n.TaskTypeRunning)
	} else if st == 2 {
		return I18n.T(I18n.TaskTypePaused)
	} else if st == 3 {
		return I18n.T(I18n.TaskTypeDied)
	} else if st == 4 {
		return I18n.T(I18n.TaskTypeCalculating)
	} else if st == 5 {
		return I18n.T(I18n.TaskTypeSpecialTaskBreaking)
	}
	return "???????"
}

func (task *Task) Finalize() {
	task.State = TaskStateDied
	task.holder.TaskMap.Delete(task.TaskId)
}

func (task *Task) Pause() {
	if task.State == TaskStatePaused {
		return
	}
	task.ContinueLock.Lock()
	if task.State == TaskStateDied {
		task.ContinueLock.Unlock()
		return
	}
	task.State = TaskStatePaused
}

func (task *Task) Resume() {
	if task.State != TaskStatePaused {
		return
	}
	if task.Type == types.TaskTypeAsync {
		task.AsyncInfo.Total -= task.AsyncInfo.Built
		task.AsyncInfo.Built = 0
		task.AsyncInfo.BeginTime = time.Now()
	}
	task.State = TaskStateRunning
	task.ContinueLock.Unlock()
}

func (task *Task) Break() {
	if task.OutputChannel == nil {
		task.State = TaskStateSpecialBrk
		return
	}
	if task.State != TaskStatePaused {
		task.Pause()
	}
	if task.State == TaskStateDied {
		return
	}
	chann := task.OutputChannel
	for {
		_, ok := <-chann
		if !ok {
			break
		}
		if false {
			//fmt.Printf("%v\n",blk)
		}
	}
	if task.Type == types.TaskTypeAsync {
		// Avoid progress displaying
		if task.State != TaskStatePaused {
			return
		}
		task.State = TaskStateCalculating
		task.ContinueLock.Unlock()
		return
	}
	task.Resume()
}

func (holder *TaskHolder) FindTask(taskId int64) *Task {
	t, _ := holder.TaskMap.Load(taskId)
	ta, _ := t.(*Task)
	return ta
}

func CreateTask(commandLine string, env *environment.PBEnvironment) *Task {
	holder := env.TaskHolder.(*TaskHolder)
	gameInterface := env.GameInterface
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err != nil {
		gameInterface.Output(fmt.Sprintf(I18n.T(I18n.TaskFailedToParseCommand), err))
		return nil
	}
	fcfg := configuration.ConcatFullConfig(cfg, configuration.GlobalFullConfig(env).Delay())
	dcfg := fcfg.Delay()

	gameInterface.SendWSCommand("gamemode c")
	blockschannel := make(chan *types.Module, 10240)
	task := &Task{
		TaskId:        holder.TaskIdCounter.Add(1),
		CommandLine:   commandLine,
		OutputChannel: blockschannel,
		State:         TaskStateCalculating,
		Type:          configuration.GlobalFullConfig(env).Global().TaskCreationType,
		Config:        fcfg,
		holder:        holder,
	}
	taskid := task.TaskId
	holder.TaskMap.Store(taskid, task)
	var asyncblockschannel chan *types.Module
	if task.Type == types.TaskTypeAsync {
		asyncblockschannel = blockschannel
		blockschannel = make(chan *types.Module)
		task.OutputChannel = blockschannel
		go func() {
			var blocks []*types.Module
			for {
				curblock, ok := <-asyncblockschannel
				if !ok {
					break
				}
				blocks = append(blocks, curblock)
			}
			task.State = TaskStateRunning
			t1 := time.Now()
			total := len(blocks)
			task.AsyncInfo = AsyncInfo{
				Built:     0,
				Total:     total,
				BeginTime: t1,
			}
			skipBlocks := int(cfg.ResumeFrom * float64(task.AsyncInfo.Total) / 100.0)
			skipBlocks -= 10
			if skipBlocks <= 0 {
				skipBlocks = 0
			} else {
				if skipBlocks > task.AsyncInfo.Total {
					skipBlocks = task.AsyncInfo.Total
				}
				fmt.Printf(I18n.T(I18n.Task_ResumeBuildFrom)+"\n", skipBlocks)
			}
			for _, blk := range blocks {
				if skipBlocks != 0 && task.AsyncInfo.Built == skipBlocks-1 {
					skipBlocks = 0
					task.AsyncInfo.Total -= task.AsyncInfo.Built
					task.AsyncInfo.Built = 0
					continue
				}
				if task.AsyncInfo.Built >= skipBlocks {
					blockschannel <- blk
					if skipBlocks != 0 {
						skipBlocks = 0
					}
				}
				task.AsyncInfo.Built++
			}
			close(blockschannel)
		}()
	} else {
		task.State = TaskStateRunning
	}
	go func() {
		var doDelay func()
		if runtime.GOOS == "windows" {
			delayTime := time.Duration(dcfg.Delay*100) * time.Microsecond
			oneHundredCounter := 0
			doDelay = func() {
				if oneHundredCounter == 100 {
					time.Sleep(delayTime)
					oneHundredCounter = 0
				}
				oneHundredCounter++
			}
		} else {
			delayTime := time.Duration(dcfg.Delay) * time.Microsecond
			doDelay = func() {
				time.Sleep(delayTime)
			}
		}
		t1 := time.Now()
		blkscounter := 0
		tothresholdcounter := 0
		isFastMode := false
		if dcfg.DelayMode == types.DelayModeDiscrete || dcfg.DelayMode == types.DelayModeNone {
			isFastMode = true
		} else {
			//isFastMode=false
			gameInterface.SendWSCommand("gamemode c")
			gameInterface.SendWSCommand("gamerule sendcommandfeedback true")
		}
		for {
			task.ContinueLock.Lock()
			task.ContinueLock.Unlock()
			curblock, ok := <-blockschannel
			if !ok {
				if blkscounter == 0 {
					gameInterface.Output(fmt.Sprintf(I18n.T(I18n.Task_D_NothingGenerated), taskid))
					runtime.GC()
					task.Finalize()
					return
				}
				timeUsed := time.Now().Sub(t1)
				gameInterface.Output(fmt.Sprintf(I18n.T(I18n.Task_Summary_1), taskid, blkscounter))
				gameInterface.Output(fmt.Sprintf(I18n.T(I18n.Task_Summary_2), taskid, timeUsed.Seconds()))
				gameInterface.Output(fmt.Sprintf(I18n.T(I18n.Task_Summary_3), taskid, float64(blkscounter)/timeUsed.Seconds()))
				runtime.GC()
				task.Finalize()
				return
			}
			if blkscounter%20 == 0 {
				gameInterface.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", curblock.Point.X, curblock.Point.Y, curblock.Point.Z), true)
			}
			blkscounter++
			if curblock.NBTMap != nil {
				err := NBTAssigner.PlaceBlockWithNBTData(
					gameInterface,
					curblock,
					&NBTAssigner.BlockAdditionalData{
						Settings: cfg,
						FastMode: isFastMode,
						Others:   nil,
					},
				)
				if err != nil {
					pterm.Warning.Printf("CreateTask: %v\n", err)
				}
			} else if !cfg.ExcludeCommands && curblock.CommandBlockData != nil {
				newStruct := NBTAssigner.CommandBlock{
					BlockEntity: &NBTAssigner.BlockEntity{
						Interface: gameInterface,
						AdditionalData: NBTAssigner.BlockAdditionalData{
							Position: [3]int32{int32(curblock.Point.X), int32(curblock.Point.Y), int32(curblock.Point.Z)},
							Settings: cfg,
							FastMode: isFastMode,
							Others:   nil,
						},
					},
					ShouldPlaceBlock: false,
				}
				err := newStruct.PlaceCommandBlockLegacy(curblock, cfg)
				if err != nil {
					pterm.Warning.Printf("%v\n", err)
				}
			} else if curblock.ChestSlot != nil {
				gameInterface.SendSettingsCommand(commands_generator.ReplaceItemInContainerRequest(curblock, ""), true)
			} else if len(cfg.Entity) != 0 {
				gameInterface.SendSettingsCommand(commands_generator.SummonRequest(curblock, cfg), true)
			} else {
				gameInterface.SendSettingsCommand(commands_generator.SetBlockRequest(curblock, cfg), true)
			}
			if dcfg.DelayMode == types.DelayModeContinuous {
				doDelay()
			} else if dcfg.DelayMode == types.DelayModeDiscrete {
				tothresholdcounter++
				if tothresholdcounter >= dcfg.DelayThreshold {
					tothresholdcounter = 0
					time.Sleep(time.Duration(dcfg.Delay) * time.Second)
				}
			}
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				gameInterface.Output(fmt.Sprintf("[Task %d] Fatal error: %v", taskid, err))
				close(blockschannel)
			}
		}()
		if task.Type == types.TaskTypeAsync {
			err := builder.Generate(cfg, asyncblockschannel)
			close(asyncblockschannel)
			if err != nil {
				gameInterface.Output(fmt.Sprintf("[%s %d] %s: %v", I18n.T(I18n.TaskTTeIuKoto), taskid, I18n.T(I18n.ERRORStr), err))
			}
			return
		}
		err := builder.Generate(cfg, blockschannel)
		close(blockschannel)
		if err != nil {
			gameInterface.Output(fmt.Sprintf("[%s %d] %s: %v", I18n.T(I18n.TaskTTeIuKoto), taskid, I18n.T(I18n.ERRORStr), err))
		}
	}()
	return task
}

func CheckHasWorkingTask(env *environment.PBEnvironment) bool {
	holder := env.TaskHolder.(*TaskHolder)
	has := false
	holder.TaskMap.Range(func(_tid interface{}, _v interface{}) bool {
		has = true
		return false
	})
	return has
}

func InitTaskStatusDisplay(env *environment.PBEnvironment) {
	holder := env.TaskHolder.(*TaskHolder)
	go func() {
		for {
			str := <-holder.BrokSender
			env.GameInterface.Output(str)
		}
	}()
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			<-ticker.C
			env.ActivateTaskStatus <- true
		}
	}()
	go func() {
		for {
			<-env.ActivateTaskStatus
			if configuration.GlobalFullConfig(env).Global().TaskDisplayMode == types.TaskDisplayNo {
				continue
			}
			var displayStrs []string
			holder.TaskMap.Range(func(_tid interface{}, _v interface{}) bool {
				tid, _ := _tid.(int64)
				v, _ := _v.(*Task)

				addstr := fmt.Sprintf("Task ID %d - %s - %s [%s]", tid, v.Config.Main().Execute, GetStateDesc(v.State), types.MakeTaskType(v.Type))
				if v.Type == types.TaskTypeAsync && v.State == TaskStateRunning {
					addstr = fmt.Sprintf("%s\nProgress: %s", addstr, ProgressThemes[0](&v.AsyncInfo))
				}
				displayStrs = append(displayStrs, addstr)
				commands_generator.AdditionalTitleCb(addstr)
				return true
			})
			displayStrs = append(displayStrs, holder.ExtraDisplayStrings...)
			if len(displayStrs) == 0 {
				continue
			}
			env.GameInterface.Title(strings.Join(displayStrs, "\n"))
		}
	}()
}
