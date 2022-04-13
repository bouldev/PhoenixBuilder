package mainframe

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"sync"
	"time"
)

type PlayerKitOmega struct {
	uq                   *uqHolder.UQHolder
	ctrl                 *GameCtrl
	name                 string
	persistStorageCached map[string]bool
	persistStorage       map[string][]byte
	violatedStorage      map[string]interface{}
	OnParamMsg           func(chat *defines.GameChat) (catch bool)
	playerUQ             *uqHolder.Player
}

func (p *PlayerKitOmega) SetOnParamMsg(f func(chat *defines.GameChat) (catch bool)) error {
	if p.OnParamMsg != nil {
		return fmt.Errorf("player busy")
	}
	p.OnParamMsg = f
	return nil
}

func (p *PlayerKitOmega) GetOnParamMsg() func(chat *defines.GameChat) (catch bool) {
	f := p.OnParamMsg
	p.OnParamMsg = nil
	return f
}

func (p *PlayerKitOmega) GetPersistStorage(k string) string {
	return p.ctrl.playerStorageDB.Get("." + p.name + k)
}

func (p *PlayerKitOmega) GetViolatedStorage() map[string]interface{} {
	return p.violatedStorage
}

func (p *PlayerKitOmega) CommitPersistStorageChange(k string, v string) {
	if v == "" {
		p.ctrl.playerStorageDB.Delete("." + p.name + k)
		return
	}
	p.ctrl.playerStorageDB.Commit("."+p.name+k, v)
}

// not tested
func (p *PlayerKitOmega) preparePrePlayerStorage() {
	uq := p.GetRelatedUQ()
	if uq != nil {
		ud := uq.UUID.String()
		currentNameKey := fmt.Sprintf(".%v.current_name.name", ud)
		currentTimeKey := fmt.Sprintf(".%v.current_name.time", ud)
		nameHistoryKey := fmt.Sprintf(".%v.current_name.history", ud)
		currentTime := utils.TimeToString(time.Now())
		record := p.ctrl.playerNameDB.Get(currentNameKey)
		if record == "" {
			m, _ := json.Marshal([][]string{
				[]string{currentTime, p.name},
			})
			p.ctrl.playerNameDB.Commit(nameHistoryKey, string(m))
		} else if record != p.name {
			oldName := record
			newName := p.name
			records := p.ctrl.playerNameDB.Get(nameHistoryKey)
			var his [][]string
			err := json.Unmarshal([]byte(records), &his)
			if err != nil {
				fmt.Println(err)
			}
			his = append(his, []string{currentTime, newName})
			m, _ := json.Marshal([][]string{
				[]string{currentTime, newName},
			})
			p.ctrl.playerNameDB.Commit(nameHistoryKey, string(m))
			p.ctrl.playerStorageDB.IterWithPrefix(func(key string, v string) (stop bool) {
				newKey := strings.Replace(key, oldName, newName, 1)
				p.ctrl.playerStorageDB.Commit(newKey, v)
				p.ctrl.playerStorageDB.Delete(key)
				return false
			}, "."+oldName)
		}
		p.ctrl.playerNameDB.Commit(currentNameKey, p.name)
		p.ctrl.playerNameDB.Commit(currentTimeKey, currentTime)
		p.CommitPersistStorageChange(".last_login_time", currentTime)
	}
}

func newPlayerKitOmega(uq *uqHolder.UQHolder, ctrl *GameCtrl, name string) *PlayerKitOmega {
	pko, k := ctrl.perPlayerStorage[name]
	if k {
		return pko
	}
	player := &PlayerKitOmega{
		uq:                   uq,
		ctrl:                 ctrl,
		name:                 name,
		persistStorageCached: map[string]bool{},
		persistStorage:       map[string][]byte{},
		violatedStorage:      map[string]interface{}{},
		OnParamMsg:           nil,
	}
	player.preparePrePlayerStorage()
	ctrl.perPlayerStorage[name] = player
	return player
}

