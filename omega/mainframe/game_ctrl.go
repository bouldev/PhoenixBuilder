package mainframe

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PlayerKitOmega struct {
	uq              *uqHolder.UQHolder
	ctrl            *GameCtrl
	name            string
	persistStorage  map[string]string
	violatedStorage map[string]interface{}
	OnParamMsg      func(chat *defines.GameChat) (catch bool)
	playerUQ        *uqHolder.Player
	Permission      map[string]bool
}

func (p *PlayerKitOmega) HasPermission(key string) bool {
	if auth, hasK := p.Permission[key]; hasK && auth {
		return true
	}
	return false
}
func (b *PlayerKitOmega) GetPlayerNameByUUid(Theuuid string) string {
	UUID, err := uuid.Parse(Theuuid)
	if err != nil {
		fmt.Println(err)
	}
	if player := b.ctrl.GetPlayerKitByUUID(UUID); player != nil {
		username := player.GetRelatedUQ().Username
		return username
	}
	return ""
}
func (p *PlayerKitOmega) GetPos(selector string) chan *define.CubePos {
	s := utils.FormatByReplacingOccurrences(selector, map[string]interface{}{
		"[player]": "\"" + p.name + "\"",
	})
	c := make(chan *define.CubePos)
	sent := false
	send := func(d *define.CubePos) {
		if sent {
			return
		}
		sent = true
		c <- d
	}
	var QueryResults []struct {
		Position *struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"position"`
		Uuid string `json:"uniqueId"`
	}

	p.ctrl.SendCmdAndInvokeOnResponse("querytarget "+s, func(output *packet.CommandOutput) {
		//fmt.Println(output.OutputMessages)
		//list := make(map[string][]int64)
		if output.SuccessCount > 0 {
			for _, v := range output.OutputMessages {
				//fmt.Println("v.message:")
				for _, j := range v.Parameters {
					//fmt.Println("\nj:", j)
					err := json.Unmarshal([]byte(j), &QueryResults)
					if err != nil {
						send(nil)
					}
					for _, u := range QueryResults {
						send(&define.CubePos{
							int(u.Position.X),
							int(u.Position.Y),
							int(u.Position.Z),
						})
					}
				}
			}

		}
		send(nil)
	})
	go func() {
		<-time.NewTicker(time.Second * 3).C
		send(nil)
	}()
	return c
}

func (p *PlayerKitOmega) SetPermission(key string, b bool) {
	p.Permission[key] = b
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

//func (p *PlayerKitOmega) GetPersistStorage(k string) string {
//	if val, hasK := p.persistStorage[k]; !hasK {
//		return ""
//	} else {
//		return val
//	}
//}

func (p *PlayerKitOmega) GetViolatedStorage() map[string]interface{} {
	return p.violatedStorage
}

//func (p *PlayerKitOmega) CommitPersistStorageChange(k string, v string) {
//	if _, hasK := p.persistStorage[k]; !hasK {
//		return
//	}
//	if v == "" {
//		delete(p.persistStorage, k)
//		p.ctrl.playerStorageDB.Delete("." + p.name + k)
//		return
//	}
//	p.persistStorage[k] = v
//	p.ctrl.playerStorageDB.Commit("."+p.name+k, v)
//}
//
//// not tested
//func (p *PlayerKitOmega) preparePrePlayerStorage() {
//	uq := p.GetRelatedUQ()
//	if uq != nil {
//		ud := uq.UUID.String()
//		currentNameKey := fmt.Sprintf(".%v.current_name.name", ud)
//		currentTimeKey := fmt.Sprintf(".%v.current_name.time", ud)
//		nameHistoryKey := fmt.Sprintf(".%v.current_name.history", ud)
//		currentTime := utils.TimeToString(time.Now())
//		record := p.ctrl.playerNameDB.Get(currentNameKey)
//		if record == "" {
//			m, _ := json.Marshal([][]string{
//				[]string{currentTime, p.name},
//			})
//			p.ctrl.playerNameDB.Commit(nameHistoryKey, string(m))
//		} else if record != p.name {
//			oldName := record
//			newName := p.name
//			records := p.ctrl.playerNameDB.Get(nameHistoryKey)
//			var his [][]string
//			err := json.Unmarshal([]byte(records), &his)
//			if err != nil {
//				fmt.Println(err)
//			}
//			his = append(his, []string{currentTime, newName})
//			m, _ := json.Marshal([][]string{
//				[]string{currentTime, newName},
//			})
//			p.ctrl.playerNameDB.Commit(nameHistoryKey, string(m))
//			p.ctrl.playerStorageDB.IterWithPrefix(func(key string, v string) (stop bool) {
//				newKey := strings.Replace(key, oldName, newName, 1)
//				p.ctrl.playerStorageDB.Commit(newKey, v)
//				p.ctrl.playerStorageDB.Delete(key)
//				return false
//			}, "."+oldName)
//		}
//		p.ctrl.playerNameDB.Commit(currentNameKey, p.name)
//		p.ctrl.playerNameDB.Commit(currentTimeKey, currentTime)
//		p.CommitPersistStorageChange(".last_login_time", currentTime)
//	}
//	p.ctrl.playerStorageDB.IterWithPrefix(func(key string, v string) (stop bool) {
//		p.persistStorage[key] = v
//		return false
//	}, "."+p.name)
//	if p.ctrl.PlayerPermission[p.name] == nil {
//		p.ctrl.PlayerPermission[p.name] = map[string]bool{}
//	}
//	p.Permission = p.ctrl.PlayerPermission[p.name]
//}

func newPlayerKitOmega(uq *uqHolder.UQHolder, ctrl *GameCtrl, name string) *PlayerKitOmega {
	pko, k := ctrl.perPlayerStorage[name]
	if k {
		return pko
	}
	player := &PlayerKitOmega{
		uq:   uq,
		ctrl: ctrl,
		name: name,
		//persistStorage:  map[string]string{},
		violatedStorage: map[string]interface{}{},
		OnParamMsg:      nil,
	}
	//player.preparePrePlayerStorage()
	ctrl.perPlayerStorage[name] = player
	return player
}

func (p *PlayerKitOmega) Say(msg string) {
	p.ctrl.SayTo(p.name, msg)
}

func (p *PlayerKitOmega) RawSay(msg string) {
	p.ctrl.RawSayTo(p.name, msg)
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

type timeCmdPair struct {
	time time.Time
	cmd  string
}
type PacketOutAnalyzer struct {
	WriteFn                  func(p packet.Packet) (err error)
	packetSendStats          map[time.Time]map[uint32]int
	packetSendStatsTimeStamp []time.Time
	mu                       sync.Mutex
	currentFiveSecondStats   map[uint32]int
	cmdMapLastInterval       []timeCmdPair
	cmdMapThisInterval       []timeCmdPair
}

func NewPacketOutAnalyzer(writeFn func(p packet.Packet) (err error)) *PacketOutAnalyzer {
	p := &PacketOutAnalyzer{
		WriteFn:                  writeFn,
		packetSendStats:          make(map[time.Time]map[uint32]int),
		packetSendStatsTimeStamp: make([]time.Time, 0),
		mu:                       sync.Mutex{},
		currentFiveSecondStats:   make(map[uint32]int),
		cmdMapLastInterval:       make([]timeCmdPair, 0),
		cmdMapThisInterval:       make([]timeCmdPair, 0),
	}
	go func() {
		t := time.NewTicker(time.Second * 5)
		cmdInfoCounter := 0
		for _ = range t.C {
			cmdInfoCounter++
			t := time.Now()
			p.mu.Lock()
			p.packetSendStatsTimeStamp = append(p.packetSendStatsTimeStamp, t)
			p.packetSendStats[t] = p.currentFiveSecondStats
			p.currentFiveSecondStats = make(map[uint32]int)
			if len(p.packetSendStats) > (60 / 5) {
				timeStampToDrop := p.packetSendStatsTimeStamp[0]
				p.packetSendStatsTimeStamp = p.packetSendStatsTimeStamp[1:]
				delete(p.packetSendStats, timeStampToDrop)
			}
			if cmdInfoCounter > 12 {
				cmdInfoCounter = 0
				p.cmdMapLastInterval = p.cmdMapThisInterval
				p.cmdMapThisInterval = make([]timeCmdPair, 0)
			}
			p.mu.Unlock()
		}
	}()
	return p
}

func (o *PacketOutAnalyzer) PrintAnalysis() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	infoStr := "最近一分钟机器人发送数据包的统计信息:\n"
	perStripInfoStrAll := ""
	allPacketsStats := make(map[uint32]int)
	for _, t := range o.packetSendStatsTimeStamp {
		count := 0
		secondsBefore := time.Since(t).Seconds()
		perStripInfoStr := fmt.Sprintf("%.2f ~ %.2f 秒前发送数据包的统计信息:\n", secondsBefore-5, secondsBefore)
		for pkName, pkCount := range o.packetSendStats[t] {
			perStripInfoStr += fmt.Sprintf("\t%v: %v\n", utils.PktIDInvMapping[int(pkName)], pkCount)
			allPacketsStats[pkName] += pkCount
			count += pkCount
		}
		perStripInfoStrAll += perStripInfoStr + fmt.Sprintf("共计 %v 个数据包\n", count)
	}
	count := 0
	for pkName, pkCount := range allPacketsStats {
		infoStr += fmt.Sprintf("\t%v: %v\n", utils.PktIDInvMapping[int(pkName)], pkCount)
		count += pkCount
	}
	infoStr += fmt.Sprintf("共计 %v 个数据包\n", count) + perStripInfoStrAll
	return infoStr
}

func (o *PacketOutAnalyzer) GenSendedCmdList() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	records := make([]string, 0)
	for _, p := range o.cmdMapLastInterval {
		sendTime := time.Since(p.time).Seconds()
		records = append(records, fmt.Sprintf("%.2f秒前发送指令: %v", sendTime, p.cmd))
	}
	for _, p := range o.cmdMapThisInterval {
		sendTime := time.Since(p.time).Seconds()
		records = append(records, fmt.Sprintf("%.2f秒前发送指令: %v", sendTime, p.cmd))
	}
	return strings.Join(records, "\n")
}

func (o *PacketOutAnalyzer) Write(p packet.Packet) error {
	o.mu.Lock()
	if p.ID() == packet.IDCommandRequest {
		o.cmdMapThisInterval = append(o.cmdMapThisInterval, timeCmdPair{time.Now(), p.(*packet.CommandRequest).CommandLine})
	}
	o.currentFiveSecondStats[p.ID()]++
	o.mu.Unlock()
	return o.WriteFn(p)
}

type GameCtrl struct {
	analyzer            *PacketOutAnalyzer
	WriteBytesFn        func([]byte) error
	WriteFn             func(packet packet.Packet) error
	ExpectedCmdFeedBack bool
	CurrentCmdFeedBack  bool
	CmdFeedBackOnSent   bool
	NeedFeedBackPackets []packet.Packet
	uuidMaps            map[string]func(*packet.CommandOutput)
	uuidLock            sync.Mutex
	uq                  *uqHolder.UQHolder
	perPlayerStorage    map[string]*PlayerKitOmega
	//playerNameDB        defines.NoSqlDB
	//playerStorageDB     defines.NoSqlDB
	PlayerPermission      map[string]map[string]bool
	onBlockActorCbs       map[define.CubePos]func(define.CubePos, *packet.BlockActorData)
	placeCommandBlockLock sync.Mutex
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

func (g *GameCtrl) SendCmdAndInvokeOnResponseWithFeedback(cmd string, cb func(*packet.CommandOutput)) {
	if !g.CurrentCmdFeedBack && !g.CmdFeedBackOnSent {
		g.turnOnFeedBack()
	}
	ud, _ := uuid.NewUUID()
	g.uuidLock.Lock()
	g.uuidMaps[ud.String()] = cb
	g.uuidLock.Unlock()
	pkt := g.packCmdWithUUID(cmd, ud, false)
	if g.CurrentCmdFeedBack {
		g.WriteFn(pkt)
	} else {
		g.NeedFeedBackPackets = append(g.NeedFeedBackPackets)
	}
	g.WriteFn(pkt)
}

func (g *GameCtrl) SendCmdAndInvokeOnResponse(cmd string, cb func(*packet.CommandOutput)) {
	//if !g.CurrentCmdFeedBack && !g.CmdFeedBackOnSent {
	//	g.turnOnFeedBack()
	//}
	ud, _ := uuid.NewUUID()
	g.uuidLock.Lock()
	g.uuidMaps[ud.String()] = cb
	g.uuidLock.Unlock()
	pkt := g.packCmdWithUUID(cmd, ud, true)
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
	if line == "" {
		return
	}
	content := ToJsonRawString(line)
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

func (g *GameCtrl) ActionBarTo(target string, line string) {
	content := ToJsonRawString(line)
	if strings.HasPrefix(target, "@") {
		g.SendWOCmd(fmt.Sprintf("titleraw %v actionbar %v", target, content))
	} else {
		g.SendWOCmd(fmt.Sprintf("titleraw \"%v\" actionbar %v", target, content))
	}

}

func (g *GameCtrl) TitleTo(target string, line string) {
	content := ToJsonRawString(line)
	if strings.HasPrefix(target, "@") {
		g.SendCmd(fmt.Sprintf("titleraw %v title %v", target, content))
	} else {
		g.SendCmd(fmt.Sprintf("titleraw \"%v\" title %v", target, content))
	}
}

func (g *GameCtrl) SubTitleTo(target string, line string) {
	content := ToJsonRawString(line)
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
	g.CurrentCmdFeedBack = true
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
	if cb, found := g.onBlockActorCbs[pos]; found {
		delete(g.onBlockActorCbs, pos)
		cb(pos, p)
	}
}

func (g *GameCtrl) PlaceCommandBlock(pos define.CubePos, commandBlockName string, commandBlockData int,
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
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], strings.Replace(commandBlockName, "minecraft:", "", 1), commandBlockData)
		g.SendWOCmd(cmd)
		g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
			go func() {
				g.placeCommandBlockLock.Lock()
				g.SendCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
				time.Sleep(50 * time.Millisecond)
				g.SendMCPacket(updatePacket)
				g.placeCommandBlockLock.Unlock()
				g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
					onDone(true)
					done <- true
				}
			}()
		}

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

func (g *GameCtrl) SendBytes(data []byte) {
	g.WriteBytesFn(data)
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
	analyzer := NewPacketOutAnalyzer(o.adaptor.Write)

	c := &GameCtrl{
		WriteFn:             analyzer.Write,
		WriteBytesFn:        o.adaptor.WriteBytes,
		ExpectedCmdFeedBack: o.OmegaConfig.CommandFeedBackByDefault,
		CurrentCmdFeedBack:  false,
		uuidLock:            sync.Mutex{},
		uuidMaps:            make(map[string]func(output *packet.CommandOutput)),
		NeedFeedBackPackets: make([]packet.Packet, 0),
		uq:                  o.uqHolder,
		perPlayerStorage:    make(map[string]*PlayerKitOmega),
		//playerNameDB:        o.GetNoSqlDB("playerNameDB"),
		//playerStorageDB:     o.GetNoSqlDB("playerStorageDB"),
		onBlockActorCbs:       make(map[define.CubePos]func(define.CubePos, *packet.BlockActorData)),
		placeCommandBlockLock: sync.Mutex{},
	}
	c.analyzer = analyzer

	err := o.GetJsonData("playerPermission.json", &c.PlayerPermission)
	if err != nil {
		panic(err)
	}
	if c.PlayerPermission == nil {
		c.PlayerPermission = map[string]map[string]bool{}
	}
	o.CloseFns = append(o.CloseFns, func() error {
		fmt.Println("正在保存 playerPermission.json")
		return o.WriteJsonData("playerPermission.json", c.PlayerPermission)
	})
	c.toExpectedFeedBackStatus()
	c.SendCmd("gamemode c @s")
	return c
}

func (o *Omega) GetGameControl() defines.GameControl {
	return o.GameCtrl
}
