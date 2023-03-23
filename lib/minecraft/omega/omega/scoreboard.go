package omega

import (
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol"
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol/packet"
	"fastbuilder-core/lib/minecraft/omega/uq_holder"
	"sync"
)

type Entry struct {
	protocol.ScoreboardEntry
	uqHolder *uq_holder.UQHolder
}

func (o *Entry) IsPlayer() bool {
	return o.IdentityType == protocol.ScoreboardIdentityPlayer
}

func (o *Entry) IsFixedName() bool {
	return o.IdentityType == protocol.ScoreboardIdentityFakePlayer
}

func (o *Entry) IsEntity() bool {
	return o.IdentityType == protocol.ScoreboardIdentityEntity
}

func (o *Entry) IsBot() bool {
	return o.IdentityType == protocol.ScoreboardIdentityPlayer && o.EntityUniqueID == o.uqHolder.BotUniqueID
}

func (o *Entry) IsOnlinePlayer() bool {
	return o.IsBot() || (o.IsPlayer() && o.uqHolder.PlayersByEntityID[o.EntityUniqueID] != nil)
}

func (o *Entry) TypeStr() string {
	if o.IsFixedName() {
		return "fixedName"
	} else if o.IsEntity() {
		return "entity"
	} else if o.IsBot() {
		return "bot"
	} else if o.IsOnlinePlayer() {
		return "onlinePlayer"
	} else {
		return "offlinePlayer"
	}
}

type ScoreBoard struct {
	h          *ScoreBoardHolder
	mu         sync.Mutex
	ObjectName string
	Entries    map[int64]*Entry
}

func newScoreboard(h *ScoreBoardHolder, name string) *ScoreBoard {
	return &ScoreBoard{
		h:          h,
		mu:         sync.Mutex{},
		ObjectName: name,
		Entries:    make(map[int64]*Entry),
	}
}

func (o *ScoreBoard) update(e protocol.ScoreboardEntry) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if entry, found := o.Entries[e.EntryID]; found {
		entry.Score = e.Score
	} else {
		o.Entries[e.EntryID] = &Entry{
			e, o.h.uqHolder,
		}
	}

}

func (o *ScoreBoard) delete(e protocol.ScoreboardEntry) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.Entries, e.EntryID)
}

func (o *ScoreBoard) IsVisible() bool {
	for n, v := range o.h.visibleSlot {
		if n == "" {
			continue
		}
		if v.Name == o.ObjectName {
			return true
		}
	}
	return false
}

func (o *ScoreBoard) Access(cb func(map[int64]*Entry)) {
	o.mu.Lock()
	cb(o.Entries)
	o.mu.Unlock()
}

type ScoreboardDisplay struct {
	Name        string
	DisplayName string
}
type ScoreBoardHolder struct {
	uqHolder                 *uq_holder.UQHolder
	Scoreboards              map[string]*ScoreBoard
	mu                       sync.Mutex
	onUpdateCallbacks        []func(*ScoreBoard)
	onVisibleChangeCallbacks []func(map[string]*ScoreBoard)
	visibleSlot              map[string]ScoreboardDisplay
}

func NewScoreBoardHolder(uqholder *uq_holder.UQHolder) *ScoreBoardHolder {
	o := &ScoreBoardHolder{
		uqHolder:                 uqholder,
		Scoreboards:              make(map[string]*ScoreBoard),
		mu:                       sync.Mutex{},
		onUpdateCallbacks:        make([]func(*ScoreBoard), 0),
		onVisibleChangeCallbacks: make([]func(map[string]*ScoreBoard), 0),
		visibleSlot:              make(map[string]ScoreboardDisplay),
	}
	return o
}

func (o *ScoreBoardHolder) Access(cb func(visibleSlot map[string]ScoreboardDisplay, Scoreboards map[string]*ScoreBoard)) {
	o.mu.Lock()
	cb(o.visibleSlot, o.Scoreboards)
	o.mu.Unlock()
}

func (o *ScoreBoardHolder) UpdateFromSetDisplayPacket(p *packet.SetDisplayObjective) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.visibleSlot[p.DisplaySlot] = ScoreboardDisplay{
		Name:        p.ObjectiveName,
		DisplayName: p.DisplayName,
	}
	if _, found := o.Scoreboards[p.ObjectiveName]; !found {
		scoreboard := newScoreboard(o, p.ObjectiveName)
		o.Scoreboards[p.ObjectiveName] = scoreboard
	}
}

func (o *ScoreBoardHolder) UpdateFromScorePacket(p *packet.SetScore) {
	o.mu.Lock()
	defer o.mu.Unlock()
	updated := make(map[string]*ScoreBoard)
	if p.ActionType == packet.ScoreboardActionModify {
		for _, e := range p.Entries {
			scoreboardName := e.ObjectiveName
			if scoreboard, found := o.Scoreboards[scoreboardName]; found {
				updated[scoreboardName] = scoreboard
				scoreboard.update(e)
			} else {
				scoreboard := newScoreboard(o, scoreboardName)
				o.Scoreboards[scoreboardName] = scoreboard
				updated[scoreboardName] = scoreboard
				o.Scoreboards[scoreboardName].update(e)
			}
		}
	} else {
		for _, e := range p.Entries {
			if scoreboard, found := o.Scoreboards[e.ObjectiveName]; found {
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
