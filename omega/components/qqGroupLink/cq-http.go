package qqGroupLink

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

//type Group struct {
//	GID   int64  `json:"群号"`
//	GName string `json:"游戏显示出的群名"`
//}

type QGroupLink struct {
	Frame                     defines.MainFrame
	Address                   string                        `json:"CQHTTP正向Websocket代理地址"`
	GameMessageFormat         string                        `json:"游戏消息格式化模版"`
	QQMessageFormat           string                        `json:"Q群消息格式化模版"`
	Groups                    map[string]int64              `json:"链接的Q群"`
	Selector                  string                        `json:"游戏内可以听到QQ消息的玩家的选择器"`
	NoBotMsg                  bool                          `json:"不要转发机器人的消息"`
	ChatOnly                  bool                          `json:"只转发聊天消息"`
	MuteIgnored               bool                          `json:"屏蔽其他群的消息"`
	FilterQQToServerMsgByHead string                        `json:"仅仅转发开头为以下特定字符的消息到服务器"`
	FilterServerToQQMsgByHead string                        `json:"仅仅转发开头为以下特定字符的消息到QQ"`
	AllowedCmdExecutor        map[int64]bool                `json:"允许这些人透过QQ执行命令"`
	AllowdFakeCmdExecutor     map[int64]map[string][]string `json:"允许这些人透过QQ执行伪命令"`
	DenyCmds                  map[string]string             `json:"屏蔽这些指令"`
	upgrader                  *websocket.Upgrader
	conn                      *websocket.Conn
	connectLock               chan int
	initLock                  chan int
	inited                    bool
	firstInit                 bool
	sendChan                  chan string
	connectionFalseHintReduce int
	cqStartPrintErr           bool
	loginTruncated            bool
}

func (cq *QGroupLink) cqStartPrintErrRoutine() {
	if cq.cqStartPrintErr {
		return
	}
	cq.cqStartPrintErr = true
	for {
		pterm.Error.Println("Q群链接组件: 和CQ-HTTP连接出现故障, 请排除错误并重启 Omega ")
		<-time.NewTimer(time.Minute * 2).C
	}
}

// receiveRoutine 接收并处理协议端的消息 from QQ
func (cq *QGroupLink) receiveRoutine() {
	fmt.Println("CQ-CHAT: Receive Routine Start")
	for {
		_, data, err := cq.conn.ReadMessage()
		if err != nil {
			cq.Frame.GetBackendDisplay().Write(fmt.Sprintf("Q群链接组件: 和CQ-HTTP连接出现故障:" + err.Error()))
			cq.cqStartPrintErrRoutine()
			return
			// cq.Frame.GetBackendDisplay().Write(fmt.Sprintf("10秒后重连"))
			time.Sleep(10 * time.Second)
			cq.conn.Close()
			// 如果发送协程还没有尝试重连，那么由发送线程尝试重连
			if cq.inited {
				cq.connect()
			} else {
				<-cq.connectLock
			}
		}
		// 先解析出事件种类(event或message)
		post, err := ParseMetaPost(data)
		if post.PostType == "meta_event" && post.MetaEventType == "lifecycle" {
			cq.Frame.GetBackendDisplay().Write(fmt.Sprintf("Q群链接组件: 已成功连接"))
			if !cq.inited {
				cq.inited = true
				close(cq.initLock)
			}

		}
		if post.PostType == "message" && err == nil {
			action, err := GetMessageData(data)
			if err != nil || action == nil {
				continue
			}
			if msg, succ := action.(GroupMessage); succ {
				cq.onNewQQMessage(msg)
			}
		}
		continue
	}
}

// SendMessage
func (cq *QGroupLink) sendRoutine() {
	<-cq.initLock
	lastSend := ""
	for {
		lastSend = <-cq.sendChan
		echo, _ := uuid.NewUUID()
		for _, gid := range cq.Groups {
			qmsg := QMessage{
				Action: "send_group_msg",
				Params: struct {
					GroupID int64  `json:"group_id"`
					Message string `json:"message"`
				}{
					GroupID: gid,
					Message: lastSend,
				},
				Echo: echo.String(),
			}
			data, _ := json.Marshal(qmsg)
			err := cq.conn.WriteMessage(1, data)
			if err != nil {
				cq.conn.Close()
				cq.cqStartPrintErrRoutine()
				return
				// 如果接收协程还没有尝试重连，那么由发送线程尝试重连
				if cq.inited {
					cq.connect()
				}
				<-cq.initLock
			}
		}
	}
}

func (cq *QGroupLink) connect() chan int {
	cq.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	for {
		cq.inited = false
		cq.initLock = make(chan int)
		cq.connectLock = make(chan int)
		u := url.URL{Scheme: "ws", Host: cq.Address}
		var err error
		cq.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			if cq.firstInit {
				panic(cq.firstInitErr(err))
			} else {
				cq.Frame.GetBackendDisplay().Write(fmt.Sprintf("Q群链接组件: 和CQ-HTTP连接出现故障:" + err.Error()))
			}
		} else {
			close(cq.connectLock)
			break
		}
	}
	if cq.firstInit {
		go cq.receiveRoutine()
		go cq.sendRoutine()
	}
	cq.firstInit = false
	return cq.initLock
}

