package micro_omega

import (
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol"
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol/packet"
	"fastbuilder-core/lib/minecraft/omega/omega"
)

type Reactor struct {
	ctrl                       *GameCtrl
	OnAnyPacketCallBack        []func(packet.Packet)
	OnTypedPacketCallBacks     map[uint32][]func(packet.Packet)
	GameChatInterceptors       []func(chat *omega.GameChat) (stop bool)
	OnKnownPlayerExistCallback []func(string)
	scoreboardHolder           *omega.ScoreBoardHolder
}

func (o *Reactor) SetGameChatInterceptor(f func(chat *omega.GameChat) (stop bool)) {
	o.GameChatInterceptors = append(o.GameChatInterceptors, f)
}

func (o *Reactor) SetOnAnyPacketCallBack(cb func(packet.Packet)) {
	o.OnAnyPacketCallBack = append(o.OnAnyPacketCallBack, cb)
}

func (o *Reactor) SetOnTypedPacketCallBack(pktID uint32, cb func(packet.Packet)) {
	if _, ok := o.OnTypedPacketCallBacks[pktID]; !ok {
		o.OnTypedPacketCallBacks[pktID] = make([]func(packet2 packet.Packet), 0, 1)
	}
	o.OnTypedPacketCallBacks[pktID] = append(o.OnTypedPacketCallBacks[pktID], cb)
}

func (o *Reactor) AppendLoginInfoCallback(cb func(entry protocol.PlayerListEntry)) {
	o.SetOnTypedPacketCallBack(packet.IDPlayerList, func(p packet.Packet) {
		pk := p.(*packet.PlayerList)
		if pk.ActionType == packet.PlayerListActionRemove {
			return
		}
		for _, player := range pk.Entries {
			cb(player)
		}
	})
}

func (o *Reactor) AppendLogoutInfoCallback(cb func(entry protocol.PlayerListEntry)) {
	o.SetOnTypedPacketCallBack(packet.IDPlayerList, func(p packet.Packet) {
		pk := p.(*packet.PlayerList)
		if pk.ActionType == packet.PlayerListActionAdd {
			return
		}
		for _, player := range pk.Entries {
			cb(player)
		}
	})
}

func (r *Reactor) React(pkt packet.Packet) {
	pktID := pkt.ID()
	switch p := pkt.(type) {
	case *packet.SetDisplayObjective:
		r.scoreboardHolder.UpdateFromSetDisplayPacket(p)
	case *packet.SetScore:
		r.scoreboardHolder.UpdateFromScorePacket(p)
	case *packet.Text:
	case *packet.GameRulesChanged:
		for _, rule := range p.GameRules {
			// o.backendLogger.Write(fmt.Sprintf("game rule update %v => %v", rule.Name, rule.Value))
			if rule.Name == "sendcommandfeedback" {
				if rule.Value == true {
					r.ctrl.onCommandFeedbackOn()
				} else {
					r.ctrl.onCommandFeedBackOff()
				}
			}
		}
		// fmt.Println(p)
	case *packet.PlayerList:
		if p.ActionType == packet.PlayerListActionAdd {
			for _, e := range p.Entries {
				for _, cb := range r.OnKnownPlayerExistCallback {
					cb(e.Username)
				}
			}
		}
	case *packet.CommandOutput:
		r.ctrl.onNewCommandFeedBack(p)
	case *packet.BlockActorData:
		r.ctrl.onBlockActor(p)
	}
	for _, cb := range r.OnAnyPacketCallBack {
		cb(pkt)
	}
	if cbs, ok := r.OnTypedPacketCallBacks[pktID]; ok {
		for _, cb := range cbs {
			cb(pkt)
		}
	}
}

func (o *Reactor) AppendOnKnownPlayerExistCallback(cb func(string)) {
	o.OnKnownPlayerExistCallback = append(o.OnKnownPlayerExistCallback, cb)
}

func NewReactor(ctrl *GameCtrl) *Reactor {
	return &Reactor{
		ctrl:                       ctrl,
		GameChatInterceptors:       make([]func(*omega.GameChat) (stop bool), 0),
		OnAnyPacketCallBack:        make([]func(packet.Packet), 0),
		OnTypedPacketCallBacks:     make(map[uint32][]func(packet.Packet), 0),
		OnKnownPlayerExistCallback: make([]func(string), 0),
		scoreboardHolder:           omega.NewScoreBoardHolder(ctrl.uq),
	}
}
