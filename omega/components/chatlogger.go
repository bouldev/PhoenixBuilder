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
	cl.logger = cl.Frame.GetLogger("聊天记录.log")
	cl.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDText, func(p packet.Packet) {
		pk := p.(*packet.Text)
		msg := strings.TrimSpace(pk.Message)
		msg = fmt.Sprintf("[%v] %v:%v", pk.TextType, pk.SourceName, msg)
		if len(pk.Parameters) != 0 {
			msg += " (" + strings.Join(pk.Parameters, ", ") + ")"
		}
		cl.logger.Write(msg)
	})
}