func (cq *QGroupLink) InterceptCmdMsg(uid int64, msg string) (stop bool) {
	// 接收的消息含有HTML转义字符，在使用前需要进行处理，否则诸如 /testfor @a[tag=a] 的指令会出现错误
	msg = html.UnescapeString(msg)
	cmds := []string{}
	if allowed, hasK := cq.AllowedCmdExecutor[uid]; hasK && allowed && strings.HasPrefix(msg, "/") {
		if strings.HasPrefix(msg, "/权限") {
			hint := "完整 MC 指令权限:"
			for qq, hasPerm := range cq.AllowedCmdExecutor {
				if hasPerm {
					hint += fmt.Sprintf("\n%v", qq)
				}
			}
			if len(cq.AllowdFakeCmdExecutor) > 0 {
				hint += "\n转义 MC 指令权限: "
				for qq, auths := range cq.AllowdFakeCmdExecutor {
					hint += fmt.Sprintf("\n%v:", qq)
					for k, _ := range auths {
						hint += "\n   -" + k
					}
				}
			}
			cq.sendQQMessage(hint)
			return true
		}
	}
	if hasPerm, hasK := cq.AllowdFakeCmdExecutor[uid]; hasK {
		for perm, tmps := range hasPerm {
			if !strings.HasPrefix(perm, "/") {
				perm = "/" + perm
			}
			if strings.HasPrefix(msg, perm) {
				args := msg[len(perm):]
				for _, tmp := range tmps {
					cmds = append(cmds, strings.ReplaceAll(tmp, "[args]", args))
				}
			}
		}
	}
	if len(cmds) == 0 {
		if allowed, hasK := cq.AllowedCmdExecutor[uid]; hasK && allowed && strings.HasPrefix(msg, "/") {
			for cmd, resp := range cq.DenyCmds {
				if strings.Contains(msg, cmd) {
					cq.sendQQMessage(resp)
					return true
				}
			}
			cmds = append(cmds, msg)
		}
	}
	if len(cmds) == 0 {
		return strings.HasPrefix(msg, "/")
	} else {
		result := ""
		cmdI := 1
		var sendNext func(thisCmd string, nextCmds []string)
		sendNext = func(thisCmd string, nextCmds []string) {
			cq.Frame.GetGameControl().SendCmdAndInvokeOnResponse(thisCmd, func(output *packet.CommandOutput) {
				if len(nextCmds) == 0 && cmdI == 1 {
					if output.SuccessCount > 0 {
						result += thisCmd + " ✓"
					} else {
						result += thisCmd + " ✗"
					}
				} else {
					if output.SuccessCount > 0 {
						result += fmt.Sprintf("%v. %v ", cmdI, thisCmd) + "✓"
					} else {
						result += fmt.Sprintf("%v. %v ", cmdI, thisCmd) + "✗"
					}
				}
				cmdI++
				if len(output.OutputMessages) > 0 {
					result += "\n---"
				}
				for _, r := range output.OutputMessages {
					if r.Success {
						result += "\n✓ "
					} else {
						result += "\n✗ "
					}
					result += r.Message + " " + fmt.Sprintf("%v", r.Parameters)
				}
				if len(nextCmds) > 0 {
					_next := nextCmds[1:]
					result += "\n\n"
					sendNext(nextCmds[0], _next)
				} else {
					cq.sendQQMessage(result)
				}
			})
		}
		sendNext(cmds[0], cmds[1:])
		return true
	}
}

func (cq *QGroupLink) onNewQQMessage(msg IMessage) {
	if msg.ID() != IDGroupMessage {
		return
	}
	groupMsg := msg.(GroupMessage)
	gid := groupMsg.GroupID
	msgText := groupMsg.Message
	if cq.FilterQQToServerMsgByHead != "" && !strings.HasPrefix(msgText, "/") {
		if !strings.HasPrefix(msgText, cq.FilterQQToServerMsgByHead) {
			return
		}
	}
	msgText = GetRawTextFromCQMessage(msgText)
	qqUserName := groupMsg.Sender.Nickname

	for gname, sourceGid := range cq.Groups {
		if sourceGid == gid {
			uid := groupMsg.PrivateMessage.UserId
			if cq.InterceptCmdMsg(uid, msgText) {
				return
			}
			m := utils.FormatByReplacingOccurrences(cq.QQMessageFormat, map[string]interface{}{
				"[groupName]":  gname,
				"[QQUserName]": qqUserName,
				"[msg]":        msgText,
			})
			cq.Frame.GetBackendDisplay().Write("QQ->MC: " + m)
			m = strings.Replace(m, "[Error]:", "", 1)
			cq.Frame.GetGameControl().SayTo(cq.Selector, m)
			return
		}
	}
	if !cq.MuteIgnored {
		pterm.Warning.Sprintln("来自群 %v 的消息: %v 被忽略（因为配置中没有指明需要转发该群消息到游戏）", gid, msgText)
	}

}

