package mainframe

import (
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
)

type Entry struct {
	protocol.ScoreboardEntry
	PlayerLink *uqHolder.Player
}

type scoreBoard struct {
	ObjectName string
	Entries    map[int64]*Entry
}

func newScoreboard(name string) *scoreBoard {
	return &scoreBoard{
		ObjectName: name,
		Entries:    make(map[int64]*Entry),
	}
}

func (o *scoreBoard) update(e protocol.ScoreboardEntry) {
	if entry, found := o.Entries[e.EntryID]; found {
		entry.Score = e.Score
	} else {
		o.Entries[e.EntryID] = &Entry{
			e, nil,
		}

	}

}

func (o *scoreBoard) delete(e protocol.ScoreboardEntry) {

}

type ScoreBoardHolder struct {
	uqHolder          *uqHolder.UQHolder
	visibleScoreboard map[string]*scoreBoard
	mu                sync.Mutex
	onUpdateCallbacks []func(*scoreBoard)
}

func newScoreBoardHolder(uqholder *uqHolder.UQHolder) *ScoreBoardHolder {
	o := &ScoreBoardHolder{
		uqHolder: uqholder,
	}
	return o
}

func (o *ScoreBoardHolder) Lock() {
	o.mu.Lock()
}

func (o *ScoreBoardHolder) UnLock() {
	o.mu.Unlock()
}

func (o *ScoreBoardHolder) updateFromPacket(p *packet.SetScore) {
	o.mu.Lock()
	defer o.mu.Unlock()
	updated := make(map[string]*scoreBoard)
	if p.ActionType == packet.ScoreboardActionModify {
		for _, e := range p.Entries {
			scoreboardName := e.ObjectiveName
			if scoreboard, found := o.visibleScoreboard[scoreboardName]; found {
				updated[scoreboardName] = scoreboard
				scoreboard.update(e)
			} else {
				scoreboard := newScoreboard(scoreboardName)
				o.visibleScoreboard[scoreboardName] = scoreboard
				updated[scoreboardName] = scoreboard
				o.visibleScoreboard[scoreboardName].update(e)
			}
		}
	} else {
		for _, e := range p.Entries {
			if scoreboard, found := o.visibleScoreboard[e.ObjectiveName]; found {
				scoreboard.delete(e)
				updated[e.ObjectiveName] = scoreboard
			}
		}
	}
	for _, scoreboard := range updated {
		for _, fn := range o.onUpdateCallbacks {
			fn(scoreboard)
		}
	}
}
