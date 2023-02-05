package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/global"
	"phoenixbuilder/omega/utils"
	"sort"
	"time"

	"github.com/pterm/pterm"
)

type scoreRecord struct {
	Name  string `json:"玩家名"`
	UUID  string `json:"uuid"`
	Score int    `json:"分数"`
	Rank  int    `json:"排名"`
}

type scoreRecords struct {
	ascending bool
	records   []*scoreRecord
}

type RankRenderOption struct {
	MaxCount          int               `json:"最大显示的人数"`
	DefaultRenderFmt  string            `json:"默认渲染样式"`
	SpecificRenderFmt map[string]string `json:"特定排名渲染样式"`
	PlayerRenderFmt   string            `json:"查询者渲染样式"`
	HintOnNoPlayer    string            `json:"当查询者没有相关分数时显示"`
	Head              string            `json:"开头"`
	Tail              string            `json:"结尾"`
	Split             string            `json:"隔断"`
	// RenderScoreBoard   bool              `json:"渲染计分板"`
	// ClearScoreBoardCmd string            `json:"清除计分板指令"`
	// ScoreBoardSetFmt   string            `json:"计分板渲染指令"`
}

type RankScoreboardRenderOption struct {
	EnableScoreboardRender  bool   `json:"启用计分板渲染"`
	TargetDisplayScoreboard string `json:"渲染计分板名[数据可能会被清除]"`
	ScoreboardRenderName    string `json:"计分板项目名渲染方式"`
	ScoreboardRenderScoreBy string `json:"计分板项目分数[序号升序/序号降序/实际分数]"`
	RenderOnlinePlayersOnly bool   `json:"仅仅渲染在线玩家的计分板"`
}
type RankingAuxData struct {
	MaintainedNames map[string]int
}

type Ranking struct {
	*defines.BasicComponent
	Triggers         []string                    `json:"触发词"`
	Usage            string                      `json:"提示信息"`
	ScoreboardName   string                      `json:"计分板名"`
	FileName         string                      `json:"排名记录文件"`
	MaxSaveCount     int                         `json:"最多保存多少记录在文件中"`
	Ascending        bool                        `json:"升序"`
	Period           int                         `json:"刷新周期"`
	Render           *RankRenderOption           `json:"渲染选项"`
	ScoreBoardRender *RankScoreboardRenderOption `json:"计分板渲染选项"`
	records          *scoreRecords
	fileChange       bool
	auxData          *RankingAuxData
	auxDataChanged   bool
	playerMapping    map[string]*scoreRecord
	// scoreboardRenderCache []string
}

func (ss *scoreRecords) Len() int { return len(ss.records) }
func (ss *scoreRecords) Less(i, j int) bool {
	if ss.ascending {
		return ss.records[i].Score > ss.records[j].Score
	} else {
		return ss.records[i].Score < ss.records[j].Score
	}
}
func (ss *scoreRecords) Swap(i, j int) {
	t := ss.records[i]
	ss.records[i] = ss.records[j]
	ss.records[j] = t
}

func (ss *scoreRecords) freshOrder() {
	for i, r := range ss.records {
		r.Rank = i + 1
	}
}

func (o *Ranking) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	if cfg.Version == "0.0.1" {
		delete((cfg.Configs["渲染选项"]).(map[string]interface{}), "渲染计分板")
		delete((cfg.Configs["渲染选项"]).(map[string]interface{}), "清除计分板指令")
		delete((cfg.Configs["渲染选项"]).(map[string]interface{}), "计分板渲染指令")
		cfg.Configs["计分板渲染选项"] = map[string]interface{}{
			"启用计分板渲染":                 true,
			"渲染计分板名[数据可能会被清除]":        "时间显示",
			"计分板项目名渲染方式":              "§6[I++].§a[player]",
			"计分板项目分数[序号升序/序号降序/实际分数]": "实际分数",
			"仅仅渲染在线玩家的计分板":            true,
		}
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.auxData = &RankingAuxData{
		MaintainedNames: map[string]int{},
	}
	o.records = &scoreRecords{
		ascending: o.Ascending,
	}
	// o.scoreboardRenderCache = []string{}
	if o.ScoreBoardRender.EnableScoreboardRender && o.ScoreBoardRender.TargetDisplayScoreboard == "" {
		panic("如果启用计分板渲染，则必须设置渲染计分板名")
	}
}

