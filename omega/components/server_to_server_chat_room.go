package components

import (
	"encoding/json"
	"fmt"
	"net"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type SeverToServerChatRoom struct {
	*defines.BasicComponent
	ServerName           string                            `json:"服务器名"`
	Mode                 string                            `json:"协议模式"`
	ServerAddr           string                            `json:"中心服务器地址"`
	ProtocolSpecificData map[string]map[string]interface{} `json:"特定协议附加数据"`
	AutoReconnect        bool                              `json:"断线自动重连"`
	ExtraFmt             *SuperLinkScoreboardData
	APIDataFmt           *SuperLinkAPIData
	TimeOut              int
	sendJson             func(data interface{}) error
}

type SeverToServerChatroomFmt struct {
	UploadScoreSuccess         string
	DownloadScoreSuccess       string
	ServerNoSuchScoreboard     string
	ClientScoreboardNotAllowed string
	ClientScoreLack            string
	ServerScoreLack            string
	CondNotMatch               string
}

type SuperLink struct {
	ConnectOK     bool
	Retry         int
	RetryWaitTime int
	CmdUIDRecv    map[string]map[string]interface{}
	apiMsgFmt     map[string]interface{}
}

type SuperLinkScoreboardData struct {
	NegativeScoreSupport bool   `json:"服务端分数记录为负仍然允许上传分数"`
	UpdateNeed           string `json:"符合条件才可以上传"`
	UploadScoreSuccess   string `json:"成功存入分数提示"`
	DownloadScoreSuccess string `json:"成功提取分数提示"`
	ClientScoreboardLack string `json:"在租赁服计分板不存在或分数不足"`
	ServerScoreboardLack string `json:"服务端计分板不存在"`
	CondNotMatch         string `json:"不符合分数上传条件"`
}

type SuperLinkAPIData struct {
	APIExecuteCmds []string    `json:"收到相同API事件后执行的指令"`
	APITriggerMsg  string      `json:"发送API事件的触发词"`
	UseExtraData   interface{} `json:"使用额外数据接口"`
	SendBack       bool        `json:"测试模式(将这个消api息转发出去的同时也转发到自己服)"`
	Notice         string      `json:"说明"`
}

var super_Link SuperLink

func (o *SeverToServerChatRoom) Init(cfg *defines.ComponentConfig) {
	if cfg.Version == "0.0.1" {
		// There's no users use this version, is it?
		cfg.Configs["服务器名"] = ""
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	if cfg.Version == "0.0.2" {
		cfg.Configs["特定协议附加数据"] = map[string]map[string]interface{}{
			"SuperScript@DotCS": map[string]interface{}{},
			"SuperScript@DotCS-V2": map[string]interface{}{
				"token": "",
				"频道":    "默认大区",
			},
		}
		cfg.Version = "0.0.3"
		cfg.Upgrade()
	}
	if cfg.Version == "0.0.3" {
		if cfg.Configs["中心服务器地址"] == "124.222.13.238" {
			pterm.Warning.Println("服服互通： 原服务器地址 124.222.13.238 已被弃用， 请在之后更改")
		}
		cfg.Configs["断线自动重连"] = false
		cfg.Version = "0.0.4"
		cfg.Upgrade()
	}
	if cfg.Version == "0.0.4" {
		table := (cfg.Configs["特定协议附加数据"]).(map[string]interface{})
		table["SuperScript@DotCS-V3"] = map[string]interface{}{
			"协议说明": []string{
				"默认互通模式： 只转发和接收同一个频道的其他租赁服的信息",
				"计分板互通： 相当于跨服银行的功能，同一个频道才支持计分板互通，若设置'允许互通的计分板及计分板提示词'为空，则只支持消息互通；",
				"计分板互通： 如果输入的是正数，就是向服务端存入分数(租赁服部分会克扣相应分数哦)； 如果输入的是负数，就是向服务端提取分数",
				"计分板互通： 注意：在默认大区互通计分板有风险，建议在自己的私人频道使用计分板互通(把选项'频道'改成随便一个名字即可，两个互通的服务器都使用这个频道名)",
				"自定义API事件互通： 传输自定义的API事件，具有非常强大的扩展性功能，可以同时***使用3个通道传输消息***；支持使用目标选择器，计分板分数和普通文字",
				"自定义API事件互通： 使用命令方块/tellraw <机器人名字> {'rawtext':[{'text':'发送API事件的关键词'},数据1,数据2,数据3]} (记得把单引号改成双引号)",
				"自定义API事件互通： 数据1/2/3可以是 {'text':'文本'}，也可以是{'selector':'<目标选择器>'}，还可以是{'score':{'name':'计分板项','objective':'计分板名'}} (记得把单引号改成双引号)，具体看下面的example",
				"自定义API事件互通： 如果同一个频道中，其它租赁服也有相同的API事件名，那这个租赁服接收到API事件名后就会执行API对应的指令",
			},
			"token":      "",
			"频道":         "默认大区",
			"启用跨服聊天":     true,
			"上传计分板分数触发词": []string{"上传"},
			"计分板名不在允许范围内的提示词": "§c该计分板不允许互通哦",
			"允许互通的计分板及计分板提示词": map[string]map[string]interface{}{
				"示例_money": {
					"服务端分数记录为负仍然允许上传分数": false,
					"符合条件才可以上传":         "@a[scores={示例_money=-999999..},tag=!ban]",
					"成功存入分数提示":          "§a成功存入了[count]积分， 余额为[left]积分",
					"成功提取分数提示":          "§a成功提取了[count]积分， 余额为[left]积分",
					"在租赁服计分板不存在或分数不足":   "§c你的积分不足哦， 或者你在这个计分板上没有分数，无法上传",
					"在中心服务器存储的计分板分数不足":  "§c您的积分已不足， 余额为[left]，  无法提取",
					"服务端计分板不存在":         "服务端该计分板不存在!",
					"不符合分数上传条件":         "被ban的玩家不能上传哦",
				},
				"示例二_super": {
					"服务端分数记录为负仍然允许上传分数": false,
					"符合条件才可以上传":         "@a[tag=amazing_tag]",
					"成功存入分数提示":          "§a成功捉走了[count]只super， 你还可以再抓[left]只喔",
					"成功提取分数提示":          "§a成功丢入了[count]只super， 你还可以再抓[left]只喔",
					"在租赁服计分板不存在或分数不足":   "§c啊呀，你没有这么多只super可以上传啦",
					"在中心服务器存储的计分板分数不足":  "§c啊呀，你只在云端存了[left]只super，  不可以贪心哦",
					"服务端计分板不存在":         "哈呀..云端不支持存放super啦..或许你可以试试改成CMA2401PT？",
					"不符合分数上传条件":         "你有神奇的标签，不能向云端存入或提取super哦",
				},
			},
			"自定义API数据和事件名": map[string]map[string]interface{}{
				"示例": {
					"说明": "触发方法：在地上放置一个命令方块，输入：/tellraw <机器人名字> {'rawtext':[{'text':'example'},{'text':'hello'},{'selector':'@p'},{'score':{'name':'@p','objective':'雪球菜单'}}]}并激活 (记得把单引号全部改成双引号)",
					"收到相同API事件后执行的指令": []string{
						"/say 我接收到了一条示例API事件：",
						"/say 值1是： [APIData1]， 显示的内容应当是： hello，",
						"/say 值2是： [APIData2]， 显示的内容应当是： 最近的玩家的名字",
						"/say 值3是： [APIData3]， 显示的内容应当是： 这个玩家的雪球菜单计分板分数",
					},
					"发送API事件的关键词": "example",
					"测试模式(将这个消api息转发出去的同时也转发到自己服)": true,
				},
				"跨服大喇叭": {
					"说明": "触发方法：在地上放置一个命令方块，输入：/tellraw <机器人名字> {'rawtext':[{'text':'广播'},{'selector':'@p'},{'text':'这是广播内容'},{'score':{'name':'@p','objective':'雪球菜单'}}]}并激活 (记得把单引号全部改成双引号)",
					"收到相同API事件后执行的指令": []string{
						"/tellraw @a {\"rawtext\":[{\"text\":\"§l§a跨服广播 §d[APIData1] §f说： §e[APIData2]\"}]}",
					},
					"发送API事件的关键词": "广播",
					"测试模式(将这个消api息转发出去的同时也转发到自己服)": true,
				},
			},
		}
		cfg.Version = "0.0.5"
		cfg.Upgrade()
		panic("配置文件升级完成，你需要重启Omega以启用该配置(你还可能需要认真查看配置)")

	}
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	if cfg.Configs["协议模式"] == "SuperScript@DotCS-V3" {
		t := cfg.Configs["特定协议附加数据"].(map[string]interface{})["SuperScript@DotCS-V3"].(map[string]interface{})
		scbMsgFmt := t["允许互通的计分板及计分板提示词"].(map[string]interface{})
		for _, v := range scbMsgFmt {
			m, err := json.Marshal(v)
			if err != nil {
				panic(fmt.Sprintf("服服互通 计分板互通配置部分有误 > %v %v", err, v))
			}
			if err := json.Unmarshal(m, &o.ExtraFmt); err != nil {
				panic(fmt.Sprintf("服服互通 计分板互通配置部分有误 < %v %v", err, v))
			}
		}
		apiMsgFmt := t["自定义API数据和事件名"].(map[string]interface{})
		for _, v := range apiMsgFmt {
			m, err := json.Marshal(v)
			if err != nil {
				panic(fmt.Sprintf("服服互通 API事件配置部分有误 > %v %v", err, v))
			}
			if err := json.Unmarshal(m, &o.APIDataFmt); err != nil {
				panic(fmt.Sprintf("服服互通 API事件配置部分有误 < %v %v", err, v))
			}
		}
		super_Link.apiMsgFmt = apiMsgFmt
	}
}

func (o *SeverToServerChatRoom) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if o.ServerName == "" {
		code, err := o.Frame.QuerySensitiveInfo(defines.SENSITIVE_INFO_SERVER_CODE_HASH)
		if err == nil {
			if len(code) > 6 {
				code = code[:6]
			}
			o.ServerName = "租赁服-" + code
		} else {
			o.ServerName = "匿名租赁服"
		}
		pterm.Warning.Println("服服互通: 还没有设置在公屏聊天栏显示的服务器名, 将会显示为", o.ServerName, ", 你可以在 配置文件-服服互通 中以更改")
	}
	if o.ServerAddr == "需要自行填写" {
		panic(fmt.Errorf("服服互通： 还没有填写 中心服务器地址， 你需要自行填写或者查找可用的服务器"))
	}
	if !strings.Contains(o.ServerAddr, ":") {
		panic(fmt.Errorf("服服互通： 中心服务器地址有误， 你可能忘记填写端口了(格式可能是：x.x.x.x:24013)"))
	}
	pterm.Info.Println("连接模式: ", o.Mode)
	if o.Mode == "SuperScript@DotCS" {
		panic(fmt.Errorf("连接模式: %v 被弃用， 请在配置内更换为 SuperScript@DotCS-V2", o.Mode))
	} else if o.Mode == "SuperScript@DotCS-V2" {
		cliVersion := 4
		super_Link.Retry = 0
		super_Link.RetryWaitTime = 30
		super_Link.ConnectOK = false // 可以不加这行的..不过不知道为什么还是打算加
		additonalData := o.ProtocolSpecificData[o.Mode]
		printErr := func(err error) {
			pterm.Error.Println("连接 服服互通 中心服务器出现问题： ", err, " 连接终止")
		}
		if additonalData == nil {
			panic(fmt.Errorf("该协议需要附加数据"))
		}
		if len(o.ServerName) > 15 {
			for _p := 0; _p < 3; _p++ {
				pterm.Warning.Println("服服互通： 服务器名长度大于15， 很可能无法连接至中心服务器！")
			}
		}
		pterm.Info.Println("正在尝试登录到服服互通服务端..")
		go func() {
			for {
				if super_Link.Retry > 0 {
					pterm.Info.Println(fmt.Sprintf("服服互通 第%v次重连中心服务器， 下次重连间隔将会变为%v秒", super_Link.Retry, super_Link.RetryWaitTime))
				}
				super_Link.ConnectOK = false
				conn, err := net.Dial("tcp", o.ServerAddr)
				if err != nil {
					pterm.Error.Println("无法连接至服服互通服务器: ", err)
					return
				} else {
					pterm.Info.Println("已连接至服服互通服务器", o.ServerAddr)
					super_Link.ConnectOK = true
					decoder := json.NewDecoder(conn)
					// encoder := json.NewEncoder(conn)
					var loginData struct {
						NeedToken  bool `json:"needToken"`
						ServerVers int  `json:"version"`
					}
					err = decoder.Decode(&loginData)
					if err != nil {
						printErr(err)
						return
					}
					if cliVersion != loginData.ServerVers {
						pterm.Warning.Println("服服互通： 客户端版本与服务端版本不匹配， 很可能连接出错")
					}
					// pterm.Info.Println(loginData)
					var token interface{}
					token = nil
					if loginData.NeedToken {
						if additonalData == nil || additonalData["token"] == nil {
							panic(fmt.Errorf("服服互通： 该中心服务器需要 token"))
						} else {
							token = additonalData["token"]
						}
					}
					sendJson := func(data interface{}) error {
						if super_Link.ConnectOK {
							if buf, err := json.Marshal(data); err != nil {
								pterm.Error.Println(err)
								return err
							} else {
								// pterm.Info.Println(buf)
								if _, err := conn.Write(buf); err != nil {
									printErr(err)
									return err
								}
							}
						}
						return nil
					}
					if err := sendJson(map[string]interface{}{
						"KeyCode":    "RentalServerLink made by 2401 & SuperScript",
						"server":     o.ServerName,
						"serverName": o.ServerName,
						"token":      token,
						"channel":    additonalData["频道"],
						"robotType":  "Original_Omega",
					}); err != nil {
						super_Link.ConnectOK = false
						if o.AutoReconnect {
							pterm.Warning.Println(fmt.Sprintf("登录 服服互通服务端 时发生错误， 但是自动重连已打开， 将在 %vs 后尝试重连", super_Link.RetryWaitTime))
							time.Sleep(time.Duration(super_Link.RetryWaitTime) * time.Second)
							break
						} else {
							return
						}
					}
					// connectOK
					super_Link.Retry = 0
					// Listener
					o.Frame.GetGameListener().SetGameChatInterceptor(func(chat *defines.GameChat) (stop bool) {
						if chat.Type != packet.TextTypeChat {
							return false
						}
						if err := sendJson(map[string]interface{}{
							"data_type": "msg",
							"data":      fmt.Sprintf("§7<§f%v§7> §7%v", chat.Name, strings.Join(chat.Msg, " ")),
						}); err != nil {
							return
						}
						return false
					})
					o.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
						sendJson(map[string]interface{}{
							"data_type": "player.join",
							"data":      utils.ToPlainName(entry.Username),
							"isAdmin":   false,
						})
					})
					o.Frame.GetGameListener().AppendLogoutInfoCallback(func(entry protocol.PlayerListEntry) {
						player := o.Ctrl.GetPlayerKitByUUID(entry.UUID)
						if player == nil || player.GetRelatedUQ() == nil || player.GetRelatedUQ().Username == "" {
							return
						}
						sendJson(map[string]interface{}{
							"data_type": "player.left",
							"data":      player.GetRelatedUQ().Username,
							"isAdmin":   false,
						})
					})
					// Sender
					for {
						var msg struct {
							Action     string      `json:"data_type"`
							ServerName string      `json:"serverName"`
							Data       interface{} `json:"data"`
							ChatColor  string      `json:"msgColor"`
							ChatData   string      `json:"msgInfo"`
						}
						if err := decoder.Decode(&msg); err != nil {
							printErr(err)
							super_Link.ConnectOK = false
							if o.AutoReconnect {
								pterm.Warning.Println(fmt.Sprintf("与 服服互通服务端 断开连接， 但是自动重连已打开， 将在 %vs 后尝试重连", super_Link.RetryWaitTime))
								time.Sleep(time.Duration(super_Link.RetryWaitTime) * time.Second)
								break
							} else {
								return
							}
						} else {
							switch msg.Action {
							case "msg":
								o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v", msg.ServerName, msg.Data))
							case "connected":
								o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v 加入了互通", msg.ServerName))
							case "disconnected":
								o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v 加入了互通", msg.ServerName))
							case "consolemsg":
								pterm.Info.Println(msg.Data)
							case "player.join":
								o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v 加入了游戏", msg.ServerName, msg.Data))
							case "player.left":
								o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v 退出了游戏", msg.ServerName, msg.Data))
							case "get_data_serverlist":
								o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("目前连接的服服互通服务器列表: %v", msg.Data))
							case "kick":
								return
							}
						}
					}
				}
				super_Link.Retry++
				if super_Link.RetryWaitTime < 3600 {
					super_Link.RetryWaitTime *= 2
				}
			}
		}()
		//
	} else if o.Mode == "SuperScript@DotCS-V3" {
		cliVersion := 5
		super_Link.Retry = 0
		super_Link.RetryWaitTime = 30
		super_Link.ConnectOK = false
		additonalData := o.ProtocolSpecificData[o.Mode]
		super_Link.CmdUIDRecv = map[string]map[string]interface{}{}
		ScbUpdateTriggers := []string{}
		rawScbUpdateTriggers := additonalData["上传计分板分数触发词"].([]interface{})
		enableServerChatroom := additonalData["启用跨服聊天"].(bool)
		for _, _v := range rawScbUpdateTriggers {
			val, ok := _v.(string)
			if ok {
				ScbUpdateTriggers = append(ScbUpdateTriggers, val)
			} else {
				panic(fmt.Sprintf("解析关键词出现问题"))
			}
		}
		if !true {
			panic(fmt.Sprintf("服服互通： 无法识别 \"上传计分板分数触发词\"：%v - %v", additonalData["上传计分板分数触发词"], ScbUpdateTriggers))
		}
		printErr := func(err error) {
			pterm.Error.Println("连接 服服互通 中心服务器出现问题： ", err, " 连接终止")
		}
		if additonalData == nil {
			panic(fmt.Errorf("该协议需要附加数据"))
		}
		if len(o.ServerName) > 15 {
			for _p := 0; _p < 3; _p++ {
				pterm.Warning.Println("服服互通： 服务器名长度大于15， 很可能无法连接至中心服务器！")
			}
		}
		o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
			MenuEntry: defines.MenuEntry{
				Triggers:     ScbUpdateTriggers,
				ArgumentHint: "[计分板名] [分数]",
				Usage:        "跨服上传和提取计分板分数",
				FinalTrigger: false,
			},
			OptionalOnTriggerFn: o.uploadScore,
		})
		pterm.Info.Println("正在尝试登录到服服互通服务端..")
		go func() {
			for {
				if super_Link.Retry > 0 {
					pterm.Info.Println(fmt.Sprintf("服服互通 第%v次重连中心服务器， 下次重连间隔将会变为%v秒", super_Link.Retry, super_Link.RetryWaitTime))
				}
				super_Link.ConnectOK = false
				conn, err := net.Dial("tcp", o.ServerAddr)
				if err != nil {
					pterm.Error.Println("无法连接至服服互通服务器: ", err)
					return
				} else {
					pterm.Info.Println("已连接至服服互通服务器", o.ServerAddr)
					super_Link.ConnectOK = true
					super_Link.Retry = 0
					super_Link.RetryWaitTime = 30
					decoder := json.NewDecoder(conn)
					// encoder := json.NewEncoder(conn)
					var loginData struct {
						NeedToken  bool `json:"needToken"`
						ServerVers int  `json:"version"`
					}
					err = decoder.Decode(&loginData)
					if err != nil {
						printErr(err)
						return
					}
					if cliVersion != loginData.ServerVers {
						pterm.Warning.Println("服服互通： 客户端版本与服务端版本不匹配， 很可能连接出错")
					}
					// pterm.Info.Println(loginData)
					var token interface{}
					token = nil
					if loginData.NeedToken {
						if additonalData == nil || additonalData["token"] == nil {
							panic(fmt.Errorf("服服互通： 该中心服务器需要 token"))
						} else {
							token = additonalData["token"]
						}
					}
					o.sendJson = func(data interface{}) error {
						if super_Link.ConnectOK {
							if buf, err := json.Marshal(data); err != nil {
								pterm.Error.Println(err)
								return err
							} else {
								// pterm.Info.Println(buf)
								if _, err := conn.Write(buf); err != nil {
									printErr(err)
									return err
								}
							}
						}
						return nil
					}
					if err := o.sendJson(map[string]interface{}{
						"KeyCode":    "RentalServerLink made by 2401 & SuperScript",
						"server":     o.ServerName,
						"serverName": o.ServerName,
						"token":      token,
						"channel":    additonalData["频道"],
						"robotType":  "Original_Omega",
					}); err != nil {
						super_Link.ConnectOK = false
						if o.AutoReconnect {
							pterm.Warning.Println(fmt.Sprintf("登录 服服互通服务端 时发生错误， 但是自动重连已打开， 将在 %vs 后尝试重连", super_Link.RetryWaitTime))
							time.Sleep(time.Duration(super_Link.RetryWaitTime) * time.Second)
							break
						} else {
							return
						}
					}
					// connectOK
					super_Link.Retry = 0
					// Listener
					o.Frame.GetGameListener().SetGameChatInterceptor(func(chat *defines.GameChat) (stop bool) {
						if chat.Type == 9 {
							go o.handleAPIEventSend(chat.Msg)
						}
						if chat.Type != packet.TextTypeChat {
							return false
						}
						if err := o.sendJson(map[string]interface{}{
							"data_type": "msg",
							"data":      fmt.Sprintf("§7<§f%v§7> §7%v", chat.Name, strings.Join(chat.Msg, " ")),
						}); err != nil {
							return
						}
						return false
					})
					o.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
						o.sendJson(map[string]interface{}{
							"data_type": "player.join",
							"data":      utils.ToPlainName(entry.Username),
							"isAdmin":   false,
						})
					})
					o.Frame.GetGameListener().AppendLogoutInfoCallback(func(entry protocol.PlayerListEntry) {
						player := o.Ctrl.GetPlayerKitByUUID(entry.UUID)
						if player == nil || player.GetRelatedUQ() == nil || player.GetRelatedUQ().Username == "" {
							return
						}
						o.sendJson(map[string]interface{}{
							"data_type": "player.left",
							"data":      player.GetRelatedUQ().Username,
							"isAdmin":   false,
						})
					})
					// Sender
					for {
						var msg struct {
							Action     string      `json:"data_type"`
							ServerName string      `json:"serverName"`
							Data       interface{} `json:"data"`
							ChatColor  string      `json:"msgColor"`
							ChatData   string      `json:"msgInfo"`
							APIData    interface{} `json:"APIData"`
							UID        string      `json:"UID"`
						}
						if err := decoder.Decode(&msg); err != nil {
							printErr(err)
							super_Link.ConnectOK = false
							if o.AutoReconnect {
								pterm.Warning.Println(fmt.Sprintf("与 服服互通服务端 断开连接， 但是自动重连已打开， 将在 %vs 后尝试重连", super_Link.RetryWaitTime))
								time.Sleep(time.Duration(super_Link.RetryWaitTime) * time.Second)
								break
							} else {
								return
							}
						} else {
							if enableServerChatroom {
								switch msg.Action {
								case "msg":
									o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v", msg.ServerName, msg.Data))
								case "connected":
									o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v 加入了互通", msg.ServerName))
								case "disconnected":
									o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v 加入了互通", msg.ServerName))
								case "player.join":
									o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v 加入了游戏", msg.ServerName, msg.Data))
								case "player.left":
									o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v 退出了游戏", msg.ServerName, msg.Data))
								case "get_data_serverlist":
									o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("目前连接的服服互通服务器列表: %v", msg.Data))
								}
							}
							switch msg.Action {
							case "consolemsg":
								pterm.Info.Println(msg.Data)
							case "kick":
								return
							case "upload_scb.result":
								table, ok := (msg.Data).(map[string]interface{})
								if ok {
									super_Link.CmdUIDRecv[msg.UID] = table
								} else {
									pterm.Error.Println("Data type is not expected: ", msg.Data)
								}
							case "api.data":
								go o.handleAPIEventRecv(msg.Data, msg.APIData)
							}

						}
					}
				}
				super_Link.Retry++
				if super_Link.RetryWaitTime < 3600 {
					super_Link.RetryWaitTime *= 2
				}
			}
		}()
	} else {
		panic(fmt.Errorf("意外的连接模式: %v 未支持，你需要使用已有连接模式", o.Mode))
	}
}

