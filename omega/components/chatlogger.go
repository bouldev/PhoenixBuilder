package components

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

type ChatLogger struct {
	*BasicComponent
	logger defines.LineDst
}

func (cl *ChatLogger) Inject(frame defines.MainFrame) {
	cl.Frame = frame
	cl.logger = &utils.MultipleLogger{
		Loggers: []defines.LineDst{
			cl.Frame.GetLogger("聊天记录.log"),
			cl.Frame.GetBackendDisplay(),
		},
	}
	// botName := cl.Frame.GetUQHolder().GetBotName()
	cl.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDText, func(p packet.Packet) {
		pk := p.(*packet.Text)
		//TODO don't do this
		msg := strings.TrimSpace(pk.Message)
		// fmt.Println(msg)
		if strings.Contains(msg, "alive") {
			return
		}
		_l := len(msg)
		if _l > 200 {
			msg = msg[:200] + fmt.Sprintf("...[还有%v字]", _l-200)
		}
		msg = fmt.Sprintf("[%v] %v:%v", pk.TextType, pk.SourceName, msg)
		if len(pk.Parameters) != 0 {
			msg += " (" + strings.Join(pk.Parameters, ", ") + ")"
		}
		cl.logger.Write(msg)
	})
}
