package QGuildLink

import (
	"container/list"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

type QGuildLink struct {
	*defines.BasicComponent
	ChatOnly            bool                        `json:"只转发聊天消息"`
	NoBotMsg            bool                        `json:"不要转发机器人的消息"`
	SendJoinAndLeaveMsg bool                        `json:"向频道发送玩家进出消息"`
	ShowExchangeDetail  bool                        `json:"在控制台显示消息转发详情"`
	GameMessageFormat   string                      `json:"游戏消息格式化模版"`
	QGuildMessageFormat string                      `json:"频道消息格式化模版"`
	Address             string                      `json:"CQHTTP正向Websocket代理地址"`
	Selector            string                      `json:"游戏内可以听到QQ消息的玩家的选择器"`
	ServerToQQMsgFilter string                      `json:"仅仅转发开头为以下特定字符的消息到QQ"`
	QQToServerMsgFilter string                      `json:"仅仅转发开头为以下特定字符的消息到服务器"`
	DenyCmds            []string                    `json:"不允许执行这些指令"`
	PublicCmds          []string                    `json:"允许所有频道成员使用这些指令"`
	CmdExecutor         []string                    `json:"允许这些身份组的频道成员透过QQ执行指令"`
	LinkChannelNames    map[string][]string         `json:"链接的子频道"`
	CmdsExecutorIn      map[string]map[string][]any `json:"允许这些身份组的频道成员透过QQ执行指令序列"`
	conn                *websocket.Conn
	mu                  *sync.Mutex
	connecting          bool
	sendChan            chan CQRequest
	guildsName          map[string]string
	channelsName        map[string]string
	channelIDs          map[string][]string
	roleMap             map[string]map[string]*userRoles
	cmdsExecutor        map[string]map[string][]defines.Cmd
	awaitCmdQueue       *list.List
}

// 为获取的身份组列表设置过期时间, 届时会重新向go-cqhttp请求获取最新的身份组列表
type userRoles struct {
	roles  []string
	expire time.Time
}

// 尝试与go-cqhttp建立连接, 并返回结果(err)
func (o *QGuildLink) connect() (err error) {
	url := url.URL{Scheme: "ws", Host: o.Address}
	o.conn, _, err = websocket.DefaultDialer.Dial(url.String(), nil)
	if err == nil {
		go o.receiveRoutine()
		go o.sendRoutine()
		o.connecting = true
	}
	return err
}

// 从go-cqhttp接收数据包
func (o *QGuildLink) receiveRoutine() {
	for {
		_, pkt, err := o.conn.ReadMessage()
		if err != nil {
			pterm.Error.Println("[频服互通]: 与 go-cqhttp 的连接出现故障, 将尝试重连")
			pterm.Error.Println(err)
			o.connecting = false
			o.conn.Close()
			// 启动重连协程
			go func() {
				count := 0
				for {
					if count++; count > 10 {
						pterm.Error.Println("[频服互通]: 无法与 go-cqhttp 重新建立连接, 请重新配置后重启 Omega")
						return
					}
					pterm.Warning.Printfln("[频服互通]: 尝试与 go-cqhttp 重新建立连接, 次数：%d", count)
					if err := o.connect(); err == nil {
						return
					}
					<-time.After(time.Second * time.Duration(10))
				}
			}()
			return
		}
		// 先对数据包进行初步解析
		data, err := ParseRecvPkt(pkt)
		if err != nil {
			continue
		}
		// 接收到生命周期消息时, 请求获取频道列表信息
		if data.MetaEventType == "lifecycle" {
			o.Frame.GetBackendDisplay().Write("[频服互通]: 成功建立与 go-cqhttp 的连接")
			if len(o.channelIDs) == 0 {
				o.sendChan <- CQRequest{
					Action: "get_guild_list",
					Echo:   "get_guild_list",
				}
			} else {
				o.sendAllChannelsMessage("[频服互通]: 已重新连接")
			}
			continue
		}
		// 接收到频道列表信息时, 请求获取需要的子频道列表信息
		if data.Echo == "get_guild_list" {
			if data.Status != "ok" || data.Data == nil {
				continue
			}
			if bytes, err := json.Marshal(data.Data); err == nil {
				msg := []GuildInfoResponse{}
				if err = json.Unmarshal(bytes, &msg); err == nil {
					for _, value := range msg {
						if _, ok := o.LinkChannelNames[value.GuildName]; ok {
							o.guildsName[value.GuildID] = value.GuildName
							o.sendChan <- CQRequest{
								Action: "get_guild_channel_list",
								Params: GuildChannelListRequest{
									GuildID: value.GuildID,
								},
								Echo: "get_guild_channel_list|" + value.GuildName,
							}
						}
					}
				}
			}
			continue
		}
		// 接收到子频道列表信息时
		if strings.HasPrefix(data.Echo, "get_guild_channel_list") {
			if data.Status != "ok" || data.Data == nil {
				continue
			}
			guildName := strings.Split(data.Echo, "|")[1]
			if bytes, err := json.Marshal(data.Data); err == nil {
				msg := []GuildChannelListResponse{}
				if err = json.Unmarshal(bytes, &msg); err == nil {
					for _, value := range msg {
						for _, channelName := range o.LinkChannelNames[guildName] {
							if channelName == value.ChannelName {
								o.channelsName[value.ChannelID] = value.ChannelName
								o.channelIDs[value.OwnerGuildID] = append(o.channelIDs[value.OwnerGuildID], value.ChannelID)
							}
						}
					}
					o.sendAllChannelsMessage("[频服互通]: 机器人已成功连接到服务器")
				}
			}
			continue
		}
		// 接收到频道成员信息时; 约定在发送此请求时, 将与其对应的指令请求放在等待队列中
		if strings.HasPrefix(data.Echo, "get_guild_member_profile") {
			// 先从队列中取出待处理的指令消息, 等待重新执行
			front := o.awaitCmdQueue.Front()
			o.awaitCmdQueue.Remove(front)
			// 如果获取失败(例如该成员没有任何身份组), 将放弃对这条消息的处理
			if data.Status != "ok" || data.Data == nil {
				continue
			}
			guildID := strings.Split(data.Echo, "|")[1]
			if bytes, err := json.Marshal(data.Data); err == nil {
				msg := GetGuildMemberProfileResponse{}
				if err = json.Unmarshal(bytes, &msg); err == nil {
					o.roleMap[guildID] = make(map[string]*userRoles)
					ur := &userRoles{
						roles:  []string{},
						expire: time.Now().Add(time.Minute * time.Duration(10)),
					}
					for _, role := range msg.Roles {
						ur.roles = append(ur.roles, role.RoleName)
					}
					o.roleMap[guildID][msg.TinyID] = ur
				}
			}
			o.dealCmdMessage(front.Value.(GuildMsg))
			continue
		}
		// 接收到消息时
		if data.PostType == "message" {
			msg := GuildMsg{}
			err = json.Unmarshal(pkt, &msg)
			if err != nil || msg.MessageType != "guild" {
				continue
			}
			o.onGuildMessage(msg)
		}
	}
}

// 向go-cqhttp发送数据包
func (o *QGuildLink) sendRoutine() {
	for {
		lastSend := <-o.sendChan
		pkt, _ := json.Marshal(lastSend)
		o.mu.Lock()
		err := o.conn.WriteMessage(1, pkt)
		o.mu.Unlock()
		if err != nil {
			o.conn.Close()
			return
		}
	}
}

// 收到游戏内消息时
func (o *QGuildLink) onGameMessage(chat *defines.GameChat) bool {
	if chat.FrameWorkTriggered {
		return false
	}
	if o.ChatOnly && chat.Type != packet.TextTypeChat {
		return false
	}
	if o.Frame.GetUQHolder().GetBotName() == chat.Name && o.NoBotMsg {
		return false
	}
	msgText := strings.Join(chat.Msg, " ")
	if o.ServerToQQMsgFilter != "" {
		if !strings.HasPrefix(msgText, o.ServerToQQMsgFilter) {
			return false
		}
	}
	msg := msgText
	// 昵称不为空时, 对消息进行格式化
	if chat.Name != "" {
		msg = utils.FormatByReplacingOccurrences(o.GameMessageFormat, map[string]interface{}{
			"[player]": chat.Name,
			"[msg]":    msgText,
		})
	}
	// 存在参数时, 向QQ群发送带参数的消息
	if len(chat.RawParameters) > 0 {
		msg = msg + " (" + strings.Join(chat.RawParameters, ", ") + ")"
	}
	if o.ShowExchangeDetail {
		o.Frame.GetBackendDisplay().Write("MC->QQ: " + msg)
	}
	o.sendAllChannelsMessage(msg)
	return false
}

// 收到频道聊天信息时
func (o *QGuildLink) onGuildMessage(msg GuildMsg) {
	// 忽略非指定频道的消息
	check := false
	if channelIDs, ok := o.channelIDs[msg.GuildID]; ok {
		for _, channelID := range channelIDs {
			if msg.ChannelID == channelID {
				check = true
				break
			}
		}
	}
	if !check {
		return
	}
	msg.Message = html.UnescapeString(msg.Message)
	// 非指令消息时
	if !strings.HasPrefix(msg.Message, "/") {
		msg.Message = GetRawTextFromCQMessage(msg.Message)
		msg.Message = utils.FormatByReplacingOccurrences(o.QGuildMessageFormat, map[string]interface{}{
			"[guildName]":   o.guildsName[msg.GuildID],
			"[channelName]": o.channelsName[msg.ChannelID],
			"[nickname]":    msg.Sender.Nickname,
			"[msg]":         msg.Message,
		})
		if o.ShowExchangeDetail {
			o.Frame.GetBackendDisplay().Write("QQ->MC: " + msg.Message)
		}
		o.Frame.GetGameControl().SayTo(o.Selector, msg.Message)
		return
	}
	o.dealCmdMessage(msg)
}

// 处理收到的指令消息
func (o *QGuildLink) dealCmdMessage(msg GuildMsg) {
	// 对指令消息进行切割
	cmd := strings.Split(strings.TrimPrefix(msg.Message, "/"), " ")
	if len(cmd) == 0 {
		return
	}
	// 检查指令是否在不允许列表中
	for _, denyCmd := range o.DenyCmds {
		if cmd[0] == denyCmd {
			o.sendChannelAtMessage(msg.GuildID, msg.ChannelID, msg.UserID, msg.Sender.Nickname, "✗ 该指令不被允许执行")
			return
		}
	}
	// 检查指令是否为公用列表中
	for _, publicCmd := range o.PublicCmds {
		if cmd[0] == publicCmd {
			o.execCmdAndFeedback(msg)
			return
		}
	}
	// 检查身份组是否已被获取
	if ur, ok := o.roleMap[msg.GuildID][msg.UserID]; ok && ur.expire.After(time.Now()) {
		// 检查是否在指令序列列表中
		for _, role := range ur.roles {
			if cmds, ok := o.cmdsExecutor[role][cmd[0]]; ok {
				go utils.LaunchCmdsArray(o.Frame.GetGameControl(), cmds, map[string]interface{}{
					"[args]": strings.Join(cmd[1:], " "),
				}, o.Frame.GetBackendDisplay())
				o.sendChannelAtMessage(msg.GuildID, msg.ChannelID, msg.UserID, msg.Sender.Nickname, "✓ 对应的指令序列已被执行")
				return
			}
		}
		// 检查是否在允许使用指令的身份组中
		for _, role := range ur.roles {
			for _, executor := range o.CmdExecutor {
				if role == executor {
					o.execCmdAndFeedback(msg)
					return
				}
			}
		}
		o.sendChannelAtMessage(msg.GuildID, msg.ChannelID, msg.UserID, msg.Sender.Nickname, "✗ 不具备指令执行权限")
	} else {
		// 请求获取身份组并放入待执行队列
		o.awaitCmdQueue.PushBack(msg)
		// 请求获取该成员身份组
		o.sendChan <- CQRequest{
			Action: "get_guild_member_profile",
			Params: GetGuildMemberProfileRequest{
				GuildID: msg.GuildID,
				UserID:  msg.UserID,
			},
			Echo: "get_guild_member_profile|" + msg.GuildID,
		}
	}
}

// 执行指令并反馈
func (o *QGuildLink) execCmdAndFeedback(msg GuildMsg) {
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(msg.Message, func(output *packet.CommandOutput) {
		result := ""
		// 判断是否成功
		if output.SuccessCount > 0 {
			result += msg.Message + "✓"
		} else {
			result += msg.Message + "✗"
		}
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
		o.sendChannelAtMessage(msg.GuildID, msg.ChannelID, msg.UserID, msg.Sender.Nickname, result)
	})
}

// 向所有连接子频道发送消息
func (o *QGuildLink) sendAllChannelsMessage(msg string) {
	if o.connecting {
		for guildID, channels := range o.channelIDs {
			for _, channelID := range channels {
				o.sendChan <- CQRequest{
					Action: "send_guild_channel_msg",
					Params: SendGuildChannelMsgRequest{
						GuildID:   guildID,
						ChannelID: channelID,
						Message:   msg,
					},
				}
			}
		}
	}
}

// 发送At消息
func (o *QGuildLink) sendChannelAtMessage(guildID, channelID, qq, nickname, msg string) {
	if o.connecting {
		o.sendChan <- CQRequest{
			Action: "send_guild_channel_msg",
			Params: SendGuildChannelMsgRequest{
				GuildID:   guildID,
				ChannelID: channelID,
				Message:   fmt.Sprintf("[CQ:at,qq=%s,name=%s]\n%s", qq, nickname, msg),
			},
		}
	}
}

func (o *QGuildLink) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if err := o.connect(); err != nil {
		panic("[频服互通]: 无法与 go-cqhttp 建立连接: " + err.Error())
	}
	o.Frame.GetGameListener().SetGameChatInterceptor(o.onGameMessage)
	if o.SendJoinAndLeaveMsg {
		o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDText, func(p packet.Packet) {
			pkt := p.(*packet.Text)
			if pkt.TextType == packet.TextTypeTranslation && pkt.Message == "§e%multiplayer.player.joined" {
				name := utils.ToPlainName(pkt.Parameters[0])
				if name != o.Frame.GetUQHolder().GetBotName() {
					o.sendAllChannelsMessage(fmt.Sprintf("%v 进入了游戏", name))
				}
			}
		})
		o.Frame.GetGameListener().AppendLogoutInfoCallback(func(entry protocol.PlayerListEntry) {
			player := o.Frame.GetGameControl().GetPlayerKitByUUID(entry.UUID)
			if player != nil {
				o.sendAllChannelsMessage(fmt.Sprintf("%v 离开了游戏", player.GetRelatedUQ().Username))
			}
		})
	}
}

func (o *QGuildLink) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	// 解析指令组
	o.cmdsExecutor = make(map[string]map[string][]defines.Cmd)
	for role, cmds := range o.CmdsExecutorIn {
		o.cmdsExecutor[role] = make(map[string][]defines.Cmd)
		for name, cmd := range cmds {
			var err error
			if o.cmdsExecutor[role][name], err = utils.ParseAdaptiveCmd(cmd); err != nil {
				panic(err)
			}
		}
	}
	o.sendChan = make(chan CQRequest, 64)
	o.guildsName = make(map[string]string)
	o.channelsName = make(map[string]string)
	o.channelIDs = make(map[string][]string)
	o.roleMap = make(map[string]map[string]*userRoles)
	o.awaitCmdQueue = list.New()
	o.mu = &sync.Mutex{}
}

func (o *QGuildLink) Stop() error {
	o.sendAllChannelsMessage("[频服互通]: 机器人已经退出服务器")
	return nil
}
