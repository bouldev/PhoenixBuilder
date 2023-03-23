package micro_omega

import (
	"encoding/json"
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol"
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol/packet"
	"fastbuilder-core/lib/minecraft/mirror/define"
	"fastbuilder-core/lib/minecraft/omega/omega"
	"fastbuilder-core/lib/minecraft/omega/uq_holder"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GameCtrl struct {
	WriteFn             func(packet packet.Packet) error
	ExpectedCmdFeedBack bool
	currentCmdFeedBack  bool
	CmdFeedBackOnSent   bool
	NeedFeedBackPackets []packet.Packet
	uuidMaps            map[string]func(*packet.CommandOutput)
	uuidLock            sync.Mutex
	uq                  *uq_holder.UQHolder
	perPlayerStorage    map[string]*PlayerKitOmega
	onBlockActorCbs     map[define.CubePos]func(define.CubePos, *packet.BlockActorData)
	blockActorLock      sync.Mutex
}

func (g *GameCtrl) GetPlayerKit(name string) omega.PlayerKit {
	return newPlayerKitOmega(g.uq, g, name)
}

func (g *GameCtrl) GetPlayerKitByUUID(ud uuid.UUID) omega.PlayerKit {
	player := g.uq.GetPlayersByUUID(ud)
	if player == nil {
		return nil
	}
	return newPlayerKitOmega(g.uq, g, player.Username)
}

func (g *GameCtrl) SendCmdAndInvokeOnResponseWithFeedback(cmd string, cb func(*packet.CommandOutput)) {
	if !g.currentCmdFeedBack && !g.CmdFeedBackOnSent {
		g.turnOnFeedBack()
	}
	ud, _ := uuid.NewUUID()
	g.uuidLock.Lock()
	g.uuidMaps[ud.String()] = cb
	g.uuidLock.Unlock()
	pkt := g.packCmdWithUUID(cmd, ud, false)
	if g.currentCmdFeedBack {
		g.WriteFn(pkt)
	} else {
		g.NeedFeedBackPackets = append(g.NeedFeedBackPackets)
	}
	g.WriteFn(pkt)
}

func (g *GameCtrl) SendCmdAndInvokeOnResponse(cmd string, cb func(*packet.CommandOutput)) {
	//if !g.currentCmdFeedBack && !g.CmdFeedBackOnSent {
	//	g.turnOnFeedBack()
	//}
	ud, _ := uuid.NewUUID()
	g.uuidLock.Lock()
	g.uuidMaps[ud.String()] = cb
	g.uuidLock.Unlock()
	pkt := g.packCmdWithUUID(cmd, ud, true)
	g.WriteFn(pkt)
}

type TellrawItem struct {
	Text string `json:"text"`
}
type TellrawStruct struct {
	RawText []TellrawItem `json:"rawtext"`
}

func toJsonRawString(line string) string {
	final := &TellrawStruct{
		RawText: []TellrawItem{{Text: line}},
	}
	content, _ := json.Marshal(final)
	return string(content)
}

func (g *GameCtrl) SayTo(target string, line string) {
	if line == "" {
		return
	}
	content := toJsonRawString(line)
	if strings.HasPrefix(target, "@") {
		g.SendWOCmd(fmt.Sprintf("tellraw %v %v", target, content))
	} else {
		g.SendWOCmd(fmt.Sprintf("tellraw \"%v\" %v", target, content))
	}
}

func (g *GameCtrl) RawSayTo(target string, line string) {
	if line == "" {
		return
	}
	content := line
	if strings.HasPrefix(target, "@") {
		g.SendWOCmd(fmt.Sprintf("tell %v %v", target, content))
	} else {
		g.SendWOCmd(fmt.Sprintf("tell \"%v\" %v", target, content))
	}
}

func (g *GameCtrl) BotSay(msg string) {
	pk := &packet.Text{
		TextType:         packet.TextTypeChat,
		NeedsTranslation: false,
		SourceName:       g.uq.BotName,
		Message:          msg,
		XUID:             "",
		PlayerRuntimeID:  fmt.Sprintf("%d", g.uq.BotRuntimeID),
	}
	g.SendMCPacket(pk)
}

func (g *GameCtrl) ActionBarTo(target string, line string) {
	content := toJsonRawString(line)
	if strings.HasPrefix(target, "@") {
		g.SendWOCmd(fmt.Sprintf("titleraw %v actionbar %v", target, content))
	} else {
		g.SendWOCmd(fmt.Sprintf("titleraw \"%v\" actionbar %v", target, content))
	}

}

func (g *GameCtrl) TitleTo(target string, line string) {
	content := toJsonRawString(line)
	if strings.HasPrefix(target, "@") {
		g.SendCmd(fmt.Sprintf("titleraw %v title %v", target, content))
	} else {
		g.SendCmd(fmt.Sprintf("titleraw \"%v\" title %v", target, content))
	}
}

func (g *GameCtrl) SubTitleTo(target string, line string) {
	content := toJsonRawString(line)
	if strings.HasPrefix(target, "@") {
		g.SendCmd(fmt.Sprintf("titleraw %v subtitle %v", target, content))
	} else {
		g.SendCmd(fmt.Sprintf("titleraw \"%v\" subtitle %v", target, content))
	}
}

func (g *GameCtrl) packCmdWithUUID(cmd string, ud uuid.UUID, ws bool) *packet.CommandRequest {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin{
		Origin:         protocol.CommandOriginAutomationPlayer,
		UUID:           ud,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	}
	if !ws {
		origin.Origin = protocol.CommandOriginPlayer
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
	g.WriteFn(g.packCmdWithUUID(cmd, ud, true))
}

func (g *GameCtrl) SendCmdWithUUID(cmd string, ud uuid.UUID, ws bool) {
	g.WriteFn(g.packCmdWithUUID(cmd, ud, ws))
}

func (g *GameCtrl) SendWOCmd(cmd string) {
	g.WriteFn(&packet.SettingsCommand{
		CommandLine:    cmd,
		SuppressOutput: true,
	})
}

// onCommandFeedbackOnCmds is called by reactor to send commands by that need feedback
func (g *GameCtrl) onCommandFeedbackOn() {
	// fmt.Println("recv sendcommandfeedback true")
	g.currentCmdFeedBack = true
	g.CmdFeedBackOnSent = false
	pkts := g.NeedFeedBackPackets
	g.NeedFeedBackPackets = make([]packet.Packet, 0)
	for _, p := range pkts {
		g.SendMCPacket(p)
	}
	if !g.ExpectedCmdFeedBack {
		g.turnOffFeedBack()
	}
}

func (g *GameCtrl) onBlockActor(p *packet.BlockActorData) {
	pos := define.CubePos{int(p.Position.X()), int(p.Position.Y()), int(p.Position.Z())}
	g.blockActorLock.Lock()
	if cb, found := g.onBlockActorCbs[pos]; found {
		delete(g.onBlockActorCbs, pos)
		g.blockActorLock.Unlock()
		cb(pos, p)
	} else {
		g.blockActorLock.Unlock()
	}
}

func (g *GameCtrl) PlaceCommandBlock(pos define.CubePos, commandBlockName string, blockDataOrStateStr string,
	withMove, withAirPrePlace bool, updatePacket *packet.CommandBlockUpdate,
	onDone func(done bool), timeOut time.Duration) {
	done := make(chan bool)
	go func() {
		select {
		case <-time.NewTimer(timeOut).C:
			onDone(false)
		case <-done:
		}
	}()
	go func() {
		if withMove {
			g.SendCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
			time.Sleep(100 * time.Millisecond)
		}
		if withAirPrePlace {
			cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], "air", 0)
			g.SendWOCmd(cmd)
			time.Sleep(100 * time.Millisecond)
		}
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], strings.Replace(commandBlockName, "minecraft:", "", 1), blockDataOrStateStr)
		g.SendWOCmd(cmd)
		g.blockActorLock.Lock()
		g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
			go func() {
				g.blockActorLock.Lock()
				g.SendCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
				time.Sleep(50 * time.Millisecond)
				g.SendMCPacket(updatePacket)
				g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
					onDone(true)
					done <- true
				}
				g.blockActorLock.Unlock()
			}()
		}
		g.blockActorLock.Unlock()

	}()
}

