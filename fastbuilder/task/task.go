package task
import (
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/builder"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/fastbuilder/i18n"
	"go.uber.org/atomic"
	"sync"
	"fmt"
	"time"
	"runtime"
	"strings"
	"github.com/google/uuid"
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
}

type AsyncInfo struct {
	Built int
	Total int
	BeginTime time.Time
}

var TaskIdCounter *atomic.Int64 = atomic.NewInt64(0)
var TaskMap sync.Map
var BrokSender chan string = make(chan string)
var ExtraDisplayStrings []string = []string{}

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
	TaskMap.Delete(task.TaskId)
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

func FindTask(taskId int64) *Task {
	t, _ := TaskMap.Load(taskId)
	ta, _ := t.(*Task)
	return ta
}

func CreateTask(commandLine string, conn *minecraft.Conn) *Task {
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig().Main())
	if err!=nil {
		command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskFailedToParseCommand),err))
		return nil
	}
	fcfg := configuration.ConcatFullConfig(cfg, configuration.GlobalFullConfig().Delay())
	dcfg := fcfg.Delay()
	/*if cfg.Execute == "" {
		return nil
	}
	Needless since it will be checked in function module.
	*/
	und, _ := uuid.NewUUID()
	command.SendWSCommand("gamemode c", und, conn)
	blockschannel := make(chan *types.Module, 10240)
	task := &Task {
		TaskId: TaskIdCounter.Add(1),
		CommandLine: commandLine,
		OutputChannel: blockschannel,
		State: TaskStateCalculating,
		Type: configuration.GlobalFullConfig().Global().TaskCreationType,
		Config: fcfg,
	}
	taskid := task.TaskId
	TaskMap.Store(taskid, task)
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
			for _, blk := range blocks {
				blockschannel <- blk
				task.AsyncInfo.Built++
			}
			close(blockschannel)
		} ()
	}else{
		task.State=TaskStateRunning
	}
	go func() {
		t1 := time.Now()
		blkscounter := 0
		tothresholdcounter := 0
		isFastMode := false
		if dcfg.DelayMode==types.DelayModeDiscrete||dcfg.DelayMode==types.DelayModeNone {
			isFastMode=true
		}else{
			//isFastMode=false
			command.SendWSCommand("gamemode c", und, conn)
			command.SendWSCommand("gamerule sendcommandfeedback true", und, conn)
		}
		for {
			task.ContinueLock.Lock()
			task.ContinueLock.Unlock()
			curblock, ok := <-blockschannel
			if !ok {
				if blkscounter == 0 {
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.Task_D_NothingGenerated),taskid))
					runtime.GC()
					task.Finalize()
					return
				}
				timeUsed := time.Now().Sub(t1)
				command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.Task_Summary_1), taskid, blkscounter))
				command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.Task_Summary_2), taskid, timeUsed.Seconds()))
				command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.Task_Summary_3), taskid, float64(blkscounter)/timeUsed.Seconds()))
				runtime.GC()
				task.Finalize()
				return
			}
			if blkscounter%20 == 0 {
				u_d, _ := uuid.NewUUID()
				command.SendWSCommand(fmt.Sprintf("tp %d %d %d",curblock.Point.X,curblock.Point.Y,curblock.Point.Z),u_d, conn)
			}
			blkscounter++
			if !cfg.ExcludeCommands && curblock.CommandBlockData != nil {
				if curblock.Block != nil {
					request:=command.SetBlockRequest(curblock, cfg)
					tuid, _ := uuid.NewUUID()
					if !isFastMode {
						wchan:=make(chan *packet.CommandOutput)
						command.UUIDMap.Store(tuid.String(),wchan)
						command.SendWSCommand(request,tuid, conn)
						<-wchan
					}else{
						command.SendWSCommand(request,tuid, conn)
					}
				}
				cbdata:=curblock.CommandBlockData
				if(cfg.InvalidateCommands){
					cbdata.Command="|"+cbdata.Command
				}
				u_d, _ := uuid.NewUUID()
				if !isFastMode {
					wchan:=make(chan *packet.CommandOutput)
					command.UUIDMap.Store(u_d.String(),wchan)
					command.SendWSCommand(fmt.Sprintf("tp %d %d %d",curblock.Point.X,curblock.Point.Y+1,curblock.Point.Z),u_d, conn)
					<-wchan
					close(wchan)
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
			}else if curblock.Block != nil || (curblock.CommandBlockData == nil&&curblock.Block != nil) {
				request := command.SetBlockRequest(curblock, cfg)
				err := command.SendSizukanaCommand(request, conn)
				if err != nil {
					panic(err)
				}
			}else if curblock.Entity != nil {
				//request := command.SummonRequest(curblock, cfg)
				//err := command.SendSizukanaCommand(request, conn)
				//if err != nil {
				//	panic(err)
				//}
			}
			if dcfg.DelayMode==types.DelayModeContinuous {
				time.Sleep(time.Duration(dcfg.Delay) * time.Microsecond)
			}else if dcfg.DelayMode==types.DelayModeDiscrete {
				tothresholdcounter++
				if tothresholdcounter>=dcfg.DelayThreshold {
					tothresholdcounter=0
					time.Sleep(time.Duration(dcfg.Delay) * time.Second)
				}
			}
		}
	} ()
	go func() {
		if task.Type==types.TaskTypeAsync {
			err := builder.Generate(cfg, asyncblockschannel)
			close(asyncblockschannel)
			if err != nil {
				command.Tellraw(conn, fmt.Sprintf("[%s %d] %s: %v",I18n.T(I18n.TaskTTeIuKoto), taskid,I18n.T(I18n.ERRORStr), err))
			}
			return
		}
		err := builder.Generate(cfg, blockschannel)
		close(blockschannel)
		if err != nil {
			command.Tellraw(conn, fmt.Sprintf("[%s %d] %s: %v",I18n.T(I18n.TaskTTeIuKoto), taskid,I18n.T(I18n.ERRORStr), err))
		}
	} ()
	return task
}

var ActivateTaskStatus chan bool=make(chan bool)

func InitTaskStatusDisplay(conn *minecraft.Conn) {
	go func() {
		for {
			str:=<-BrokSender
			command.Tellraw(conn,str)
		}
	} ()
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			<-ticker.C
			ActivateTaskStatus<-true
		}
	} ()
	go func() {
		for {
			<-ActivateTaskStatus
			//<- ticker.C
			if configuration.GlobalFullConfig().Global().TaskDisplayMode == types.TaskDisplayNo {
				continue
			}
			var displayStrs []string
			TaskMap.Range(func (_tid interface{}, _v interface{}) bool {
				tid, _:=_tid.(int64)
				v, _:=_v.(*Task)
				addstr:=fmt.Sprintf("Task ID %d - %s - %s [%s]",tid,v.Config.Main().Execute,GetStateDesc(v.State),types.MakeTaskType(v.Type))
				if v.Type==types.TaskTypeAsync && v.State == TaskStateRunning {
					addstr=fmt.Sprintf("%s\nProgress: %s",addstr,ProgressThemes[0](&v.AsyncInfo))
				}
				displayStrs=append(displayStrs,addstr)
				return true
			})
			displayStrs=append(displayStrs, ExtraDisplayStrings...)
			if len(displayStrs) == 0 {
				continue
			}
			command.Title(conn,strings.Join(displayStrs,"\n"))
		}
	} ()
}