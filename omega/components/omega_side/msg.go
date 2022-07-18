package omega_side

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/pterm/pterm"
)

type clientMsg struct {
	ID     int                    `json:"client"`
	Action string                 `json:"function"`
	Args   map[string]interface{} `json:"args"`
}

type serverResp struct {
	ID      int         `json:"client"`
	Violate bool        `json:"violate"`
	Data    interface{} `json:"data"`
}

type ServerPush struct {
	ID0     int         `json:"client"`
	Type    string      `json:"type"`
	SubType string      `json:"sub"`
	Data    interface{} `json:"data"`
}

// - 错误数据包 (仅在插件发来的数据包不符合协议的时候由omega框架发送，
// 	收到这个数据包代表程序设计存在问题，因此，不收到这个数据包并不代表执行成功)
// 	{"client":c,"violate":true,"data":{"err":reason}}
type RespViolatePkt struct {
	Err string `json:"err"`
}

type SimplifiedPlayerInfo struct {
	Name      string `json:"name"`
	RuntimeID uint64 `json:"runtimeID"`
	UUID      string `json:"uuid"`
	UniqueID  int64  `json:"uniqueID"`
}

func wrapWriteFn(msgID int, writeFn func(interface{}) error) func(interface{}) {
	return func(resp interface{}) {
		writeFn(serverResp{ID: msgID, Violate: false, Data: resp})
	}
}

func (t *omegaSideTransporter) initMapping() {
	t.funcMapping = map[string]func(args map[string]interface{}, writer func(interface{})){
		"echo": func(args map[string]interface{}, writer func(interface{})) {
			writer(args)
		},
		"regMCPkt": func(args map[string]interface{}, writer func(interface{})) {
			pktID := args["pktID"].(string)
			if pktID == "all" {
				t.regPkt(0)
				writer(map[string]interface{}{"succ": true, "err": nil})
			} else if pktIDCode, hasK := pktIDMapping[pktID]; hasK {
				t.regPkt(pktIDCode)
				writer(map[string]interface{}{"succ": true, "err": nil})
			} else {
				writer(map[string]interface{}{"succ": false, "err": fmt.Sprintf("pktID %v not found, all possible ids are %v", pktID, pktIDNames)})
			}
		},
		"send_ws_cmd": func(args map[string]interface{}, writer func(interface{})) {
			cmd := args["cmd"].(string)
			t.side.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
				writer(map[string]interface{}{"result": output})
			})
		},
		"send_player_cmd": func(args map[string]interface{}, writer func(interface{})) {
			cmd := args["cmd"].(string)
			pterm.Warning.Println("DEBUG " + cmd)
			t.side.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(cmd, func(output *packet.CommandOutput) {
				writer(map[string]interface{}{"result": output})
			})
		},
		"send_wo_cmd": func(args map[string]interface{}, writer func(interface{})) {
			cmd := args["cmd"].(string)
			t.side.Frame.GetGameControl().SendWOCmd(cmd)
			writer(map[string]interface{}{"ack": true})
		},
		"get_uqholder": func(args map[string]interface{}, writer func(interface{})) {
			writer(t.side.Frame.GetUQHolder())
		},
		"get_players_list": func(args map[string]interface{}, writer func(interface{})) {
			playerList := []SimplifiedPlayerInfo{}
			for uniqueID, p := range t.side.Frame.GetUQHolder().PlayersByEntityID {
				name := p.Username
				runtimeID := uint64(0)
				if p.Entity != nil {
					runtimeID = p.Entity.RuntimeID
				}
				UUID := p.UUID.String()
				playerList = append(playerList, SimplifiedPlayerInfo{name, runtimeID, UUID, uniqueID})
			}
			writer(playerList)
		},
	}
}

func (p *pushController) pushMCPkt(pktID int, data interface{}) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	name := pktIDInvMapping[pktID]
	if waitors, hasK := p.typedPacketWaitor[pktID]; hasK {
		for _, w := range waitors {
			w.WriteJSON(ServerPush{ID0: 0, Type: "mcPkt", SubType: name, Data: data})
		}
	}
	for _, w := range p.anyPacketWaitor {
		w.WriteJSON(ServerPush{ID0: 0, Type: "mcPkt", SubType: name, Data: data})
	}
}
