package task

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/atomic"
	"phoenixbuilder/bridge/bridge_fmt"
	"phoenixbuilder/fastbuilder/builder"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/environment"
	"runtime"
	"strings"
	"sync"
	"time"
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
	TaskId int64
	CommandLine string
	OutputChannel chan *types.Module
	ContinueLock sync.Mutex
	State byte
	Type byte
	AsyncInfo
	Config *configuration.FullConfig
	holder *TaskHolder
}

type AsyncInfo struct {
	Built int
	Total int
	BeginTime time.Time
}

type TaskHolder struct {
	TaskIdCounter *atomic.Int64
	TaskMap sync.Map
	BrokSender chan string
	ExtraDisplayStrings []string
}

func NewTaskHolder() *TaskHolder {
	return &TaskHolder {
		TaskIdCounter: atomic.NewInt64(0),
		TaskMap: sync.Map {},
		BrokSender: make(chan string),
		ExtraDisplayStrings: []string {},
	}
}

func GetStateDesc(st byte) string {
	if st == 0 {
		return I18n.T(I18n.TaskTypeUnknown)
	}else if st==1 {
		return I18n.T(I18n.TaskTypeRunning)
	}else if st==2 {
		return I18n.T(I18n.TaskTypePaused)
	}else if st==3 {
		return I18n.T(I18n.TaskTypeDied)
	}else if st==4 {
		return I18n.T(I18n.TaskTypeCalculating)
	}else if st==5 {
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
	if task.Type==types.TaskTypeAsync {
		task.AsyncInfo.Total-=task.AsyncInfo.Built
		task.AsyncInfo.Built=0
	}
	task.State = TaskStateRunning
	task.ContinueLock.Unlock()
}

func (task *Task) Break() {
	if task.OutputChannel==nil {
		task.State=TaskStateSpecialBrk
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
		_, ok := <- chann
		if !ok {
			break
		}
		if false {
			//fmt.Printf("%v\n",blk)
		}
	}
	if task.Type==types.TaskTypeAsync {
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
	holder:=env.TaskHolder.(*TaskHolder)
	conn:=env.Connection.(*minecraft.Conn)
	cmdsender:=env.CommandSender
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err!=nil {
		cmdsender.Output(fmt.Sprintf(I18n.T(I18n.TaskFailedToParseCommand),err))
		return nil
	}
	fcfg := configuration.ConcatFullConfig(cfg, configuration.GlobalFullConfig(env).Delay())
	dcfg := fcfg.Delay()
	und, _ := uuid.NewUUID()
	cmdsender.SendWSCommand("gamemode c", und)
	blockschannel := make(chan *types.Module, 10240)
	task := &Task {
		TaskId: holder.TaskIdCounter.Add(1),
		CommandLine: commandLine,
		OutputChannel: blockschannel,
		State: TaskStateCalculating,
		Type: configuration.GlobalFullConfig(env).Global().TaskCreationType,
		Config: fcfg,
		holder: holder,
	}
	taskid := task.TaskId
	holder.TaskMap.Store(taskid, task)
	var asyncblockschannel chan *types.Module
	if task.Type==types.TaskTypeAsync {
		asyncblockschannel=blockschannel
		blockschannel=make(chan *types.Module)
		task.OutputChannel=blockschannel
		go func() {
			var blocks []*types.Module
			for {
				curblock, ok := <-asyncblockschannel
				if !ok {
					break
				}
				blocks=append(blocks,curblock)
			}
			task.State=TaskStateRunning
			t1 := time.Now()
			total := len(blocks)
			task.AsyncInfo=AsyncInfo {
				Built: 0,
				Total: total,
				BeginTime: t1,
			}
			skipBlocks:=int(cfg.ResumeFrom*float64(task.AsyncInfo.Total)/100.0)
			skipBlocks-=10
			if skipBlocks<=0{
				skipBlocks=0
			}else{
				if skipBlocks>task.AsyncInfo.Total{
					skipBlocks=task.AsyncInfo.Total
				}
				bridge_fmt.Printf(I18n.T(I18n.Task_ResumeBuildFrom)+"\n",skipBlocks)
			}
			for _, blk := range blocks {
				if task.AsyncInfo.Built>=skipBlocks{
					blockschannel <- blk
				}
				task.AsyncInfo.Built++
			}
			close(blockschannel)
		} ()
	}else{
		task.State=TaskStateRunning
	}
	go func() {
		isWindows:=false
		if runtime.GOOS == "windows" {
			isWindows=true
		}
		t1 := time.Now()
		blkscounter := 0
		tothresholdcounter := 0
		isFastMode := false
		if dcfg.DelayMode==types.DelayModeDiscrete||dcfg.DelayMode==types.DelayModeNone {
			isFastMode=true
		}else{
			//isFastMode=false
			cmdsender.SendWSCommand("gamemode c", und)
			cmdsender.SendWSCommand("gamerule sendcommandfeedback true", und)
		}
		request:=commands_generator.AllocateRequestString()
		for {
			task.ContinueLock.Lock()
			task.ContinueLock.Unlock()
			curblock, ok := <-blockschannel
			if !ok {
				if blkscounter == 0 {
					cmdsender.Output(fmt.Sprintf(I18n.T(I18n.Task_D_NothingGenerated),taskid))
					runtime.GC()
					task.Finalize()
					return
				}
				timeUsed := time.Now().Sub(t1)
				cmdsender.Output(fmt.Sprintf(I18n.T(I18n.Task_Summary_1), taskid, blkscounter))
				cmdsender.Output(fmt.Sprintf(I18n.T(I18n.Task_Summary_2), taskid, timeUsed.Seconds()))
				cmdsender.Output(fmt.Sprintf(I18n.T(I18n.Task_Summary_3), taskid, float64(blkscounter)/timeUsed.Seconds()))
				runtime.GC()
				task.Finalize()
				return
			}
			if blkscounter%20 == 0 {
				u_d, _ := uuid.NewUUID()
				cmdsender.SendWSCommand(fmt.Sprintf("tp %d %d %d",curblock.Point.X,curblock.Point.Y,curblock.Point.Z),u_d)
				// SettingsCommand is unable to teleport the player.
			}
			blkscounter++
			if !cfg.ExcludeCommands && curblock.CommandBlockData != nil {
				if curblock.Block != nil {
					commands_generator.SetBlockRequest(request,curblock, cfg)
					if !isFastMode {
						//<-time.After(time.Second)
						wc:=make(chan bool)
						(*cmdsender.GetBlockUpdateSubscribeMap()).Store(protocol.BlockPos{int32(curblock.Point.X),int32(curblock.Point.Y),int32(curblock.Point.Z)},wc)
						cmdsender.SendSizukanaCommand(*request)
						select {
						case <-wc:
							break
						case <-time.After(time.Second*2):
							(*cmdsender.GetBlockUpdateSubscribeMap()).Delete(protocol.BlockPos{int32(curblock.Point.X),int32(curblock.Point.Y),int32(curblock.Point.Z)})
						}
						close(wc)
					}else{
						cmdsender.SendSizukanaCommand(*request)
					}
				}
				cbdata:=curblock.CommandBlockData
				if(cfg.InvalidateCommands){
					cbdata.Command="|"+cbdata.Command
				}
				if !isFastMode {
					UUID:=uuid.New()
					w:=make(chan *packet.CommandOutput)
					(*cmdsender.GetUUIDMap()).Store(UUID.String(), w)
					cmdsender.SendWSCommand(fmt.Sprintf("tp %d %d %d",curblock.Point.X,curblock.Point.Y+1,curblock.Point.Z), UUID)
					select {
					case <-time.After(time.Second):
						(*cmdsender.GetUUIDMap()).Delete(UUID.String())
						break
					case <-w:
					}
					close(w)
				}
				conn.WritePacket(&packet.CommandBlockUpdate {
					Block: true,
					Position: protocol.BlockPos{int32(curblock.Point.X),int32(curblock.Point.Y),int32(curblock.Point.Z)},
					Mode: cbdata.Mode,
					NeedsRedstone: cbdata.NeedRedstone,
					Conditional: cbdata.Conditional,
					Command: cbdata.Command,
					LastOutput: cbdata.LastOutput,
					Name: cbdata.CustomName,
					TickDelay: cbdata.TickDelay,
					ExecuteOnFirstTick: cbdata.ExecuteOnFirstTick,
				})
			}else if curblock.ChestSlot != nil {
				commands_generator.ReplaceItemRequest(request, curblock, cfg)
				cmdsender.SendSizukanaCommand(*request)
			}else{
				commands_generator.SetBlockRequest(request, curblock, cfg)
				err := cmdsender.SendSizukanaCommand(*request)
				if err != nil {
					panic(err)
				}
			}/*else if curblock.Entity != nil {
				//request := commands_generator.SummonRequest(curblock, cfg)
				//err := cmdsender.SendSizukanaCommand(request)
				//if err != nil {
				//	panic(err)
				//}
			}*/
			if dcfg.DelayMode==types.DelayModeContinuous {
				if isWindows{
					// the timer in windows is not the same as that in other system
					time.Sleep(time.Duration(dcfg.Delay/10) * time.Microsecond)
				}else{
					time.Sleep(time.Duration(dcfg.Delay) * time.Microsecond)
				}
			}else if dcfg.DelayMode==types.DelayModeDiscrete {
				tothresholdcounter++
				if tothresholdcounter>=dcfg.DelayThreshold {
					tothresholdcounter=0
					time.Sleep(time.Duration(dcfg.Delay) * time.Second)
				}
			}
		}
		commands_generator.FreeRequestStringPtr(request)
	} ()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				cmdsender.Output(fmt.Sprintf("[Task %d] Fatal error: %v", taskid, err))
				close(blockschannel)
			}
		} ()
		if task.Type==types.TaskTypeAsync {
			err := builder.Generate(cfg, asyncblockschannel)
			close(asyncblockschannel)
			if err != nil {
				cmdsender.Output(fmt.Sprintf("[%s %d] %s: %v",I18n.T(I18n.TaskTTeIuKoto), taskid,I18n.T(I18n.ERRORStr), err))
			}
			return
		}
		err := builder.Generate(cfg, blockschannel)
		close(blockschannel)
		if err != nil {
			cmdsender.Output(fmt.Sprintf("[%s %d] %s: %v",I18n.T(I18n.TaskTTeIuKoto), taskid,I18n.T(I18n.ERRORStr), err))
		}
	} ()
	return task
}