func (o *Ranking) onTrigger(chat *defines.GameChat) (stop bool) {
	stop = true
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	pk.Say(o.Render.Head)
	for _i, r := range o.records.records {
		if _i == o.Render.MaxCount {
			break
		}
		fmtStr := o.Render.DefaultRenderFmt
		if _f, hasK := o.Render.SpecificRenderFmt[fmt.Sprintf("%v", r.Rank)]; hasK {
			fmtStr = _f
		}
		text := utils.FormatByReplacingOccurrences(fmtStr, map[string]interface{}{
			"[i]":      r.Rank,
			"[player]": "\"" + r.Name + "\"",
			"[score]":  r.Score,
		})
		pk.Say(text)
	}
	pk.Say(o.Render.Split)
	uidStr := pk.GetRelatedUQ().UUID.String()
	if pr, hasK := o.playerMapping[uidStr]; hasK {
		fmt := o.Render.PlayerRenderFmt
		text := utils.FormatByReplacingOccurrences(fmt, map[string]interface{}{
			"[i]":      pr.Rank,
			"[player]": "\"" + pr.Name + "\"",
			"[score]":  pr.Score,
		})
		pk.Say(text)
	} else {
		pk.Say(o.Render.HintOnNoPlayer)
	}
	pk.Say(o.Render.Tail)
	return
}

func (o *Ranking) update(rankingLastFetchResult map[string]map[string]int) {
	// pterm.Info.Println(rankingLastFetchResult)
	if players, hasK := rankingLastFetchResult[o.ScoreboardName]; !hasK {
		pterm.Error.Printfln("没有计分板 %v,所有的计分板被列在下方,如果有计分板但还是出现这个错误，可能是因为没有一个玩家在这个计分板上有分数\n如果你不需要排行榜功能，可以去 配置/组件-排行榜.json 禁用这个功能以摆脱这个错误", o.ScoreboardName)
		for n, _ := range rankingLastFetchResult {
			pterm.Error.Println(n)
		}
		o.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreboard players add @s %v 0", o.ScoreboardName))
	} else {
		onlineRecords := &scoreRecords{ascending: o.Ascending, records: make([]*scoreRecord, 0)}
		needSort := false
		needRankUpdate := false
		for player, score := range players {
			for _, rp := range o.Frame.GetUQHolder().PlayersByEntityID {
				if rp.Username == player {
					uuidStr := rp.UUID.String()
					onlineRecords.records = append(onlineRecords.records, &scoreRecord{Name: player, UUID: uuidStr, Score: score})
					if record, hasK := o.playerMapping[uuidStr]; hasK {
						record.Name = rp.Username
						if record.Score != score {
							record.Score = score
							needRankUpdate = true
							needSort = true
						}
					} else {
						record := &scoreRecord{Name: player, UUID: uuidStr, Score: score}
						o.playerMapping[uuidStr] = record
						o.records.records = append(o.records.records, record)
						needRankUpdate = true
						needSort = true
					}
					break
				}
			}
		}
		sort.Sort(onlineRecords)
		if needRankUpdate {
			sort.Sort(o.records)
			o.fileChange = true
		}
		if needSort {
			o.records.freshOrder()
			o.fileChange = true
		}
		if o.ScoreBoardRender.EnableScoreboardRender {
			if o.ScoreBoardRender.RenderOnlinePlayersOnly {
				o.freshScoreboardDisplay(onlineRecords.records)
			} else {
				o.freshScoreboardDisplay(o.records.records)
			}
		}
	}
}