func (o *SeverToServerChatRoom) uploadScore(chat *defines.GameChat) (stop bool) {
	if super_Link.ConnectOK {
		if len(chat.Msg) == 2 {
			updateScore, err := strconv.Atoi(chat.Msg[1])
			updateScore *= -1
			if err != nil {
				o.Frame.GetGameControl().SayTo(chat.Name, "§c呀..格式错了， 请重新输入哦")
				return true
			}
			scbName := chat.Msg[0]
			v3data := o.ProtocolSpecificData[o.Mode]
			if v3data["允许互通的计分板及计分板提示词"].(map[string]interface{})[scbName] == nil {
				o.Frame.GetGameControl().SayTo(chat.Name, o.getScoreboardMsgFmt(scbName).ClientScoreboardNotAllowed)
				return true
			}
			selector := v3data["允许互通的计分板及计分板提示词"].(map[string]interface{})[scbName].(map[string]interface{})["符合条件才可以上传"]
			o.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(
				fmt.Sprintf("/testfor %v", selector), func(output *packet.CommandOutput) {
					if output.SuccessCount == 0 {
						o.Frame.GetGameControl().SayTo(chat.Name, o.getScoreboardMsgFmt(scbName).CondNotMatch)
					} else {
						if updateScore < 0 {
							// 存入
							cmd := fmt.Sprintf("/scoreboard players test \"%v\" %v %v", chat.Name, scbName, -1*updateScore)
							o.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(
								cmd, func(output *packet.CommandOutput) {
									if output.SuccessCount == 0 {
										o.Frame.GetGameControl().SayTo(chat.Name, o.getScoreboardMsgFmt(scbName).ClientScoreLack)
									} else {
										go o.handleScoreboardUpload(scbName, chat.Name, updateScore, true)
									}
								})

						} else if updateScore > 0 {
							// 提取
							cmd := fmt.Sprintf("/scoreboard players set __test__ %v 0", scbName)
							o.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(
								cmd, func(output *packet.CommandOutput) {
									if output.SuccessCount == 0 {
										o.Frame.GetGameControl().SayTo(chat.Name, o.getScoreboardMsgFmt(scbName).ClientScoreLack)
									} else {
										go o.handleScoreboardUpload(scbName, chat.Name, updateScore, true)
									}
								})

						}
					}
				})
			return true
		} else {
			o.Frame.GetGameControl().SayTo(chat.Name, "§c呀..格式错了， 请重新输入哦")
		}
	} else {
		o.Frame.GetGameControl().SayTo(chat.Name, "§c未连接到服服互通， 无法上传分数哦")
	}
	return true
}

