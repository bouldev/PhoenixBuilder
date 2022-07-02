package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"
)

type StructureEntry struct {
	Time      string `json:"time"`
	RealIDX   int    `json:"idx"`
	StartPos  []int  `json:"start"`
	EndPos    []int  `json:"end"`
	CenterPos []int  `json:"center"`
}

type SystemRecord struct {
	Count int `json:"总计占用数量"`
}

type StructureRecords struct {
	User   map[string]map[string]*StructureEntry `json:"玩家备份记录"`
	System SystemRecord                          `json:"系统记录"`
}

type StructureBackup struct {
	*defines.BasicComponent
	BackupTriggers        []string `json:"备份触发词"`
	RecoverTriggers       []string `json:"恢复触发词"`
	CoolDownSecond        int      `json:"请求冷却时间"`
	MaxStructureBackupNum int      `json:"最大备份数量"`
	BackupSelector        string   `json:"选择器"`
	BackupSize            int      `json:"建筑备份长宽"`
	BackupHeight          int      `json:"建筑备份高度"`
	Admin                 []string `json:"建筑恢复管理员"`
	lastRequestTime       map[string]time.Time
	Structures            *StructureRecords
	fileChange            bool
	FileName              string `json:"存档点记录文件名"`
}

func (o *StructureBackup) formatStructureName(idx int) string {
	return fmt.Sprintf("OMBKIDpt2%v", idx)
}

func (o *StructureBackup) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	o.lastRequestTime = map[string]time.Time{}
	if o.BackupSize < 2 || o.BackupSize > 64 {
		panic(fmt.Errorf("建筑备份长宽 只能是 2～63 间的整数"))
	}
	if o.BackupSize < 2 || o.BackupHeight > 252 {
		panic(fmt.Errorf("建筑备份高度只能是 2～252 间的整数"))
	}
}

func (o *StructureBackup) doBackup(user, structureName string, pos []int, idx int) {
	if idx == 0 {
		o.Structures.System.Count++
		idx = o.Structures.System.Count
	}
	cmdName := o.formatStructureName(idx)
	halfS := o.BackupSize / 2
	halfH := o.BackupHeight / 2
	startPos := []int{pos[0] - halfS, pos[1] - halfH, pos[2] - halfS}
	endPos := []int{pos[0] + halfS, pos[1] + halfH, pos[2] + halfS}
	if startPos[1] < 1 || startPos[1] > 254 || endPos[1] < 1 || endPos[1] > 254 {
		allL, allH := 1+halfH, 254-halfH
		o.Frame.GetGameControl().SayTo(user, fmt.Sprintf("不能备份...你必须位于高度 %v ~ %v 之间", allL, allH))
		return
	}
	o.Structures.User[user][structureName] = &StructureEntry{
		Time:      utils.TimeToString(time.Now()),
		RealIDX:   idx,
		StartPos:  startPos,
		EndPos:    endPos,
		CenterPos: pos,
	}
	cmd := fmt.Sprintf("structure save %v %v %v %v %v %v %v true disk ", cmdName, startPos[0], startPos[1], startPos[2], endPos[0], endPos[1], endPos[2])
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		pk := o.Frame.GetGameControl().GetPlayerKit(user)
		if output.SuccessCount > 0 {
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("备份成功 %v@%v:%v %v->%v", user, structureName, o.Structures.User[user][structureName], cmd, output))
			pk.Say("备份成功")
			o.lastRequestTime[user] = time.Now()
			o.fileChange = true
		} else {
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("备份失败 %v@%v:%v %v->%v", user, structureName, o.Structures.User[user][structureName], cmd, output))
			delete(o.Structures.User[user], structureName)
			pk.Say("备份失败")
			o.fileChange = true
		}
	})
}

func (o *StructureBackup) requestPos(user, structureName string, idx int) {
	pk := o.Frame.GetGameControl().GetPlayerKit(user)
	go func() {
		pos := <-pk.GetPos(o.BackupSelector)
		if pos == nil {
			pk.Say("备份失败...该区域可能不能备份")
			return
		}
		if _, hasK := o.Structures.User[user]; !hasK {
			o.Structures.User[user] = map[string]*StructureEntry{}
		}
		o.doBackup(user, structureName, pos, idx)
	}()
}

