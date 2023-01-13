package components

import (
	"encoding/json"
	"fmt"
	"net"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
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
}

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
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
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
		// This protocol was abandoned
		conn, err := net.Dial("tcp", o.ServerAddr)
		if err != nil {
			pterm.Error.Println("无法连接至服服互通", err)
		} else {
			pterm.Info.Println("已连接至服服互通服务端", o.ServerAddr)
			decoder := json.NewDecoder(conn)
			// encoder := json.NewEncoder(conn)
			printErr := func(err error) {
				pterm.Error.Println("连接 服服互通 服务器出现错误 ", err, " 连接终止")
			}
			go func() {
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
						return
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
						}

					}
				}
			}()
			sendJson := func(data interface{}) error {
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
				return nil
			}
			if err := sendJson(map[string]interface{}{
				"KeyCode": "SuperRentalServerLink made by SuperScript",
				"server":  o.ServerName,
				// 不泄露服务器号
				"serverName": o.ServerName,
			}); err != nil {
				printErr(err)
				return
			}
			// if err := sendJson(map[string]interface{}{
			// 	"data_type": "check", "data": "server_list",
			// }); err != nil {
			// 	return
			// }
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
		}
	} else if o.Mode == "SuperScript@DotCS-V2" {
		cliVersion := 4
		retrying := 0
		retryTime := 30
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
				if retrying > 0 {
					pterm.Info.Println(fmt.Sprintf("服服互通 第%v次重连中心服务器， 下次重连间隔将会变为%v秒", retrying, retryTime))
				}
				isConnected := false
				conn, err := net.Dial("tcp", o.ServerAddr)
				if err != nil {
					pterm.Error.Println("无法连接至服服互通服务器: ", err)
					return
				} else {
					pterm.Info.Println("已连接至服服互通服务器", o.ServerAddr)
					isConnected = true
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
						if isConnected {
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
						"robotType":  "Omega",
					}); err != nil {
						isConnected = false
						if o.AutoReconnect {
							pterm.Warning.Println(fmt.Sprintf("登录 服服互通服务端 时发生错误， 但是自动重连已打开， 将在 %vs 后尝试重连", retryTime))
							time.Sleep(time.Duration(retryTime) * time.Second)
							break
						} else {
							return
						}
					}
					// connectOK
					retrying = 0
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
							isConnected = false
							if o.AutoReconnect {
								pterm.Warning.Println(fmt.Sprintf("与 服服互通服务端 断开连接， 但是自动重连已打开， 将在 %vs 后尝试重连", retryTime))
								time.Sleep(time.Duration(retryTime) * time.Second)
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
				retrying++
				if retryTime < 3600 {
					retryTime *= 2
				}
			}
		}()
	} else {
		panic(fmt.Errorf("意外的连接模式: %v 未支持，你需要使用已有连接模式", o.Mode))
	}
}
