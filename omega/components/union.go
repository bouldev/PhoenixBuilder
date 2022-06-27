package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"sync"
	"time"
)

type UnionMember struct {
	Name  string `json:"成员名"`
	Uid   string `json:"UUID"`
	Level int    `json:"等级"`
}

const (
	UnionMemberLevelCharMan = iota + 1
	UnionMemberLevelViceCharMan
	UnionMemberLevelMember
)

type UnionPlayerGatherInfo struct {
	Name             string
	Uid              string
	IsOperator       bool
	Union            *UnionData
	UnionMemberLevel int
	UnionMemberInfo  *UnionMember
}

type UnionData struct {
	Name         string                  `json:"公会名称"`
	Chairman     *UnionMember            `json:"会长"`
	ViceChairMan map[string]*UnionMember `json:"副会长"`
	Members      map[string]*UnionMember `json:"所有成员"`
}

type unionFileData struct {
	DeferedCmds map[string]*defines.CmdsWithName `json:"下次上线时执行的指令"`
	Unions      []*UnionData                     `json:"工会信息"`
}

type UnionDisplayConfig struct {
}

type Union struct {
	*BasicComponent
	fileData      *unionFileData
	fileChange    bool
	FileName      string              `json:"数据文件"`
	LoginDelay    int                 `json:"登录时延迟发送"`
	Triggers      []string            `json:"触发词"`
	Usage         string              `json:"提示信息"`
	Operator      []string            `json:"超管"`
	DisplayConfig *UnionDisplayConfig `json:"显示配置"`
	mu            sync.RWMutex
}

func (o *Union) executeCmds(player string, cmds []defines.Cmd) (success bool) {
	resultChan := make(chan bool)
	utils.GetPlayerList(o.Frame.GetGameControl(), "@a[name=\""+player+"\"]", func(s []string) {
		if len(s) == 0 {
			resultChan <- false
		} else {
			utils.LaunchCmdsArray(o.Frame.GetGameControl(), cmds, map[string]interface{}{
				"[player]": player,
			}, o.Frame.GetBackendDisplay())
			resultChan <- true
		}
	})
	return <-resultChan
}
func (o *Union) executeCmdsWithDefer(player, uid string, cmds []defines.Cmd) {
	go func() {
		if o.executeCmds(player, cmds) {
		} else {
			o.mu.Lock()
			if cw, hasK := o.fileData.DeferedCmds[uid]; hasK {
				cw.Cmds = append(cw.Cmds, cmds...)
			} else {
				o.fileData.DeferedCmds[uid] = &defines.CmdsWithName{Name: player, Cmds: cmds}
			}
			o.mu.Unlock()
			o.fileChange = true
		}
	}()
}

func (o *Union) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	o.mu = sync.RWMutex{}
}

func (o *Union) getUnionPlayerGatherInfo(player, uid string) *UnionPlayerGatherInfo {
	inf := &UnionPlayerGatherInfo{}
	inf.Name = player
	inf.Uid = uid
	for _, o := range o.Operator {
		if player == o {
			inf.IsOperator = true
		}
	}
	for _, union := range o.fileData.Unions {
		if union.Chairman.Uid == uid {
			inf.Union = union
			inf.UnionMemberLevel = UnionMemberLevelCharMan
			inf.UnionMemberInfo = union.Chairman
			break
		}
		if member, hasK := union.ViceChairMan[uid]; hasK {
			inf.Union = union
			inf.UnionMemberLevel = UnionMemberLevelViceCharMan
			inf.UnionMemberInfo = member
			break
		}
		if member, hasK := union.Members[uid]; hasK {
			inf.Union = union
			inf.UnionMemberLevel = UnionMemberLevelMember
			inf.UnionMemberInfo = member
			break
		}
	}
	return inf
}
func (o *Union) popMenu(chat *defines.GameChat) (stop bool) {
	stop = true
	pK := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	uq := pK.GetRelatedUQ()
	if uq == nil {
		return
	}
	info := o.getUnionPlayerGatherInfo(chat.Name, uq.UUID.String())
	fmtS := fmt.Sprintf("玩家: %v 身份: ", info.Name)
	if info.IsOperator {
		fmtS += "管理员 "
	}
	if info.Union != nil {
		fmtS += fmt.Sprintf("公会<%v>", info.Union.Name)
		switch info.UnionMemberLevel {
		case UnionMemberLevelCharMan:
			fmtS += "会长"
		case UnionMemberLevelViceCharMan:
			fmtS += "副会长"
		case UnionMemberLevelMember:
			fmtS += "成员"
		}
	} else {
		fmtS += "未加入任何公会"
	}
	return
}

func (o *Union) Inject(frame defines.MainFrame) {
	o.Frame = frame
	err := frame.GetJsonData(o.FileName, &o.fileData)
	if err != nil {
		panic(err)
	}
	o.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		name := utils.ToPlainName(entry.Username)
		uid := entry.UUID.String()
		o.mu.RLock()
		if deferCmds, hasK := o.fileData.DeferedCmds[uid]; hasK {
			o.mu.RUnlock()
			timer := time.NewTimer(time.Duration(o.LoginDelay) * time.Second)
			go func() {
				<-timer.C
				if o.executeCmds(name, deferCmds.Cmds) {
					o.mu.Lock()
					delete(o.fileData.DeferedCmds, uid)
					o.fileChange = true
					o.mu.Unlock()
				}
			}()
			for _, union := range o.fileData.Unions {
				if union.Chairman.Uid == uid {
					if union.Chairman.Name != name {
						union.Chairman.Name = name
						o.fileChange = true
					}
					break
				}
				if member, hasK := union.ViceChairMan[uid]; hasK {
					if member.Name != name {
						member.Name = name
						o.fileChange = true
					}
					break
				}
				if member, hasK := union.Members[uid]; hasK {
					if member.Name != name {
						member.Name = name
						o.fileChange = true
					}
					break
				}
			}
		} else {
			o.mu.RUnlock()
		}
	})
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[子选项]",
			Usage:        o.Usage,
			FinalTrigger: false,
		},
		OptionalOnTriggerFn: o.popMenu,
	})
}

func (o *Union) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.fileData)
		}
	}
	return nil
}

func (o *Union) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.fileData)
}
