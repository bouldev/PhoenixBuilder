package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

type UIDRecord struct {
	Uid  int    `json:"uid"`
	Name string `json:"玩家名"`
}

type UIDTracking struct {
	*BasicComponent
	FileName           string        `json:"记录文件"`
	PlayerUIDFetchCmd  string        `json:"获取某玩家的uid的指令"`
	AuxUidAssign       bool          `json:"是否让Omega完成uid分配"`
	UidFetchCmd        string        `json:"让Omega分配uid时获取最后一个被分配的uid的指令"`
	UidAsignCmd        string        `json:"让Omega分配uid时为某玩家分配uid的指令"`
	cmdsBeforeUidAsign []defines.Cmd `json:"为某玩家分配uid前的准备工作"`
	cmdsAfterUidAsign  []defines.Cmd `json:"为某玩家分配uid的后续工作"`
	Delay              int           `json:"玩家上线的延迟时间"`
	DiskUUIDs          map[string]*UIDRecord
	AllUids            map[uuid.UUID]int
	IsTracking         map[int64]bool
	mu                 sync.Mutex
}

func (o *UIDTracking) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	var err error
	if o.cmdsBeforeUidAsign, err = utils.ParseAdaptiveJsonCmd(cfg.Configs,
		[]string{"为某玩家分配uid前的准备工作"}); err != nil {
		panic(err)
	}
	if o.cmdsAfterUidAsign, err = utils.ParseAdaptiveJsonCmd(cfg.Configs,
		[]string{"为某玩家分配uid的后续工作"}); err != nil {
		panic(err)
	}
	o.mu = sync.Mutex{}
	o.IsTracking = make(map[int64]bool)
	o.DiskUUIDs = map[string]*UIDRecord{}
	o.AllUids = make(map[uuid.UUID]int)
}

func (o *UIDTracking) loadTracking() {
	if err := o.Frame.GetJsonData(o.FileName, &o.DiskUUIDs); err != nil {
		panic(err)
	}
	if o.DiskUUIDs == nil {
		o.DiskUUIDs = make(map[string]*UIDRecord)
	}
	o.AllUids = map[uuid.UUID]int{}
	for uuidStr, r := range o.DiskUUIDs {
		if UUID, err := uuid.Parse(uuidStr); err != nil {
			fmt.Println(err)
		} else {
			o.AllUids[UUID] = r.Uid
		}
	}
}

func (o *UIDTracking) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.loadTracking()
	o.Frame.GetGameListener().AppendLoginInfoCallback(o.onNewPlayer)
	o.Frame.GetGameListener().AppendLogoutInfoCallback(func(entry protocol.PlayerListEntry) {
		o.mu.Lock()
		defer o.mu.Unlock()
		if _, hasK := o.IsTracking[entry.EntityUniqueID]; hasK {
			// fmt.Println("Remove Tracking", entry.EntityUniqueID)
			delete(o.IsTracking, entry.EntityUniqueID)
		}
	})
}

func (o *UIDTracking) Stop() error {
	fmt.Println("正在保存 ", o.FileName)
	return o.Frame.WriteJsonData(o.FileName, o.DiskUUIDs)
}

func (o *UIDTracking) CommitUID(name string, UUID uuid.UUID, uid int) {
	o.DiskUUIDs[UUID.String()] = &UIDRecord{
		Uid:  uid,
		Name: name,
	}
	o.AllUids[UUID] = uid
}

func (o *UIDTracking) RequestPlayerUID(name string) int {
	reqCmd := utils.FormatByReplacingOccurrences(o.PlayerUIDFetchCmd, map[string]interface{}{"[player]": "\"" + name + "\""})
	resultWaitor := make(chan int)
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(reqCmd, func(output *packet.CommandOutput) {
		// fmt.Println(output)
		if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 4 {
			pterm.Error.Println("执行指令 %v 出现错误 %v, 请确认指令/计分板名无误", reqCmd, output.OutputMessages)
			resultWaitor <- 0
			return
		} else {
			newVal, err := strconv.Atoi(output.OutputMessages[0].Parameters[3])
			if err != nil {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("解析UID时出错 %v", err))
				resultWaitor <- 0
				return
			}
			resultWaitor <- newVal
		}
	})
	r := <-resultWaitor
	// fmt.Println(r)
	return r
}

