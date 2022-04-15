package components

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strings"
)

type ChatLogger struct {
	*BasicComponent
	logger defines.LineDst
}

func (cl *ChatLogger) Inject(frame defines.MainFrame) {
	cl.Frame = frame
	cl.logger = cl.Frame.GetLogger("chat.log")
	cl.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDText, func(p packet.Packet) {
		pk := p.(*packet.Text)
		cl.logger.Write(fmt.Sprintf("[%v] %v:%v (%v)", pk.TextType, pk.SourceName, strings.TrimSpace(pk.Message), pk.Parameters))
	})
}
