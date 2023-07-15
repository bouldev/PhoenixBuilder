package cmdsender

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(sender omega.CmdSender) {}(&CmdSender{})
	}
}

type CmdSender struct {
	omega.InteractCore
	cbByUUID            sync.Map
	expectedCmdFeedBack bool
	currentCmdFeedBack  bool
	cmdFeedBackOnSent   bool
	needFeedBackPackets []packet.Packet
}

type Options struct {
	ExpectedCmdFeedBack bool
}

func MakeDefaultCmdSenderOption() *Options {
	return &Options{ExpectedCmdFeedBack: false}
}

func NewCmdSender(reactable omega.ReactCore, interactable omega.InteractCore, option *Options) omega.CmdSender {
	if option == nil {
		option = MakeDefaultCmdSenderOption()
	}
	c := &CmdSender{
		InteractCore:        interactable,
		cbByUUID:            sync.Map{},
		expectedCmdFeedBack: option.ExpectedCmdFeedBack,
		currentCmdFeedBack:  false,
		cmdFeedBackOnSent:   false,
		needFeedBackPackets: make([]packet.Packet, 0),
	}
	reactable.SetOnTypedPacketCallBack(packet.IDGameRulesChanged, func(p packet.Packet) {
		for _, rule := range p.(*packet.GameRulesChanged).GameRules {
			if rule.Name == "sendcommandfeedback" {
				if rule.Value == true {
					c.onCommandFeedbackOn()
				} else {
					c.onCommandFeedBackOff()
				}
			}
		}
	})
	reactable.SetOnTypedPacketCallBack(packet.IDCommandOutput, func(p packet.Packet) {
		// fmt.Println(p)
		c.onNewCommandFeedBack(p.(*packet.CommandOutput))
	})
	return c
}

func (c *CmdSender) SendWSCmd(cmd string) {
	ud, _ := uuid.NewUUID()
	c.SendPacket(c.packCmdWithUUID(cmd, ud, true))
}

func (c *CmdSender) SendCmdWithUUID(cmd string, ud uuid.UUID, ws bool) {
	c.SendPacket(c.packCmdWithUUID(cmd, ud, ws))
}

func (c *CmdSender) SendWOCmd(cmd string) {
	c.SendPacket(&packet.SettingsCommand{
		CommandLine:    cmd,
		SuppressOutput: true,
	})
}

func (c *CmdSender) setCB(uuidStr string, cb func(output *packet.CommandOutput)) {
	c.cbByUUID.Store(uuidStr, cb)
}

func (c *CmdSender) SendWSCmdAndInvokeOnResponse(cmd string, cb func(output *packet.CommandOutput)) {
	ud, _ := uuid.NewUUID()
	c.setCB(ud.String(), cb)
	pkt := c.packCmdWithUUID(cmd, ud, true)
	c.SendPacket(pkt)
}

func (c *CmdSender) SendPlayerCmd(cmd string) {
	ud, _ := uuid.NewUUID()
	pkt := c.packCmdWithUUID(cmd, ud, false)
	c.SendPacket(pkt)
	c.SendPacket(pkt)
}

func (c *CmdSender) SendPlayerCmdAndInvokeOnResponseWithFeedback(cmd string, cb func(output *packet.CommandOutput)) {
	if !c.currentCmdFeedBack && !c.cmdFeedBackOnSent {
		c.turnOnFeedBack()
	}
	ud, _ := uuid.NewUUID()
	c.setCB(ud.String(), cb)
	pkt := c.packCmdWithUUID(cmd, ud, false)
	if c.currentCmdFeedBack {
		c.SendPacket(pkt)
	} else {
		c.needFeedBackPackets = append(c.needFeedBackPackets)
	}
	c.SendPacket(pkt)
}

func (c *CmdSender) packCmdWithUUID(cmd string, ud uuid.UUID, ws bool) *packet.CommandRequest {
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

func (c *CmdSender) onCommandFeedbackOn() {
	// fmt.Println("recv sendcommandfeedback true")
	c.currentCmdFeedBack = true
	c.cmdFeedBackOnSent = false
	pkts := c.needFeedBackPackets
	c.needFeedBackPackets = make([]packet.Packet, 0)
	for _, p := range pkts {
		c.SendPacket(p)
	}
	if !c.expectedCmdFeedBack {
		c.turnOffFeedBack()
	}
}

func (c *CmdSender) onCommandFeedBackOff() {
	if c.expectedCmdFeedBack {
		c.turnOnFeedBack()
	}
}

func (c *CmdSender) turnOnFeedBack() {
	//fmt.Println("send sendcommandfeedback true")
	c.SendWSCmd("gamerule sendcommandfeedback true")
	c.cmdFeedBackOnSent = true
}

func (c *CmdSender) turnOffFeedBack() {
	c.currentCmdFeedBack = false
	c.cmdFeedBackOnSent = false
	//fmt.Println("send sendcommandfeedback false")
	c.SendWSCmd("gamerule sendcommandfeedback false")
}

func (c *CmdSender) onNewCommandFeedBack(p *packet.CommandOutput) {
	s := p.CommandOrigin.UUID.String()
	cb, ok := c.cbByUUID.LoadAndDelete(s)
	if ok {
		cb.(func(output *packet.CommandOutput))(p)
	}
}
