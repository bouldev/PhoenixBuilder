package qqGroupLink

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"phoenixbuilder/minecraft/protocol/packet"
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
	Address                   string            `json:"CQHTTP正向Websocket代理地址"`
	GameMessageFormat         string            `json:"游戏消息格式化模版"`
	QQMessageFormat           string            `json:"Q群消息格式化模版"`
	Groups                    map[string]int64  `json:"链接的Q群"`
	Selector                  string            `json:"游戏内可以听到QQ消息的玩家的选择器"`
	NoBotMsg                  bool              `json:"不要转发机器人的消息"`
	ChatOnly                  bool              `json:"只转发聊天消息"`
	MuteIgnored               bool              `json:"屏蔽其他群的消息"`
	FilterQQToServerMsgByHead string            `json:"仅仅转发开头为以下特定字符的消息到服务器"`
	FilterServerToQQMsgByHead string            `json:"仅仅转发开头为以下特定字符的消息到QQ"`
	AllowedCmdExecutor        map[int64]bool    `json:"允许这些人透过QQ执行命令"`
	DenyCmds                  map[string]string `json:"屏蔽这些指令"`
	upgrader                  *websocket.Upgrader
	conn                      *websocket.Conn
	connectLock               chan int
	initLock                  chan int
	inited                    bool
	firstInit                 bool
	sendChan                  chan string
}

// receiveRoutine 接收并处理协议端的消息 from QQ
func (cq *QGroupLink) receiveRoutine() {
	fmt.Println("CQ-CHAT: Receive Routine Start")
	for {
		_, data, err := cq.conn.ReadMessage()
		if err != nil {
			cq.Frame.GetBackendDisplay().Write(fmt.Sprintf("Q群链接组件: 和CQ-HTTP连接出现故障:" + err.Error()))
			cq.Frame.GetBackendDisplay().Write(fmt.Sprintf("10秒后重连"))
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

func (cq *QGroupLink) onNewQQMessage(msg IMessage) {
	if msg.ID() != IDGroupMessage {
		return
	}
	groupMsg := msg.(GroupMessage)
	gid := groupMsg.GroupID
	msgText := groupMsg.Message
	if cq.FilterQQToServerMsgByHead != "" && !strings.HasPrefix(msgText,"/") {
		if !strings.HasPrefix(msgText, cq.FilterQQToServerMsgByHead) {
			return
		}
	}
	msgText = GetRawTextFromCQMessage(msgText)
	qqUserName := groupMsg.PrivateMessage.Sender.Nickname

	for gname, sourceGid := range cq.Groups {
		if sourceGid == gid {
			uid := groupMsg.PrivateMessage.UserId
			if allowed, hasK := cq.AllowedCmdExecutor[uid]; hasK && allowed && strings.HasPrefix(msgText, "/") {
				for cmd, resp := range cq.DenyCmds {
					if strings.Contains(msgText, cmd) {
						cq.sendQQMessage(resp)
						return
					}
				}

				cq.Frame.GetGameControl().SendCmdAndInvokeOnResponse(msgText, func(output *packet.CommandOutput) {
					result := ""
					if output.SuccessCount > 0 {
						result += "执行成功✓\n---"
					} else {
						result += "执行失败✗\n---"
					}
					for _, r := range output.OutputMessages {
						if r.Success {
							result += "\n✓ "
						} else {
							result += "\n✗ "
						}
						result += r.Message + " " + fmt.Sprintf("%v", r.Parameters)
					}
					cq.sendQQMessage(result)
				})

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
	return fmt.Errorf("群服互联插件需要配置")
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
	b.Frame.GetGameListener().SetGameChatInterceptor(b.onNewGameMsg)
}

func (b *QGroupLink) Stop() error {
	return nil
}

func (b *QGroupLink) Activate() {

}
