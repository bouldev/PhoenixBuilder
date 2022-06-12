package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

type nameEntry struct {
	CurrentName string `json:"管理员当前名"`
	UUID        string `json:"管理员UUID"`
	AuthName    string `json:"管理员原名"`
}

type OpCheck struct {
	*BasicComponent
	OPS            []string `json:"管理员昵称"`
	fileChange     bool
	FileName       string        `json:"管理员改名记录文件"`
	fakeOPResponse []defines.Cmd `json:"假管理反制"`
	Records        map[string]*nameEntry
}

func (o *OpCheck) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.fakeOPResponse, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"假管理反制"})
	if err != nil {
		panic(err)
	}
}

//func (o *OpCheck) update(name, uuid string) {
//	newTime := utils.TimeToString(time.Now())
//	updateString := fmt.Sprintf("%v;%v", name, newTime)
//	if player, hasK := o.Records[uuid]; hasK {
//		if player.CurrentName != name {
//			player.CurrentName = name
//			player.LastUpdateTime = newTime
//			o.mainFrame.GetBackendDisplay().Write(
//				fmt.Sprintf("玩家%v改名了，曾用名为:%v", name, player.NameRecord),
//			)
//			player.NameRecord = append(player.NameRecord, updateString)
//		}
//	} else {
//		o.Records[uuid] = &nameEntry{
//			CurrentName:    name,
//			LastUpdateTime: newTime,
//			NameRecord: []string{
//				updateString,
//			},
//		}
//	}
//}

func (o *OpCheck) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.Records)
		}
	}
	return nil
}

func (o *OpCheck) Stop() error {
	fmt.Println("正在保存 " + o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.Records)
}

func (o *OpCheck) react(pk *packet.AdventureSettings) {
	playerUniqueID := pk.PlayerUniqueID
	playerName := ""
	playerUUID := ""
	for _, player := range o.Frame.GetUQHolder().PlayersByEntityID {
		//fmt.Println(player.Username, player.PlayerUniqueID)
		//fmt.Println(player.Username, player)
		//fmt.Println(player.Username, player.UUID.ID())
		if player.EntityUniqueID == playerUniqueID {
			playerName = player.Username
			playerUUID = player.UUID.String()
		}
	}
	if playerName == "" {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintln("发现一个OP玩家，但是无法获取其真名", pk))
		return
	}
	AuthName := playerName
	for _, op := range o.Records {
		if op.UUID == playerUUID {
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("OP 玩家登录 %v(%v): %v", playerName, AuthName, playerUUID))
			AuthName = op.AuthName
			op.CurrentName = playerName
			return
		}
	}
	for _, op := range o.OPS {
		if op == playerName {
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("OP 玩家登录 %v, 记录玩家ID以防改名: %v", playerName, playerUUID))
			o.Records[playerUUID] = &nameEntry{
				CurrentName: playerName,
				UUID:        playerUUID,
				AuthName:    op,
			}
			o.fileChange = true
			return
		}
	}
	o.Frame.GetBackendDisplay().Write(fmt.Sprintf("!发现 假OP玩家登录 %v(%v)", playerName, playerUUID))
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.fakeOPResponse, map[string]interface{}{
		"[player]": "\"" + playerName + "\"",
		"[uuid]":   playerUUID,
	}, o.Frame.GetBackendDisplay())
}

func (o *OpCheck) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Records = map[string]*nameEntry{}
	err := frame.GetJsonData(o.FileName, &o.Records)
	if err != nil {
		panic(err)
	}

	frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAdventureSettings, func(p packet.Packet) {
		pk := p.(*packet.AdventureSettings)
		if pk.PermissionLevel == packet.PermissionLevelOperator {
			//pks, _ := json.Marshal(pk)
			//fmt.Println(string(pks))
			//pks, _ = json.Marshal(o.Frame.GetUQHolder())
			//fmt.Println(string(pks))
			if pk.PlayerUniqueID == o.Frame.GetUQHolder().BotUniqueID {
				fmt.Println("Skip Bot Check")
				return
			}
			o.react(pk)
		}
	})
}

func (b *OpCheck) Activate() {
}