func (o *Ranking) freshScoreboardDisplay(records []*scoreRecord) {
	// TODO update refresh algorithm

	newItems := map[string]int{}
	totalNum := len(records)
	if totalNum > o.Render.MaxCount {
		totalNum = o.Render.MaxCount
	}
	for _i, r := range records {
		if _i == o.Render.MaxCount {
			break
		}
		i := r.Score
		if o.ScoreBoardRender.ScoreboardRenderScoreBy == "序号升序" {
			i = _i + 1
		} else if o.ScoreBoardRender.ScoreboardRenderScoreBy == "序号降序" {
			i = totalNum - _i
		}
		// pterm.Info.Println(o.Render.ScoreboardRenderName)
		name := utils.FormatByReplacingOccurrences(o.ScoreBoardRender.ScoreboardRenderName, map[string]interface{}{
			"[I++]":    _i + 1,
			"[I--]":    totalNum - _i,
			"[player]": r.Name,
			"[score]":  r.Score,
		})
		// pterm.Info.Println(name)
		newItems[name] = i
	}
	scoreboardHolder := o.Frame.GetScoreboardHolder()
	target := o.ScoreBoardRender.TargetDisplayScoreboard

	oldMaintainer := o.auxData.MaintainedNames
	newMaintained := make(map[string]int)
	for k, v := range oldMaintainer {
		newMaintained[k] = v
	}
	for k, v := range newItems {
		newMaintained[k] = v
	}
	// pterm.Info.Println(newItems)
	objectsToRemove := []string{}
	objectsToUpdate := map[string]int{}
	objectsUnchanged := map[string]int{}
	ignore := false
	scoreboardHolder.Access(func(visibleSlot map[string]defines.ScoreboardDisplay, Scoreboards map[string]*defines.ScoreBoard) {
		found := false
		for _, board := range visibleSlot {
			if board.Name == target {
				found = true
				break
			}
		}
		if !found {
			return
		}
		// fmt.Println(newItems)
		if scoreboard, found := Scoreboards[target]; found {
			scoreboard.Access(func(entries map[int64]*defines.Entry) {
				for _, e := range entries {
					// fmt.Println(e)
					if !e.IsFixedName() {
						// fmt.Println("A")
						continue
					}
					if newVal, found := newItems[e.DisplayName]; found {
						if newVal != int(e.Score) {
							// fmt.Println("B")
							objectsToUpdate[e.DisplayName] = newVal
						} else {
							// fmt.Println("C")
							objectsUnchanged[e.DisplayName] = newVal
						}
					} else {
						// fmt.Println("D")
						if _, needReset := oldMaintainer[e.DisplayName]; needReset {
							// fmt.Println("E")
							objectsToRemove = append(objectsToRemove, e.DisplayName)
						}
					}
				}
			})
		} else {
			ignore = true
		}
	})
	if ignore {
		return
	}
	// pterm.Info.Println(objectsToUpdate, objectsUnchanged, objectsToRemove)
	for name, val := range newItems {
		if _, found := objectsToUpdate[name]; !found {
			if _, found = objectsUnchanged[name]; !found {
				objectsToUpdate[name] = val
			}
		}
	}
	for name, _ := range oldMaintainer {
		if _, found := objectsToUpdate[name]; !found {
			if _, found = objectsUnchanged[name]; !found {
				objectsToRemove = append(objectsToRemove, name)
			}
		}
	}

	// pterm.Info.Println("Remove: ", objectsToRemove)
	// pterm.Info.Println("Update: ", objectsToUpdate)
	for _, name := range objectsToRemove {
		o.Frame.GetGameControl().SendWOCmd(fmt.Sprintf("scoreboard players reset %v %v", name, o.ScoreBoardRender.TargetDisplayScoreboard))
		delete(newMaintained, name)
	}
	for name, val := range objectsToUpdate {
		o.Frame.GetGameControl().SendWOCmd(fmt.Sprintf("scoreboard players set %v %v %v", name, o.ScoreBoardRender.TargetDisplayScoreboard, val))
	}

	o.auxData.MaintainedNames = newMaintained
	o.auxDataChanged = false
	if len(oldMaintainer) != len(newMaintained) {
		o.auxDataChanged = true
	} else {
		for k, _ := range oldMaintainer {
			if _, found := newMaintained[k]; !found {
				o.auxDataChanged = true
				break
			}
		}
	}

	// newCmds := []string{}
	// for _i, r := range o.records.records {
	// 	if _i == o.Render.MaxCount {
	// 		break
	// 	}
	// 	i := _i + 1
	// 	fmtStr := o.Render.ScoreBoardSetFmt
	// 	cmd := utils.FormatByReplacingOccurrences(fmtStr, map[string]interface{}{
	// 		"[i]":      i,
	// 		"[player]": "\"" + r.Name + "\"",
	// 		"[score]":  r.Score,
	// 	})
	// 	newCmds = append(newCmds, cmd)
	// }
	// needUpdate := false
	// if len(newCmds) != len(o.scoreboardRenderCache) {
	// 	needUpdate = true
	// } else {
	// 	for i := range newCmds {
	// 		if newCmds[i] != o.scoreboardRenderCache[i] {
	// 			needUpdate = true
	// 			break
	// 		}
	// 	}
	// }
	// if needUpdate {
	// 	o.scoreboardRenderCache = newCmds
	// 	go func() {
	// 		o.Frame.GetGameControl().SendCmd(o.Render.ClearScoreBoardCmd)
	// 		// time.Sleep(50 * time.Millisecond)
	// 		for _, cmd := range newCmds {
	// 			o.Frame.GetGameControl().SendCmd(cmd)
	// 			// time.Sleep(50 * time.Millisecond)
	// 		}
	// 	}()
	// }
}