func (cq *QGroupLink) sendQQMessage(msg string) {
	cq.sendChan <- msg
}

func (cq *QGroupLink) onNewGameMsg(chat *defines.GameChat) bool {
	if cq.cqStartPrintErr {
		return false
	}
	if cq.ChatOnly && chat.Type != packet.TextTypeChat {
		return false
	}
	if cq.Frame.GetUQHolder().GetBotName() == chat.Name && cq.NoBotMsg {
		return false
	}
	if chat.FrameWorkTriggered {
		return false
	}
	msgText := strings.Join(chat.Msg, " ")
	if cq.FilterServerToQQMsgByHead != "" {
		if !strings.HasPrefix(msgText, cq.FilterServerToQQMsgByHead) {
			return false
		}
	}
	msg := utils.FormatByReplacingOccurrences(cq.GameMessageFormat, map[string]interface{}{
		"[player]": chat.Name,
		"[msg]":    msgText,
	})
	cq.Frame.GetBackendDisplay().Write("MC->QQ: " + msg)
	cq.sendQQMessage(msg)
	return false
}

//go:embed how-to.txt
var help []byte

func (cq *QGroupLink) firstInitErr(err error) error {
	pterm.Error.Println("首次连接到 CQ-HTTP 时出现错误" + err.Error())
	fmt.Println(string(help))
	pterm.Info.Println("请按上述说明配置CQ-HTTP，完成后重启 Omega系统，但是保持 CQ-HTTP 运行")
	pterm.Info.Println("或者，建议的是，你可以使用 Omega 启动器 以自动配置群服互通，又简单又快捷")
	pterm.Error.Println("如果您不主动关闭 Omega，3 分钟后将在没有群服互通功能的情况下继续运行 Omega")
	time.Sleep(3 * time.Minute)
	// return fmt.Errorf("群服互联插件需要配置")
	return nil
}

func (b *QGroupLink) Init(cfg *defines.ComponentConfig) {
	b.AllowedCmdExecutor = map[int64]bool{}
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	b.sendChan = make(chan string, 64)
	b.firstInit = true
}

func (b *QGroupLink) Inject(frame defines.MainFrame) {
	b.Frame = frame
	b.connect()
	<-b.initLock
	b.Frame.GetBackendDisplay().Write("Q群链接组件: 连接成功")
	var collaborate_func collaborate.FUNC_SEND_TO_GROUP
	collaborate_func = func(msg string) {
		b.sendQQMessage(msg)
	}
	b.Frame.SetContext(collaborate.INTERFACE_SEND_TO_GROUP, collaborate_func)
	hint := "[群服互通]: 连接成功"
	if b.FilterQQToServerMsgByHead != "" {
		hint += "\n QQ->MC: 消息开头必须为" + b.FilterQQToServerMsgByHead
	}
	if b.FilterServerToQQMsgByHead != "" {
		hint += "\n MC->QQ: 消息开头必须为" + b.FilterServerToQQMsgByHead
	}
	if len(b.AllowedCmdExecutor) > 0 {
		perms := []string{}
		for qq, hasPerm := range b.AllowedCmdExecutor {
			if hasPerm {
				shortQQ := fmt.Sprintf("%v", qq)
				if len(shortQQ) > 4 {
					shortQQ = shortQQ[:4] + ".."
				}
				perms = append(perms, shortQQ)
			}
		}
		if len(perms) > 0 {
			hint += "\n MC 指令: \n" + strings.Join(perms, ",") + "\n输入 /权限 查看完整权限"
		}
	}
	b.sendQQMessage(hint)
	b.Frame.GetGameListener().SetGameChatInterceptor(b.onNewGameMsg)
	b.Frame.GetGameListener().AppendLogoutInfoCallback(func(entry protocol.PlayerListEntry) {
		player := b.Frame.GetGameControl().GetPlayerKitByUUID(entry.UUID)
		if player != nil {
			b.sendQQMessage(fmt.Sprintf("%v 离开了游戏", player.GetRelatedUQ().Username))
		}
	})
	b.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		if b.loginTruncated {
			name := utils.ToPlainName(entry.Username)
			b.sendQQMessage(fmt.Sprintf("%v 进入了游戏", name))
		}
	})
}

func (b *QGroupLink) Stop() error {
	b.sendQQMessage("[群服互通]: 机器人已经退出服务器")
	return nil
}

func (b *QGroupLink) Signal(signal int) error {
	return nil
}

func (b *QGroupLink) BeforeActivate() error {
	return nil
}

func (b *QGroupLink) Activate() {
	time.Sleep(time.Second * 5)
	b.loginTruncated = true
}