func (o *SeverToServerChatRoom) getScoreboardMsgFmt(scoreboardname string) SeverToServerChatroomFmt {
	// not a good func
	generalFmt := (o.ProtocolSpecificData[o.Mode]["允许互通的计分板及计分板提示词"]).(map[string]interface{})[scoreboardname].(map[string]interface{})
	return SeverToServerChatroomFmt{
		UploadScoreSuccess:         generalFmt["成功存入分数提示"].(string),
		DownloadScoreSuccess:       generalFmt["成功提取分数提示"].(string),
		ClientScoreLack:            generalFmt["在租赁服计分板不存在或分数不足"].(string),
		ServerNoSuchScoreboard:     generalFmt["服务端计分板不存在"].(string),
		CondNotMatch:               generalFmt["不符合分数上传条件"].(string),
		ClientScoreboardNotAllowed: o.ProtocolSpecificData[o.Mode]["计分板名不在允许范围内的提示词"].(string),
	}
}

func (o *SeverToServerChatRoom) handleScoreboardUpload(scoreboardname string, player string, score int, noNegativeScore bool) {
	// P_SCORE: store; N_SCORE: take
	o.TimeOut = 3
	msgUID := fmt.Sprintf("b%X", time.Now().UnixMilli())
	if err := o.sendJson(map[string]interface{}{
		"data_type":  "upload.scb",
		"data":       score,
		"ExtraData1": scoreboardname,
		"ExtraData2": player,
		"ExtraData3": noNegativeScore,
		"UID":        msgUID,
	}); err != nil {
		return
	}
	timer := time.Now().Unix()
	for int(time.Now().Unix()-timer) < o.TimeOut {
		time.Sleep(time.Duration(1) * time.Millisecond * 100)
		// wait for resp
		for _k, dat := range super_Link.CmdUIDRecv {
			if _k == msgUID {
				if dat["success"] == true {
					o.Frame.GetGameControl().SendCmd(
						fmt.Sprintf("/scoreboard players add %v %v %v", scoreboardname, player, score),
					)
					if score > 0 {
						o.Frame.GetGameControl().SayTo(player,
							utils.FormatByReplacingOccurrences(
								o.getScoreboardMsgFmt(scoreboardname).UploadScoreSuccess,
								map[string]interface{}{
									"[left]": dat["left"],
									"[draw]": score,
								}),
						)
					} else {
						o.Frame.GetGameControl().SayTo(player,
							utils.FormatByReplacingOccurrences(
								o.getScoreboardMsgFmt(scoreboardname).DownloadScoreSuccess,
								map[string]interface{}{
									"[left]": dat["left"],
									"[draw]": score,
								}),
						)
					}
				} else if dat["status"] == "failed.negative" {
					o.Frame.GetGameControl().SayTo(player, utils.FormatByReplacingOccurrences(
						o.getScoreboardMsgFmt(scoreboardname).ServerScoreLack,
						map[string]interface{}{
							"[left]": dat["left"],
							"[draw]": -score,
						},
					))
				} else if dat["status"] == "failed.no_such_scoreboard" {
					o.Frame.GetGameControl().SayTo(player, o.getScoreboardMsgFmt(scoreboardname).ServerNoSuchScoreboard)
				} else {
					o.Frame.GetGameControl().SayTo(player, fmt.Sprintf("无法上传， 出现问题： %v", dat["status"]))
				}
				delete(super_Link.CmdUIDRecv, _k)
				return
			}
		}
	}
	o.Frame.GetGameControl().SayTo(player, "§c计分板上传超时..")
}

