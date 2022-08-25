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

	"github.com/pterm/pterm"
)

type SeverToServerChatRoom struct {
	*defines.BasicComponent
	ServerName string `json:"服务器名"`
	Mode       string `json:"协议模式"`
	ServerAddr string `json:"中心服务器地址"`
}

func (o *SeverToServerChatRoom) Init(cfg *defines.ComponentConfig) {
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
			o.ServerName = "服务器-" + code
		} else {
			o.ServerName = "匿名服务器"
		}
	}
	pterm.Info.Println("连接模式: ", o.Mode)
	if o.Mode == "SuperScript@DotCS" {
		conn, err := net.Dial("tcp", o.ServerAddr)
		if err != nil {
			pterm.Error.Println("无法连接至服服互通", err)
		} else {
			pterm.Info.Println("已连接至服服互通服务器", o.ServerAddr)
			decoder := json.NewDecoder(conn)
			// encoder := json.NewEncoder(conn)
			printErr := func(err error) {
				pterm.Error.Println("到服服互通服务器连接出现错误 ", err, " 连接终止")
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
							o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v服 加入了互通", msg.ServerName))
						case "disconnected":
							o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v服 加入了互通", msg.ServerName))
						case "consolemsg":
							pterm.Info.Println(msg.Data)
						case "player.join":
							o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v 加入了游戏", msg.ServerName, msg.Data))
						case "player.left":
							o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("%v %v 退出了游戏", msg.ServerName, msg.Data))
						case "get_data_serverlist":
							o.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("目前连接的服服互通服务器列表: %v", msg.Data))
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
				"KeyCode":    "SuperRentalServerLink made by SuperScript",
				"server":     o.ServerName,
				"serverName": o.ServerName,
			}); err != nil {
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
					"data":      fmt.Sprintf("§7<§a%v§7> §7%v", chat.Name, strings.Join(chat.Msg, " ")),
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
	} else {
		panic(fmt.Errorf("连接模式: %v 未支持", o.Mode))
	}
}