func (o *StructureBackup) getStructureName(user string, args []string, idx int) {
	if len(args) > 0 {
		o.requestPos(user, args[0], idx)
	} else {
		pk := o.Frame.GetGameControl().GetPlayerKit(user)
		if pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				o.requestPos(chat.Name, chat.Msg[0], idx)
			}
			return true
		}) == nil {
			pk.Say("请输入这个地点的名字:")
		}
	}
}

func (o *StructureBackup) tryBackup(chat *defines.GameChat) bool {
	if t, ok := o.lastRequestTime[chat.Name]; ok {
		if t.Add(time.Duration(o.CoolDownSecond) * time.Second).After(time.Now()) {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("你请求的太频繁了,%v秒以后再试吧", int(t.Add(time.Duration(o.CoolDownSecond)*time.Second).Sub(time.Now()).Seconds())))
			return true
		}
	}
	if _, ok := o.Structures.User[chat.Name]; !ok {
		o.Structures.User[chat.Name] = map[string]*StructureEntry{}
	}
	if len(o.Structures.User[chat.Name]) >= o.MaxStructureBackupNum {
		structures := []*StructureEntry{}
		sname := []string{}
		for n, s := range o.Structures.User[chat.Name] {
			sname = append(sname, n)
			structures = append(structures, s)
		}
		hint, resolver := utils.GenStringListHintResolverWithIndex(sname)
		if o.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
			i, b, err := resolver(chat.Msg)
			if b {
				o.Frame.GetGameControl().SayTo(chat.Name, "操作已取消")
				return true
			}
			if err != nil {
				o.Frame.GetGameControl().SayTo(chat.Name, "无法理解的输入，因为"+err.Error())
				return true
			}
			n := sname[i]
			if o.Structures.User[chat.Name] != nil && o.Structures.User[chat.Name][n] != nil {
				delete(o.Structures.User[chat.Name], n)
				o.fileChange = true
			}
			o.getStructureName(chat.Name, []string{}, structures[i].RealIDX)
			return true
		}) == nil {
			for i, name := range sname {
				o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("%d: %v", i+1, name))
			}
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("备份数量已满，请选择一个要丢弃的备份，或者取消备份:\n"+hint))
		}
		return true
	}
	o.getStructureName(chat.Name, chat.Msg, 0)
	return true
}
func (o *StructureBackup) doRecovery(user string, admin string, s *StructureEntry, sname string) {
	recoverCmd := fmt.Sprintf("execute %v ~~~ structure load %v %v %v %v 0_degrees none true true", user, o.formatStructureName(s.RealIDX), s.StartPos[0], s.StartPos[1], s.StartPos[2])
	o.Frame.GetBackendDisplay().Write(recoverCmd)
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("tp \"%v\" %v %v %v", user, s.CenterPos[0], s.CenterPos[1], s.CenterPos[2]), func(output *packet.CommandOutput) {
		go func() {
			<-time.NewTimer(time.Second).C
			o.Frame.GetGameControl().SendCmd(recoverCmd)
			msg := fmt.Sprintf("尝试将 %v 的建筑 %v 恢复于 %v %v %v, 但是，不一定会成功", user, sname, s.CenterPos[0], s.CenterPos[1], s.CenterPos[2])
			o.Frame.GetGameControl().SayTo(admin, msg)
			o.Frame.GetGameControl().SayTo(user, msg)
			o.Frame.GetGameControl().SayTo(admin, "不成功请找腐竹手动恢复")
			o.Frame.GetGameControl().SayTo(user, "不成功请找腐竹手动恢复")
			o.Frame.GetBackendDisplay().Write(msg)
			<-time.NewTimer(time.Second * 3).C
			o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp \"%v\" %v %v %v", user, s.CenterPos[0], s.CenterPos[1], s.CenterPos[2]))
			o.lastRequestTime[user] = time.Now()
		}()
	})

}
func (o *StructureBackup) askForAuth(user string, admins []string, s *StructureEntry, sname string) {
	o.lastRequestTime[user] = time.Now()
	approved := false
	for _, _a := range admins {
		admin := _a
		hint, resolver := utils.GenYesNoResolver()
		if o.Frame.GetGameControl().SetOnParamMsg(admin, func(chat *defines.GameChat) (catch bool) {
			y, err := resolver(chat.Msg)
			if err != nil {
				o.Frame.GetGameControl().SayTo(chat.Name, "已弃权")
				o.Frame.GetGameControl().SayTo(user, admin+"已弃权")
				return true
			}
			if !y {
				o.Frame.GetGameControl().SayTo(chat.Name, "已拒绝")
				o.Frame.GetGameControl().SayTo(user, admin+"已拒绝")
				return true
			} else {
				o.Frame.GetGameControl().SayTo(chat.Name, "同意了请求")
				o.Frame.GetGameControl().SayTo(user, admin+"同意了请求")
				o.Frame.GetBackendDisplay().Write(fmt.Sprintf(admin+"同意了 %v 恢复中心位于 %v %v %v 的建筑 %v 的请求", user, s.CenterPos[0], s.CenterPos[1], s.CenterPos[2], sname))
				if approved {
					return true
				}
				approved = true
				o.doRecovery(user, admin, s, sname)
			}
			return true
		}) == nil {
			o.Frame.GetGameControl().SayTo(user, fmt.Sprintf("等待管理员 %v 同意恢复请求", strings.Join(admins, " 或 ")))
			o.Frame.GetGameControl().SayTo(admin, fmt.Sprintf("%v 请求恢复中心位于 %v %v %v 的建筑 %v, 要同意吗? 输入 "+hint+": ", user, s.CenterPos[0], s.CenterPos[1], s.CenterPos[2], sname))
		}
	}
}

