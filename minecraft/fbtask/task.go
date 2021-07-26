package fbtask
import (
	"phoenixbuilder/minecraft/mctype"
	"phoenixbuilder/minecraft/parse"
	"phoenixbuilder/minecraft/builder"
	"phoenixbuilder/minecraft/command"
	"phoenixbuilder/minecraft/configuration"
	"phoenixbuilder/minecraft"
	"go.uber.org/atomic"
	"sync"
	"fmt"
	"time"
	"runtime"
	"github.com/google/uuid"
)

const (
	TaskStateUnknown = 0
	TaskStateRunning = 1
	TaskStatePaused  = 2
	TaskStateDied    = 3
)

type Task struct {
	TaskId int64
	CommandLine string
	OutputChannel chan *mctype.Module
	ContinueLock sync.Mutex
	State byte
	Config *configuration.FullConfig
}

var TaskIdCounter *atomic.Int64 = atomic.NewInt64(0)
var TaskMap sync.Map

func GetStateDesc(st byte) string {
	if st == 0 {
		return "Unknown"
	}else if st==1 {
		return "Running"
	}else if st==2 {
		return "Paused"
	}else if st==3 {
		return "Died"
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
	task.State = TaskStateRunning
	task.ContinueLock.Unlock()
}

func (task *Task) Break() {
	if task.State != TaskStatePaused {
		task.Pause()
	}
	if task.State == TaskStateDied {
		return
	}
	chann := task.OutputChannel
	for {
		blk, ok := <- chann
		if !ok {
			break
		}
		if false {
			fmt.Printf("%v\n",blk)
		}
	}
	task.Resume()
}

func FindTask(taskId int64) *Task {
	t, _ := TaskMap.Load(taskId)
	ta, _ := t.(*Task)
	return ta
}

func CreateTask(commandLine string, conn *minecraft.Conn) *Task {
	cfg := parse.Parse(commandLine, configuration.GlobalFullConfig().Main())
	fcfg := configuration.ConcatFullConfig(cfg, configuration.GlobalFullConfig().Delay())
	dcfg := fcfg.Delay()
	if cfg.Execute == "" {
		return nil
	}
	blockschannel := make(chan *mctype.Module, 10240)
	task := &Task {
		TaskId: TaskIdCounter.Add(1),
		CommandLine: commandLine,
		OutputChannel: blockschannel,
		State: TaskStateRunning,
		Config: fcfg,
	}
	taskid := task.TaskId
	TaskMap.Store(taskid, task)
	go func() {
		t1 := time.Now()
		blkscounter := 0
		tothresholdcounter := 0
		for {
			task.ContinueLock.Lock()
			task.ContinueLock.Unlock()
			curblock, ok := <-blockschannel
			if !ok {
				if blkscounter == 0 {
					command.Tellraw(conn, fmt.Sprintf("[Task %d] Nothing generated.",taskid))
					runtime.GC()
					task.Finalize()
					return
				}
				timeUsed := time.Now().Sub(t1)
				command.Tellraw(conn, fmt.Sprintf("[Task %d] %v block(s) have been changed.", taskid, blkscounter))
				command.Tellraw(conn, fmt.Sprintf("[Task %d] Time used: %v second(s)", taskid, timeUsed.Seconds()))
				command.Tellraw(conn, fmt.Sprintf("[Task %d] Average speed: %v blocks/second", taskid, float64(blkscounter)/timeUsed.Seconds()))
				runtime.GC()
				task.Finalize()
				return
			}
			blkscounter++
			request := command.SetBlockRequest(curblock, cfg)
			uuid1, _ := uuid.NewUUID()
			err := command.SendCommand(request, uuid1, conn)
			if err != nil {
				panic(err)
			}
			if dcfg.DelayMode==mctype.DelayModeContinuous {
				time.Sleep(time.Duration(dcfg.Delay) * time.Microsecond)
			}else if dcfg.DelayMode==mctype.DelayModeDiscrete {
				tothresholdcounter++
				if tothresholdcounter>=dcfg.DelayThreshold {
					tothresholdcounter=0
					time.Sleep(time.Duration(dcfg.Delay) * time.Second)
				}
			}
		}
	} ()
	go func() {
		err := builder.Generate(cfg, blockschannel)
		close(blockschannel)
		if err != nil {
			command.Tellraw(conn, fmt.Sprintf("[Task %d] Error: %v", taskid, err))
		}
	} ()
	return task
}