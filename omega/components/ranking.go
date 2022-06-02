package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"sort"
	"strconv"
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
	MaxCount           int               `json:"最大显示的人数"`
	DefaultRenderFmt   string            `json:"默认渲染样式"`
	SpecificRenderFmt  map[string]string `json:"特定排名渲染样式"`
	PlayerRenderFmt    string            `json:"查询者渲染样式"`
	HintOnNoPlayer     string            `json:"当查询者没有相关分数时显示"`
	Head               string            `json:"开头"`
	Tail               string            `json:"结尾"`
	Split              string            `json:"隔断"`
	RenderScoreBoard   bool              `json:"渲染计分板"`
	ClearScoreBoardCmd string            `json:"清除计分板指令"`
	ScoreBoardSetFmt   string            `json:"计分板渲染指令"`
}

type Ranking struct {
	*BasicComponent
	Triggers              []string          `json:"触发词"`
	Usage                 string            `json:"提示信息"`
	ScoreboardName        string            `json:"计分板名"`
	FileName              string            `json:"排名记录文件"`
	MaxSaveCount          int               `json:"最多保存多少记录在文件中"`
	Ascending             bool              `json:"升序"`
	Period                int               `json:"刷新周期"`
	Render                *RankRenderOption `json:"渲染选项"`
	records               *scoreRecords
	playerMapping         map[string]*scoreRecord
	scoreboardRenderCache []string
}

var rankingLastFetchTime time.Time
var rankingLastFetchResult map[string]map[string]int

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

func (o *Ranking) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.records = &scoreRecords{
		ascending: o.Ascending,
	}
	o.scoreboardRenderCache = []string{}
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

func (o *Ranking) update() {
	if players, hasK := rankingLastFetchResult[o.ScoreboardName]; !hasK {
		pterm.Error.Println("没有计分板 %v,所有的计分板为", o.ScoreboardName)
		for n, _ := range rankingLastFetchResult {
			pterm.Error.Println(n)
		}
	} else {
		needSort := false
		needRankUpdate := false
		for player, score := range players {
			for _, rp := range o.Frame.GetUQHolder().PlayersByEntityID {
				if rp.Username == player {
					uuidStr := rp.UUID.String()
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
		if needRankUpdate {
			sort.Sort(o.records)
		}
		if needSort {
			o.records.freshOrder()
		}
		if needSort || needRankUpdate || len(o.scoreboardRenderCache) == 0 {
			o.freshScoreboardDisplay()
		}
	}
}

func (o *Ranking) freshScoreboardDisplay() {
	if !o.Render.RenderScoreBoard {
		return
	}
	newCmds := []string{}
	for _i, r := range o.records.records {
		if _i == o.Render.MaxCount {
			break
		}
		i := _i + 1
		fmtStr := o.Render.ScoreBoardSetFmt
		cmd := utils.FormatByReplacingOccurrences(fmtStr, map[string]interface{}{
			"[i]":      i,
			"[player]": "\"" + r.Name + "\"",
			"[score]":  r.Score,
		})
		newCmds = append(newCmds, cmd)
	}
	needUpdate := false
	if len(newCmds) != len(o.scoreboardRenderCache) {
		needUpdate = true
	} else {
		for i := range newCmds {
			if newCmds[i] != o.scoreboardRenderCache[i] {
				needUpdate = true
				break
			}
		}
	}
	if needUpdate {
		o.Frame.GetGameControl().SendCmd(o.Render.ClearScoreBoardCmd)
		for _, cmd := range newCmds {
			o.Frame.GetGameControl().SendCmd(cmd)
		}
		o.scoreboardRenderCache = newCmds
	}
}

func (o *Ranking) fetch(output *packet.CommandOutput) (result map[string]map[string]int) {
	result = nil
	currentPlayer := ""
	fetchResult := map[string]map[string]int{}

	for _, msg := range output.OutputMessages {
		if !msg.Success {
			return
		}
		if len(msg.Parameters) == 2 {
			_currentPlayer := msg.Parameters[1]
			if len(_currentPlayer) > 1 {
				currentPlayer = _currentPlayer[1:]
			} else {
				return
			}
		} else if len(msg.Parameters) == 3 {
			valStr, scoreboardName := msg.Parameters[0], msg.Parameters[2]
			val, err := strconv.Atoi(valStr)
			if err != nil {
				return
			}
			if players, hasK := fetchResult[scoreboardName]; !hasK {
				fetchResult[scoreboardName] = map[string]int{currentPlayer: val}
			} else {
				players[currentPlayer] = val
			}
		} else {
			return
		}
	}
	return fetchResult
}

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

func (o *Ranking) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	if o.MaxSaveCount > 0 && o.MaxSaveCount < len(o.records.records) {
		o.records.records = o.records.records[:o.MaxSaveCount]
	}
	return o.Frame.WriteJsonData(o.FileName, o.records.records)
}

func (o *Ranking) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Period))
	go func() {
		for {
			if time.Now().Sub(rankingLastFetchTime) > time.Second*time.Duration(o.Period) {
				rankingLastFetchTime = time.Now()
				o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("scoreboard players list @a", func(output *packet.CommandOutput) {
					if result := o.fetch(output); result == nil {
						// m, err := json.Marshal(output)
						// s := ""
						// if err != nil {
						// 	s = err.Error()
						// } else {
						// 	s = string(m)
						// }
						// pterm.Error.Println("抓取玩家计分板信息失败:" + s)
					} else {
						rankingLastFetchResult = result
						o.update()
					}
				})
			} else {
				if rankingLastFetchResult != nil {
					o.update()
				}
			}
			<-t.C
		}
	}()
}
