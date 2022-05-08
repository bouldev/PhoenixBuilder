package mainframe

import (
	"fmt"
	"os"
	"path"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/io"
	"phoenixbuilder/mirror/io/mcdb"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"

	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/pterm/pterm"
)

func (o *Reactor) SetGameMenuEntry(entry *defines.GameMenuEntry) {
	o.GameMenuEntries = append(o.GameMenuEntries, entry)
	interceptor := o.gameMenuEntryToStdInterceptor(entry)
	o.SetGameChatInterceptor(interceptor)
	if entry.FinalTrigger {
		o.GameChatFinalInterceptors = append(o.GameChatFinalInterceptors,
			func(chat *defines.GameChat) (stop bool) {
				return entry.OptionalOnTriggerFn(chat)
			},
		)
	}
}

func (o *Reactor) gameMenuEntryToStdInterceptor(entry *defines.GameMenuEntry) func(chat *defines.GameChat) (stop bool) {
	return func(chat *defines.GameChat) (stop bool) {
		if !chat.FrameWorkTriggered {
			return false
		}
		if trig, reducedCmds := utils.CanTrigger(chat.Msg, entry.Triggers, o.o.OmegaConfig.Trigger.AllowNoSpace,
			o.o.OmegaConfig.Trigger.RemoveSuffixColor); trig {
			_c := chat
			_c.Msg = reducedCmds
			return entry.OptionalOnTriggerFn(_c)
		}
		return false
	}
}