func (o *SeverToServerChatRoom) handleAPIEventSend(msg []string) {
	apiMsg := o.getTextFromJSON(msg[0])
	if len(apiMsg) != 4 {
		return
	}
	if !super_Link.ConnectOK {
		fmt.Println("接收到了可能有效的API消息", apiMsg, "， 但是未能连上服服互通，将忽略")
	}
	for k, v := range super_Link.apiMsgFmt {
		if v.(map[string]interface{})["发送API事件的关键词"] == apiMsg[0] {
			sendBack := v.(map[string]interface{})["测试模式(将这个消api息转发出去的同时也转发到自己服)"].(bool)
			o.sendJson(map[string]interface{}{
				"data_type": "api.data",
				"data":      k,
				"APIData": map[string]interface{}{
					"ExtraData1": apiMsg[1],
					"ExtraData2": apiMsg[2],
					"ExtraData3": apiMsg[3],
					"SendBack":   sendBack,
				},
			})
		}
	}
}

func (o *SeverToServerChatRoom) handleAPIEventRecv(apiName interface{}, apiData interface{}) {
	apiDatas, ok := apiData.(map[string]interface{})
	if !ok {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("接收到无法解析的API消息，已跳过： %v %v", apiDatas, apiData))
		return
	}
	for k, v := range super_Link.apiMsgFmt {
		if k == apiName {
			cmds := v.(map[string]interface{})["收到相同API事件后执行的指令"].([]interface{})
			for ind := range cmds {
				cmd := cmds[ind].(string)
				o.Frame.GetGameControl().SendWOCmd(utils.FormatByReplacingOccurrences(
					cmd,
					map[string]interface{}{
						"[APIData1]": apiDatas["ExtraData1"],
						"[APIData2]": apiDatas["ExtraData2"],
						"[APIData3]": apiDatas["ExtraData3"],
					},
				))
			}
		}
	}
}

func (o *SeverToServerChatRoom) getTextFromJSON(rawJson string) []string {
	output := []string{}
	defer func() {
		if err := recover(); err != nil {
			// fmt.Println("§c解析APIData失败，原因：", err)
		}
	}()
	k := make(map[string]interface{}, 0)
	json.Unmarshal([]byte(rawJson), &k)
	for ind := range k["rawtext"].([]interface{}) {
		output = append(output, k["rawtext"].([]interface{})[ind].(map[string]interface{})["text"].(string))
	}
	return output
}