func (g *GameCtrl) PlaceSignBlock(pos define.CubePos, signBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.BlockActorData, onDone func(done bool), timeOut time.Duration) {
	done := make(chan bool)
	go func() {
		select {
		case <-time.NewTimer(timeOut).C:
			onDone(false)
		case <-done:
		}
	}()
	go func() {
		if withMove {
			g.SendCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
			time.Sleep(100 * time.Millisecond)
		}
		if withAirPrePlace {
			cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], "air", 0)
			g.SendWOCmd(cmd)
			time.Sleep(100 * time.Millisecond)
		}
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], strings.Replace(signBlockName, "minecraft:", "", 1), blockDataOrStateStr)
		g.SendWOCmd(cmd)
		g.blockActorLock.Lock()
		g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
			go func() {
				g.blockActorLock.Lock()
				g.SendCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
				time.Sleep(50 * time.Millisecond)
				g.SendMCPacket(updatePacket)
				g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
					onDone(true)
					done <- true
				}
				g.blockActorLock.Unlock()
			}()
		}
		g.blockActorLock.Unlock()
	}()
}

func (g *GameCtrl) onCommandFeedBackOff() {
	if g.ExpectedCmdFeedBack {
		g.turnOnFeedBack()
	}
}

func (g *GameCtrl) onNewCommandFeedBack(p *packet.CommandOutput) {
	s := p.CommandOrigin.UUID.String()
	g.uuidLock.Lock()
	cb, hasK := g.uuidMaps[s]
	g.uuidLock.Unlock()
	if hasK {
		//fmt.Println("Hit!")
		cb(p)
		g.uuidLock.Lock()
		delete(g.uuidMaps, s)
		g.uuidLock.Unlock()
	}
}