func (o *UIDTracking) doAssignPlayerUID(currentUID int, name string, UUID uuid.UUID) {

	replacement := map[string]interface{}{
		"[player]": "\"" + name + "\"",
		"[uid]":    currentUID,
		"[uid+1]":  currentUID + 1,
	}
	// pterm.Info.Println(replacement)
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.cmdsBeforeUidAsign, replacement, o.Frame.GetBackendDisplay())
	assignCmd := utils.FormatByReplacingOccurrences(o.UidAsignCmd, replacement)
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(assignCmd, func(output *packet.CommandOutput) {
		// fmt.Println(output)
		if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 3 {
			pterm.Error.Printf("执行指令 %v 出现错误 %v, 无法为玩家分配 uid\n", assignCmd, output)
			return
		} else {
			uid, err := strconv.Atoi(output.OutputMessages[0].Parameters[2])
			if err != nil {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("解析下一个待分配 uid时出错 %v", err))
				return
			}
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("新玩家 UID 分配记录 %v %v %v", name, UUID.String(), uid))
			o.CommitUID(name, UUID, uid)
			go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.cmdsAfterUidAsign, replacement, o.Frame.GetBackendDisplay())
		}
	})

}

func (o *UIDTracking) AssignPlayerUID(name string, UUID uuid.UUID, uid int) {
	o.Frame.GetBackendDisplay().Write(fmt.Sprintf("开始为玩家 %v %v %v 分配 UID", name, UUID.String(), uid))
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(o.UidFetchCmd, func(output *packet.CommandOutput) {
		fmt.Println(output)
		if output.SuccessCount == 0 || len(output.OutputMessages) == 0 || len(output.OutputMessages[0].Parameters) != 4 {
			pterm.Error.Printfln("执行指令 %v 出现错误 %v, 无法确认下一个待分配 uid 请确认指令/计分板名无误", o.UidFetchCmd, output.OutputMessages)
			return
		} else {
			// pterm.Info.Println(output.OutputMessages[0].Parameters)
			newVal, err := strconv.Atoi(output.OutputMessages[0].Parameters[3])
			if err != nil {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("解析下一个待分配 uid时出错 %v", err))
				return
			}
			o.doAssignPlayerUID(newVal, name, UUID)
		}
	})
}

func (o *UIDTracking) onNewPlayer(entry protocol.PlayerListEntry) {
	if uid, hasK := o.AllUids[entry.UUID]; hasK {
		o.CommitUID(entry.Username, entry.UUID, uid)
		return
	}
	go func() {
		t := time.NewTimer(time.Second * time.Duration(o.Delay))
		<-t.C
		if player, hasK := o.Frame.GetUQHolder().PlayersByEntityID[entry.EntityUniqueID]; !hasK {
			return
		} else {
			UUID := player.UUID
			Name := entry.Username
			uid := o.RequestPlayerUID(Name)
			if uid != 0 {
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf("新玩家 UID 记录 %v %v %v", Name, UUID.String(), uid))
				o.CommitUID(Name, UUID, uid)
				return
			}
			if o.AuxUidAssign {
				o.AssignPlayerUID(Name, UUID, uid)
				return
			} else {
				o.mu.Lock()
				defer o.mu.Unlock()
				if o.IsTracking[entry.EntityUniqueID] {
					return
				}
				// fmt.Println("stark tracking ", entry.EntityUniqueID)
				o.IsTracking[entry.EntityUniqueID] = true
				// fmt.Println("Trcking " + entry.Username)
				trackerC := time.NewTicker(10 * time.Second)
				for {
					// fmt.Println("Polling " + entry.Username)
					<-trackerC.C
					if player, hasK := o.Frame.GetUQHolder().PlayersByEntityID[entry.EntityUniqueID]; hasK {
						UUID := player.UUID
						Name := entry.Username
						uid := o.RequestPlayerUID(Name)
						if uid != 0 {
							o.CommitUID(Name, UUID, uid)
							return
						}
					} else {
						// fmt.Println("Stop Trcking " + entry.Username)
						if _, hasK := o.IsTracking[entry.EntityUniqueID]; hasK {
							// fmt.Println("Remove Tracking", entry.EntityUniqueID)
							delete(o.IsTracking, entry.EntityUniqueID)
						}
						return
					}
				}
			}
		}
	}()
}