func (p *PlayerKitOmega) Say(msg string) {
	p.ctrl.SayTo(p.name, msg)
}

func (p *PlayerKitOmega) ActionBar(msg string) {
	p.ctrl.ActionBarTo(p.name, msg)
}

func (p *PlayerKitOmega) Title(msg string) {
	p.ctrl.TitleTo(p.name, msg)
}

func (p *PlayerKitOmega) SubTitle(msg string) {
	p.ctrl.SubTitleTo(p.name, msg)
}

func (p *PlayerKitOmega) GetRelatedUQ() *uqHolder.Player {
	if p.playerUQ != nil {
		return p.playerUQ
	}
	for _, player := range p.uq.PlayersByEntityID {
		if player.Username == p.name {
			p.playerUQ = player
			return player
		}
	}
	return nil
}

type GameCtrl struct {
	WriteFn             func(packet packet.Packet)
	ExpectedCmdFeedBack bool
	CurrentCmdFeedBack  bool
	CmdFeedBackOnSent   bool
	NeedFeedBackPackets []packet.Packet
	uuidMaps            map[string]func(*packet.CommandOutput)
	uuidLock            sync.Mutex
	uq                  *uqHolder.UQHolder
	perPlayerStorage    map[string]*PlayerKitOmega
	playerNameDB        defines.NoSqlDB
	playerStorageDB     defines.NoSqlDB
}

func (g *GameCtrl) GetPlayerKit(name string) defines.PlayerKit {
	return newPlayerKitOmega(g.uq, g, name)
}

func (g *GameCtrl) GetPlayerKitByUUID(ud uuid.UUID) defines.PlayerKit {
	player := g.uq.GetPlayersByUUID(ud)
	if player == nil {
		return nil
	}
	return newPlayerKitOmega(g.uq, g, player.Username)
}

func (g *GameCtrl) SendCmdAndInvokeOnResponse(cmd string, cb func(*packet.CommandOutput)) {
	//if !g.CurrentCmdFeedBack && !g.CmdFeedBackOnSent {
	//	g.turnOnFeedBack()
	//}
	ud, _ := uuid.NewUUID()
	g.uuidLock.Lock()
	g.uuidMaps[ud.String()] = cb
	g.uuidLock.Unlock()
	pkt := g.packCmdWithUUID(cmd, ud)
	//if g.CurrentCmdFeedBack {
	//	g.WriteFn(pkt)
	//} else {
	//	g.NeedFeedBackPackets = append(g.NeedFeedBackPackets)
	//}
	g.WriteFn(pkt)
}

type TellrawItem struct {
	Text string `json:"text"`
}
type TellrawStruct struct {
	RawText []TellrawItem `json:"rawtext"`
}

func ToJsonRawString(line string) string {
	var items []TellrawItem
	msg := strings.Replace(line, "schematic", "sc***atic", -1)
	items = append(items, TellrawItem{Text: msg})
	final := &TellrawStruct{
		RawText: items,
	}
	content, _ := json.Marshal(final)
	return string(content)
}

func (g *GameCtrl) SayTo(target string, line string) {
	content := ToJsonRawString(line)
	g.SendCmd(fmt.Sprintf("tellraw %v %v", target, content))
}

func (g *GameCtrl) ActionBarTo(target string, line string) {
	content := ToJsonRawString(line)
	g.SendCmd(fmt.Sprintf("titleraw %v actionbar %v", target, content))
}

func (g *GameCtrl) TitleTo(target string, line string) {
	content := ToJsonRawString(line)
	g.SendCmd(fmt.Sprintf("titleraw %v title %v", target, content))
}

func (g *GameCtrl) SubTitleTo(target string, line string) {
	content := ToJsonRawString(line)
	g.SendCmd(fmt.Sprintf("titleraw %v subtitle %v", target, content))
}