func (o *StructureBackup) tryRecover(chat *defines.GameChat) bool {
	if t, ok := o.lastRequestTime[chat.Name]; ok {
		if t.Add(time.Duration(o.CoolDownSecond) * time.Second).After(time.Now()) {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("你请求的太频繁了,%v秒以后再试吧", int(t.Add(time.Duration(o.CoolDownSecond)*time.Second).Sub(time.Now()).Seconds())))
			return true
		}
	}
	admins := []string{}
	for _, p := range o.Frame.GetUQHolder().PlayersByEntityID {
		for _, op := range o.Admin {
			if op == p.Username {
				admins = append(admins, op)
			}
		}
	}
	if len(admins) == 0 {
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("建筑备份管理员不在线...把 %v 叫上线再试吧", strings.Join(o.Admin, " 或 ")))
		return true
	}
	structures := []*StructureEntry{}
	sname := []string{}
	for n, s := range o.Structures.User[chat.Name] {
		sname = append(sname, n)
		structures = append(structures, s)
	}
	hint, resolver := utils.GenStringListHintResolverWithIndex(sname)
	if o.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
		i, b, err := resolver(chat.Msg)
		if b {
			o.Frame.GetGameControl().SayTo(chat.Name, "操作已取消")
			return true
		}
		if err != nil {
			o.Frame.GetGameControl().SayTo(chat.Name, "无法理解的输入，因为"+err.Error())
			return true
		}
		o.askForAuth(chat.Name, admins, structures[i], sname[i])
		return true
	}) == nil {
		for i, name := range sname {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("%d: %v", i+1, name))
		}
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("选择一个备份以恢复，或者取消恢复:\n"+hint))
	}
	return true
}

func (o *StructureBackup) StructureBackup(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", &o.Structures)
		}
	}
	return nil
}

func (o *StructureBackup) Stop() error {
	fmt.Println("正在保存 " + o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", &o.Structures)
}

func (o *StructureBackup) Inject(frame defines.MainFrame) {
	o.Frame = frame
	err := frame.GetJsonData(o.FileName, &o.Structures)
	if err != nil {
		panic(err)
	}
	if o.Structures == nil || o.Structures.User == nil || len(o.Structures.User) == 0 {
		o.Structures = &StructureRecords{
			User: map[string]map[string]*StructureEntry{},
			System: SystemRecord{
				0,
			},
		}
	}
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.BackupTriggers,
			ArgumentHint: "[地点名]",
			FinalTrigger: false,
			Usage:        fmt.Sprintf("备份以你为中心的长宽%v,高%v的区域", o.BackupSize, o.BackupHeight),
		},
		OptionalOnTriggerFn: o.tryBackup,
	})
	frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.RecoverTriggers,
			ArgumentHint: "[地点名]",
			FinalTrigger: false,
			Usage:        "申请恢复一个备份",
		},
		OptionalOnTriggerFn: o.tryRecover,
	})
}