// func (o *Ranking) fetch(output *packet.CommandOutput) (result map[string]map[string]int) {
// 	result = nil
// 	currentPlayer := ""
// 	fetchResult := map[string]map[string]int{}

// 	for _, msg := range output.OutputMessages {
// 		if !msg.Success {
// 			return
// 		}
// 		if len(msg.Parameters) == 2 {
// 			_currentPlayer := msg.Parameters[1]
// 			if len(_currentPlayer) > 1 {
// 				currentPlayer = _currentPlayer[1:]
// 			} else {
// 				return
// 			}
// 		} else if len(msg.Parameters) == 3 {
// 			valStr, scoreboardName := msg.Parameters[0], msg.Parameters[2]
// 			val, err := strconv.Atoi(valStr)
// 			if err != nil {
// 				return
// 			}
// 			if players, hasK := fetchResult[scoreboardName]; !hasK {
// 				fetchResult[scoreboardName] = map[string]int{currentPlayer: val}
// 			} else {
// 				players[currentPlayer] = val
// 			}
// 		} else {
// 			return
// 		}
// 	}
// 	return fetchResult
// }

func (o *Ranking) Inject(frame defines.MainFrame) {
	o.Frame = frame
	plainRecords := make([]*scoreRecord, 0)
	o.Frame.GetJsonData(o.FileName, &plainRecords)
	o.records.records = plainRecords
	o.playerMapping = make(map[string]*scoreRecord)
	for _, r := range plainRecords {
		o.playerMapping[r.UUID] = r
	}
	sort.Sort(o.records)
	o.records.freshOrder()
	o.Frame.GetJsonData(o.FileName+".aux", &o.auxData)
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.onTrigger,
	})
}

func (o *Ranking) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.records.records)
		}
		if o.auxDataChanged {
			o.auxDataChanged = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName+".aux", ".ckpt", o.auxData)
		}
	}
	return nil
}

func (o *Ranking) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	if o.MaxSaveCount > 0 && o.MaxSaveCount < len(o.records.records) {
		o.records.records = o.records.records[:o.MaxSaveCount]
	}
	o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.records.records)
	return o.Frame.WriteJsonDataWithTMP(o.FileName+".aux", ".final", o.auxData)
}

func (o *Ranking) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Period))
	<-t.C
	go func() {
		time.Sleep(time.Second * 3)
		for {
			global.UpdateScore(o.Frame.GetGameControl(), time.Second*time.Duration(o.Period), func(m map[string]map[string]int) {
				o.update(m)
			})
			<-t.C
		}
	}()
}