func InitTaskStatusDisplay(env *environment.PBEnvironment) {
	holder:=env.TaskHolder.(*TaskHolder)
	go func() {
		for {
			str:=<-holder.BrokSender
			env.CommandSender.Output(str)
		}
	} ()
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			<-ticker.C
			env.ActivateTaskStatus<-true
		}
	} ()
	go func() {
		for {
			<-env.ActivateTaskStatus
			if configuration.GlobalFullConfig(env).Global().TaskDisplayMode == types.TaskDisplayNo {
				continue
			}
			var displayStrs []string
			holder.TaskMap.Range(func (_tid interface{}, _v interface{}) bool {
				tid, _:=_tid.(int64)
				v, _:=_v.(*Task)

				addstr:=fmt.Sprintf("Task ID %d - %s - %s [%s]",tid,v.Config.Main().Execute,GetStateDesc(v.State),types.MakeTaskType(v.Type))
				if v.Type==types.TaskTypeAsync && v.State == TaskStateRunning {
					addstr=fmt.Sprintf("%s\nProgress: %s",addstr,ProgressThemes[0](&v.AsyncInfo))
				}
				displayStrs=append(displayStrs,addstr)
				commands_generator.AdditionalTitleCb(addstr)
				return true
			})
			displayStrs=append(displayStrs, holder.ExtraDisplayStrings...)
			if len(displayStrs) == 0 {
				continue
			}
			env.CommandSender.Title(strings.Join(displayStrs,"\n"))
		}
	} ()
}