package components

import (
	"encoding/json"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

type SkinCheck struct {
	*defines.BasicComponent
	DebugDisplay bool                   `json:"玩家上线时在后台显示皮肤名"`
	RegexStr     map[string]interface{} `json:"匹配到皮肤名时则执行以下命令"`
	IgnoreBot    bool                   `json:"忽略机器人"`
	compiledReg  map[string]regexp.Regexp
	compiledCmds map[string][]defines.Cmd
}

func (o *SkinCheck) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	if cfg.Version == "0.0.1" {
		cfg.Configs["忽略机器人"] = true
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	o.compiledCmds = make(map[string][]defines.Cmd)
	o.compiledReg = make(map[string]regexp.Regexp)
	for regStr, cmds := range o.RegexStr {
		if cc, err := utils.ParseAdaptiveCmd(cmds); err != nil {
			panic(err)
		} else {
			o.compiledCmds[regStr] = cc
		}
		o.compiledReg[regStr] = *regexp.MustCompile(regStr)

	}
}

func (o *SkinCheck) onSkin(playerName, skinID string, UUID uuid.UUID, uniqueID int64) {
	isBot := o.Frame.GetUQHolder().BotUniqueID == uniqueID
	if o.DebugDisplay {
		isBotStr := "[玩家]"
		if isBot {
			isBotStr = "[机器人]"
		}
		pterm.Info.Printfln("皮肤信息: %v %v(UUID=%v) 皮肤名: %v", isBotStr, playerName, UUID.String(), skinID)
	}
	if isBot && o.IgnoreBot {
		return
	}
	for regStr, cmds := range o.compiledCmds {
		reg := o.compiledReg[regStr]
		if reg.Match([]byte(skinID)) {
			go utils.LaunchCmdsArray(o.Frame.GetGameControl(), cmds, map[string]interface{}{
				"[player]": "\"" + playerName + "\"",
				"[skinID]": skinID,
				"[UUID]":   UUID.String(),
			}, o.Frame.GetBackendDisplay())
		}
	}
}

func (o *SkinCheck) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDPlayerList, func(p packet.Packet) {
		pk := p.(*packet.PlayerList)
		for _, entry := range pk.Entries {
			skinID := entry.Skin.SkinID
			playerName := entry.Username
			playerUUID := entry.UUID
			o.onSkin(playerName, skinID, playerUUID, entry.EntityUniqueID)
		}
	})

}
