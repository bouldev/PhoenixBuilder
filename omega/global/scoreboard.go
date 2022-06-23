package global

import (
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strconv"
	"time"
)

var rankingLastFetchTime time.Time
var rankingLastFetchResult map[string]map[string]int

func UpdateScore(ctrl defines.GameControl, allowDuration time.Duration, onUpdateDone func(map[string]map[string]int)) {
	if rankingLastFetchResult != nil {
		if time.Since(rankingLastFetchTime) < allowDuration {
			onUpdateDone(rankingLastFetchResult)
		}
	}
	ctrl.SendCmdAndInvokeOnResponse("scoreboard players list @a", func(output *packet.CommandOutput) {
		fetch := func(output *packet.CommandOutput) (result map[string]map[string]int) {
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
		if result := fetch(output); result == nil {
		} else {
			rankingLastFetchResult = result
			rankingLastFetchTime = time.Now()
			onUpdateDone(rankingLastFetchResult)
		}
	})
}
