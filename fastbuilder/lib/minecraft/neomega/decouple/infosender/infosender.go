package infosender

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
)

func init() {
	if false {
		func(sender omega.InfoSender) {}(&InfoSender{})
	}
}

type InfoSender struct {
	omega.InteractCore
	omega.CmdSender
	omega.BotBasicInfoHolder
}

func NewInfoSender(interactable omega.InteractCore, cmdSender omega.CmdSender, info omega.BotBasicInfoHolder) omega.InfoSender {
	return &InfoSender{
		InteractCore:       interactable,
		CmdSender:          cmdSender,
		BotBasicInfoHolder: info,
	}
}

func (i *InfoSender) BotSay(msg string) {
	pk := &packet.Text{
		TextType:         packet.TextTypeChat,
		NeedsTranslation: false,
		SourceName:       i.GetBotName(),
		Message:          msg,
		XUID:             "",
		PlayerRuntimeID:  fmt.Sprintf("%d", i.GetBotRuntimeID()),
	}
	i.SendPacket(pk)
}

func (i *InfoSender) SayTo(target string, msg string) {
	//TODO implement me
	panic("implement me")
}

func (i *InfoSender) RawSayTo(target string, msg string) {
	//TODO implement me
	panic("implement me")
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

func (i *InfoSender) ActionBarTo(target string, msg string) {
	content := toJsonRawString(msg)
	if strings.HasPrefix(target, "@") {
		i.SendWOCmd(fmt.Sprintf("titleraw %v actionbar %v", target, content))
	} else {
		i.SendWOCmd(fmt.Sprintf("titleraw \"%v\" actionbar %v", target, content))
	}
}

func (i *InfoSender) TitleTo(target string, msg string) {
	content := toJsonRawString(msg)
	if strings.HasPrefix(target, "@") {
		i.SendWSCmd(fmt.Sprintf("titleraw %v title %v", target, content))
	} else {
		i.SendWSCmd(fmt.Sprintf("titleraw \"%v\" title %v", target, content))
	}
}

func (i *InfoSender) SubTitleTo(target string, msg string) {
	content := toJsonRawString(msg)
	if strings.HasPrefix(target, "@") {
		i.SendWSCmd(fmt.Sprintf("titleraw %v subtitle %v", target, content))
	} else {
		i.SendWSCmd(fmt.Sprintf("titleraw \"%v\" subtitle %v", target, content))
	}
}