func (g *GameCtrl) packCmdWithUUID(cmd string, ud uuid.UUID) *packet.CommandRequest {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin{
		Origin:         protocol.CommandOriginAutomationPlayer,
		UUID:           ud,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	}
	commandRequest := &packet.CommandRequest{
		CommandLine:   cmd,
		CommandOrigin: origin,
		Internal:      false,
		UnLimited:     false,
	}
	return commandRequest

}

func (g *GameCtrl) SendCmd(cmd string) {
	ud, _ := uuid.NewUUID()
	g.WriteFn(g.packCmdWithUUID(cmd, ud))
}

// onCommandFeedbackOnCmds is called by reactor to send commands by that need feedback
func (g *GameCtrl) onCommandFeedbackOn() {
	fmt.Println("recv sendcommandfeedback ture")
	g.CurrentCmdFeedBack = true
	g.CmdFeedBackOnSent = false
	//pkts := g.NeedFeedBackPackets
	//g.NeedFeedBackPackets = make([]packet.Packet, 0)
	//for _, p := range pkts {
	//	g.SendMCPacket(p)
	//}
	if !g.ExpectedCmdFeedBack {
		g.turnOffFeedBack()
	}
}

func (g *GameCtrl) onCommandFeedBackOff() {
	if g.ExpectedCmdFeedBack {
		g.turnOnFeedBack()
	}
}

func (g *GameCtrl) onNewCommandFeedBack(p *packet.CommandOutput) {
	s := p.CommandOrigin.UUID.String()
	if cb, hasK := g.uuidMaps[s]; hasK {
		//fmt.Println("Hit!")
		cb(p)
		g.uuidLock.Lock()
		delete(g.uuidMaps, s)
		g.uuidLock.Unlock()
	}
}

func (g *GameCtrl) turnOnFeedBack() {
	//fmt.Println("send sendcommandfeedback ture")
	g.SendCmd("gamerule sendcommandfeedback true")
	g.CmdFeedBackOnSent = true
}

func (g *GameCtrl) turnOffFeedBack() {
	g.CurrentCmdFeedBack = false
	g.CmdFeedBackOnSent = false
	//fmt.Println("send sendcommandfeedback false")
	g.SendCmd("gamerule sendcommandfeedback false")
}

func (g *GameCtrl) toExpectedFeedBackStatus() {
	if g.ExpectedCmdFeedBack {
		g.turnOnFeedBack()
	} else {
		g.turnOffFeedBack()
	}
}

func (g *GameCtrl) SendMCPacket(p packet.Packet) {
	g.WriteFn(p)
}

func (g *GameCtrl) SetOnParamMsg(name string, cb func(chat *defines.GameChat) (catch bool)) error {
	player := g.GetPlayerKit(name)
	if player != nil {
		return player.SetOnParamMsg(cb)
	} else {
		return fmt.Errorf("没有这个玩家" + name)
	}
}

func newGameCtrl(o *Omega) *GameCtrl {
	c := &GameCtrl{
		WriteFn:             o.adaptor.Write,
		ExpectedCmdFeedBack: o.fullConfig.CommandFeedBackByDefault,
		CurrentCmdFeedBack:  o.uqHolder.GameRules["sendcommandfeedback"].Value.(bool),
		uuidLock:            sync.Mutex{},
		uuidMaps:            make(map[string]func(output *packet.CommandOutput)),
		NeedFeedBackPackets: make([]packet.Packet, 0),
		uq:                  o.uqHolder,
		perPlayerStorage:    make(map[string]*PlayerKitOmega),
		playerNameDB:        o.GetNoSqlDB("playerNameDB"),
		playerStorageDB:     o.GetNoSqlDB("playerStorageDB"),
	}
	c.toExpectedFeedBackStatus()
	return c
}

func (o *Omega) GetGameControl() defines.GameControl {
	return o.GameCtrl
}