func (o *Reactor) SetGameChatInterceptor(f func(chat *defines.GameChat) (stop bool)) {
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

func (o *Omega) convertTextPacket(p *packet.Text) *defines.GameChat {
	name := p.SourceName
	name = utils.ToPlainName(name)

	msg := strings.TrimSpace(p.Message)
	msgs := utils.GetStringContents(msg)
	c := &defines.GameChat{
		Name: name,
		Msg:  msgs,
		Type: p.TextType,
		Aux:  p,
	}
	c.FrameWorkTriggered, c.Msg = utils.CanTrigger(
		msgs,
		o.OmegaConfig.Trigger.TriggerWords,
		o.OmegaConfig.Trigger.AllowNoSpace,
		o.OmegaConfig.Trigger.RemoveSuffixColor,
	)
	return c
}
func (o *Reactor) GetTriggerWord() string {
	return o.o.OmegaConfig.Trigger.DefaultTigger
}

func (o *Omega) GetGameListener() defines.GameListener {
	return o.Reactor
}

func (r *Reactor) Throw(chat *defines.GameChat) {
	o := r.o
	flag := true
	catchForParams := false
	if player := o.GetGameControl().GetPlayerKit(chat.Name); player != nil {
		if paramCb := player.GetOnParamMsg(); paramCb != nil {
			if !chat.FrameWorkTriggered {
				catchForParams = paramCb(chat)
			}
		}
	}
	if catchForParams {
		return
	}
	for _, interceptor := range r.GameChatInterceptors {
		if stop := interceptor(chat); stop {
			flag = false
			return
		}
	}
	chat.FallBack = true
	if flag && chat.FrameWorkTriggered {
		for _, interceptor := range r.GameChatFinalInterceptors {
			if stop := interceptor(chat); stop {
				return
			}
		}
	}
}

func (r *Reactor) React(pkt packet.Packet) {
	o := r.o
	pktID := pkt.ID()
	if pkt == nil {
		return
	}
	switch p := pkt.(type) {
	case *packet.Text:
		o.backendLogger.Write(fmt.Sprintf("%v(%v):%v", p.SourceName, p.TextType, p.Message))
		chat := o.convertTextPacket(p)
		if p.TextType == packet.TextTypeWhisper && o.OmegaConfig.Trigger.AllowWisper {
			chat.FrameWorkTriggered = true
		}
		r.Throw(chat)
	case *packet.GameRulesChanged:
		for _, rule := range p.GameRules {
			// o.backendLogger.Write(fmt.Sprintf("game rule update %v => %v", rule.Name, rule.Value))
			if rule.Name == "sendcommandfeedback" {
				if rule.Value == true {
					o.GameCtrl.onCommandFeedbackOn()
				} else {
					o.GameCtrl.onCommandFeedBackOff()
				}
			}
		}
		// fmt.Println(p)
	case *packet.PlayerList:
		if p.ActionType == packet.PlayerListActionAdd {
			for _, e := range p.Entries {
				for _, cb := range r.OnFirstSeePlayerCallback {
					cb(e.Username)
				}
			}
		}
	case *packet.CommandOutput:
		o.GameCtrl.onNewCommandFeedBack(p)
	case *packet.LevelChunk:
		if r.CurrentWorld != nil {
			chunkData := io.NEMCPacketToChunkData(p)
			if chunkData == nil {
				break
			}
			if err := r.CurrentWorld.Write(chunkData); err != nil {
				o.GetBackendDisplay().Write("Decode Chunk Error " + err.Error())
			} else {
				//fmt.Println("saving chunk @ ", p.ChunkX<<4, p.ChunkZ<<4)
			}
		}
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

type Reactor struct {
	o                         *Omega
	OnAnyPacketCallBack       []func(packet.Packet)
	OnTypedPacketCallBacks    map[uint32][]func(packet.Packet)
	GameMenuEntries           []*defines.GameMenuEntry
	GameChatInterceptors      []func(chat *defines.GameChat) (stop bool)
	GameChatFinalInterceptors []func(chat *defines.GameChat) (stop bool)
	OnFirstSeePlayerCallback  []func(string)
	CurrentWorld              mirror.ChunkProvider
}

func (o *Reactor) AppendOnFirstSeePlayerCallback(cb func(string)) {
	o.OnFirstSeePlayerCallback = append(o.OnFirstSeePlayerCallback, cb)
}

func (o *Reactor) onBootstrap() {
	worldDir := path.Join(o.o.GetWorldsDir(), "current")
	provider, err := mcdb.New(worldDir, opt.FlateCompression)
	if err != nil {
		pterm.Error.Println("创建镜像存档(" + worldDir + ")时出现错误,正在尝试移除文件夹, 错误为" + err.Error())
		if err = os.Rename(worldDir, path.Join(o.o.GetWorldsDir(), "损坏的存档")); err != nil {
			pterm.Error.Println("移除失败，错误为" + err.Error())
			//panic(err)
		}
		if provider, err = mcdb.New(worldDir, opt.FlateCompression); err != nil {
			pterm.Error.Println("修复也失败了，错误为" + err.Error())
			//panic(err)
		}
		if provider == nil {
			for i := 0; i < 10; i++ {
				pterm.Error.Println("将在没有存档相关功能的情况下运行!")
			}
		}
	} else {
		o.o.GetBackendDisplay().Write(pterm.Success.Sprint("镜像存档@" + worldDir))
	}
	if provider != nil {
		provider.D.LevelName = "MirrorWorld"
	}
	o.CurrentWorld = provider
	o.o.CloseFns = append(o.o.CloseFns, func() error {
		fmt.Println("正在关闭反射世界")
		if provider != nil {
			return provider.Close()
		}
		return nil
	})
}

func newReactor(o *Omega) *Reactor {
	return &Reactor{
		o:                         o,
		GameMenuEntries:           make([]*defines.GameMenuEntry, 0),
		GameChatInterceptors:      make([]func(chat *defines.GameChat) (stop bool), 0),
		GameChatFinalInterceptors: make([]func(chat *defines.GameChat) (stop bool), 0),
		OnAnyPacketCallBack:       make([]func(packet2 packet.Packet), 0),
		OnTypedPacketCallBacks:    make(map[uint32][]func(packet.Packet), 0),
		OnFirstSeePlayerCallback:  make([]func(string), 0),
	}
}