func (g *GameCtrl) turnOnFeedBack() {
	//fmt.Println("send sendcommandfeedback true")
	g.SendCmd("gamerule sendcommandfeedback true")
	g.CmdFeedBackOnSent = true
}

func (g *GameCtrl) turnOffFeedBack() {
	g.currentCmdFeedBack = false
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

func (g *GameCtrl) SetOnParamMsg(name string, cb func(chat *omega.GameChat) (catch bool)) error {
	player := g.GetPlayerKit(name)
	if player != nil {
		return player.SetOnParamMsg(cb)
	} else {
		return fmt.Errorf("没有这个玩家" + name)
	}
}

func NewGameCtrl(uq *uq_holder.UQHolder, writeFn func(packet.Packet) error) *GameCtrl {
	c := &GameCtrl{
		uq:                  uq,
		WriteFn:             writeFn,
		ExpectedCmdFeedBack: false,
		currentCmdFeedBack:  false,
		uuidLock:            sync.Mutex{},
		uuidMaps:            make(map[string]func(output *packet.CommandOutput)),
		NeedFeedBackPackets: make([]packet.Packet, 0),
		perPlayerStorage:    make(map[string]*PlayerKitOmega),
		//playerNameDB:        o.GetNoSqlDB("playerNameDB"),
		//playerStorageDB:     o.GetNoSqlDB("playerStorageDB"),
		onBlockActorCbs: make(map[define.CubePos]func(define.CubePos, *packet.BlockActorData)),
		blockActorLock:  sync.Mutex{},
	}
	c.toExpectedFeedBackStatus()
	c.SendCmd("gamemode c @s")
	return c
}
