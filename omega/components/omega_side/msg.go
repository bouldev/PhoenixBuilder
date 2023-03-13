package omega_side

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/items"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/global"
	"phoenixbuilder/omega/utils"
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

//   - 错误数据包 (仅在插件发来的数据包不符合协议的时候由omega框架发送，
//     收到这个数据包代表程序设计存在问题，因此，不收到这个数据包并不代表执行成功)
//     {"client":c,"violate":true,"data":{"err":reason}}
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
			} else if pktIDCode, hasK := utils.PktIDMapping[pktID]; hasK {
				t.regPkt(pktIDCode)
				writer(map[string]interface{}{"succ": true, "err": nil})
			} else {
				writer(map[string]interface{}{"succ": false, "err": fmt.Sprintf("pktID %v not found, all possible ids are %v", pktID, utils.PktIDNames)})
			}
		},
		"reg_mc_packet": func(args map[string]interface{}, writer func(interface{})) {
			pktID := args["pktID"].(string)
			if pktID == "all" {
				t.regPkt(0)
				writer(map[string]interface{}{"succ": true, "err": nil})
			} else if pktIDCode, hasK := utils.PktIDMapping[pktID]; hasK {
				t.regPkt(pktIDCode)
				writer(map[string]interface{}{"succ": true, "err": nil})
			} else {
				writer(map[string]interface{}{"succ": false, "err": fmt.Sprintf("pktID %v not found, all possible ids are %v", pktID, utils.PktIDNames)})
			}
		},
		"query_packet_name": func(args map[string]interface{}, writer func(interface{})) {
			pktID := int(args["pktID"].(float64))
			pktName := utils.PktIDInvMapping[pktID]
			writer(map[string]interface{}{"name": pktName})
		},
		"send_packet": func(args map[string]interface{}, writer func(interface{})) {
			_pk, ok := packet.NewPool()[uint32(args["packetID"].(float64))]
			if ok {
				pk := _pk()
				_err := json.Unmarshal([]byte(args["jsonStr"].(string)), &pk)
				if _err != nil {
					writer(map[string]interface{}{"succ": false, "err": string(_err.Error())})
				} else {
					t.side.Frame.GetGameControl().SendMCPacket(pk)
					writer(map[string]interface{}{"succ": true, "err": nil})
				}
			} else {
				writer(map[string]interface{}{"succ": false, "err": "packetID is not in pool"})
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
			// pterm.Warning.Println("DEBUG " + cmd)
			t.side.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(cmd, func(output *packet.CommandOutput) {
				writer(map[string]interface{}{"result": output})
			})
		},
		"send_wo_cmd": func(args map[string]interface{}, writer func(interface{})) {
			cmd := args["cmd"].(string)
			t.side.Frame.GetGameControl().SendWOCmd(cmd)
			writer(map[string]interface{}{"ack": true})
		},
		"send_fb_cmd": func(args map[string]interface{}, writer func(interface{})) {
			cmd := args["cmd"].(string)
			t.side.Frame.FBEval(cmd)
			writer(map[string]interface{}{"ack": true})
		},
		"get_uqholder": func(args map[string]interface{}, writer func(interface{})) {
			writer(t.side.Frame.GetUQHolder())
		},
		"get_new_uqholder": func(args map[string]interface{}, writer func(interface{})) {
			writer(t.side.Frame.GetNewUQHolder())
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
		"reg_menu": func(args map[string]interface{}, writer func(interface{})) {
			itriggers := args["triggers"].([]interface{})
			triggers := make([]string, len(itriggers))
			for i, t := range itriggers {
				triggers[i] = t.(string)
			}
			argumentHint := args["argument_hint"].(string)
			usage := args["usage"].(string)
			subType := args["sub_id"].(string)

			t.side.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
				MenuEntry: defines.MenuEntry{
					Triggers:     triggers,
					ArgumentHint: argumentHint,
					FinalTrigger: false,
					Usage:        usage,
				},
				OptionalOnTriggerFn: func(chat *defines.GameChat) (stop bool) {
					t.writeToConn(ServerPush{
						ID0:     0,
						Type:    "menuTriggered",
						SubType: subType,
						Data:    chat,
					})
					return true
				},
			})
			writer(map[string]interface{}{
				"sub_id": subType,
			})
		},
		"player.next_input": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			hint := args["hint"].(string)
			if err := t.side.Frame.GetGameControl().SetOnParamMsg(player, func(chat *defines.GameChat) (catch bool) {
				writer(map[string]interface{}{
					"success": true,
					"player":  player,
					"input":   chat.Msg,
					"err":     nil,
				})
				return true
			}); err == nil {
				if hint != "" {
					t.side.Frame.GetGameControl().SayTo(player, hint)
				}
			} else {
				writer(map[string]interface{}{
					"success": false,
					"player":  player,
					"input":   "",
					"err":     err.Error(),
				})
			}
		},
		"player.say_to": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			msg := args["msg"].(string)
			t.side.Frame.GetGameControl().SayTo(player, msg)
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"player.title_to": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			msg := args["msg"].(string)
			t.side.Frame.GetGameControl().TitleTo(player, msg)
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"player.subtitle_to": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			msg := args["msg"].(string)
			t.side.Frame.GetGameControl().SubTitleTo(player, msg)
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"player.actionbar_to": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			msg := args["msg"].(string)
			t.side.Frame.GetGameControl().ActionBarTo(player, msg)
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"player.pos": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			limit := args["limit"].(string)
			go func() {
				pos := <-t.side.Frame.GetGameControl().GetPlayerKit(player).GetPos(limit)
				if pos != nil {
					writer(map[string]interface{}{
						"success": true,
						"pos":     []int{pos.X(), pos.Y(), pos.Z()},
					})
				} else {
					writer(map[string]interface{}{
						"success": false,
						"pos":     nil,
					})
				}
			}()
		},
		"player.set_data": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			entry := args["entry"].(string)
			data := args["data"]
			if player_data, hasK := t.side.PlayerData[player]; !hasK {
				player_data = map[string]interface{}{}
				player_data[entry] = data
				t.side.PlayerData[player] = player_data
			} else {
				player_data[entry] = data
			}
			t.side.fileChange = true
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"player.get_data": func(args map[string]interface{}, writer func(interface{})) {
			player := args["player"].(string)
			entry := args["entry"].(string)
			if player_data, hasK := t.side.PlayerData[player]; !hasK {
				writer(map[string]interface{}{
					"found": false,
					"data":  nil,
				})
			} else {
				data, hasK := player_data[entry]
				writer(map[string]interface{}{
					"found": hasK,
					"data":  data,
				})
			}
		},
		"reg_login": func(args map[string]interface{}, writer func(interface{})) {
			t.side.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
				t.writeToConn(ServerPush{
					ID0:     0,
					Type:    "playerLogin",
					SubType: "",
					Data:    SimplifiedPlayerInfo{entry.Username, 0, entry.UUID.String(), entry.EntityUniqueID},
				})
			})
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"reg_logout": func(args map[string]interface{}, writer func(interface{})) {
			t.side.Frame.GetGameListener().AppendLogoutInfoCallback(func(entry protocol.PlayerListEntry) {
				player := t.side.Frame.GetGameControl().GetPlayerKitByUUID(entry.UUID)
				if player != nil {
					t.writeToConn(ServerPush{
						ID0:     0,
						Type:    "playerLogout",
						SubType: "",
						Data:    SimplifiedPlayerInfo{player.GetRelatedUQ().Username, 0, player.GetRelatedUQ().UUID.String(), player.GetRelatedUQ().EntityUniqueID},
					})
				} else {
					t.writeToConn(ServerPush{
						ID0:     0,
						Type:    "playerLogout",
						SubType: "",
						Data:    SimplifiedPlayerInfo{"unknown", 0, entry.UUID.String(), 0},
					})
				}
			})
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"reg_block_update": func(args map[string]interface{}, writer func(interface{})) {
			t.side.Frame.GetGameListener().AppendOnBlockUpdateInfoCallBack(func(pos define.CubePos, origRTID, currentRTID uint32) {
				originBlock, found := chunk.RuntimeIDToLegacyBlock(origRTID)
				if !found {
					originBlock = nil
				}
				currentBlock, found := chunk.RuntimeIDToLegacyBlock(currentRTID)
				if !found {
					currentBlock = nil
				}
				oname, oprop, ofound := chunk.RuntimeIDToState(origRTID)
				cname, cprop, cfound := chunk.RuntimeIDToState(currentRTID)
				t.writeToConn(ServerPush{
					ID0:     0,
					Type:    "blockUpdate",
					SubType: "",

					Data: map[string]interface{}{
						"pos":                        pos,
						"origin_block_runtime_id":    origRTID,
						"origin_block_simple_define": originBlock,
						"origin_block_full_define": map[string]interface{}{
							"name": oname, "props": oprop, "found": ofound,
						},
						"new_block_runtime_id":    currentRTID,
						"new_block_simple_define": currentBlock,
						"new_block_full_define": map[string]interface{}{
							"name": cname, "props": cprop, "found": cfound,
						},
					},
				})
				writer(map[string]interface{}{
					"ack": true,
				})
			})
		},
		"query_item_mapping": func(args map[string]interface{}, writer func(interface{})) {
			writer(map[string]interface{}{
				"mapping": items.RuntimeIDToItemNameMapping,
			})
		},
		"query_block_mapping": func(args map[string]interface{}, writer func(interface{})) {
			writer(map[string]interface{}{
				"blocks":        chunk.Blocks,
				"simple_blocks": chunk.LegacyBlocks,
				"java_blocks":   chunk.JavaStrToRuntimeIDMapping,
			})
		},
		"query_memory_scoreboard": func(args map[string]interface{}, writer func(interface{})) {
			global.UpdateScore(t.side.Frame.GetGameControl(), 0, func(m map[string]map[string]int) {
				writer(m)
			})
		},
		"send_qq_msg": func(args map[string]interface{}, writer func(interface{})) {
			msg := args["msg"].(string)
			if send_func, hasK := t.side.Frame.GetContext(collaborate.INTERFACE_SEND_TO_GROUP); hasK {
				send_func.(collaborate.FUNC_SEND_TO_GROUP)(msg)
			}
			writer(map[string]interface{}{
				"ack": true,
			})
		},
		"data.get_root_dir": func(args map[string]interface{}, writer func(interface{})) {
			writer(map[string]interface{}{
				"side":  t.side.getWorkingDir(),
				"omega": t.side.Frame.GetStorageRoot(),
			})
		},
		"data.list_dir": func(args map[string]interface{}, writer func(interface{})) {
			// dir := args["dir"].(string)
			// mode := "side"
			// if _mode, hasK := args["mode"]; hasK {
			// 	mode = _mode.(string)
			// }
			// if mode == "omega" {
			// 	dir = path.Join(t.side.Frame.GetStorageRoot(), dir)
			// } else {
			// 	dir = path.Join(t.side.getWorkingDir(), dir)
			// } else {
			// 	writer(map[string]interface{}{
			// 		"side":  t.side.getWorkingDir(),
			// 		"omega": t.side.Frame.GetStorageRoot(),
			// 	})
			// }
			// ioutil.ReadDir(dir)
		},
	}
}
