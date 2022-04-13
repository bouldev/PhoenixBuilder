package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"
)

type Bonjour struct {
	*BasicComponent
	Delay      int      `json:"login_delay" yaml:"login_delay"`
	LoginCmds  []string `json:"login_cmds" yaml:"login_cmds"`
	LogoutCmds []string `json:"logout_cmds" yaml:"logout_cmds"`
	logger     defines.LineDst
}

func (b *Bonjour) Init(cfg *defines.ComponentConfig) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, b); err != nil {
		panic(err)
	}
}

func (b *Bonjour) Inject(frame defines.MainFrame) {
	b.BasicComponent.Inject(frame)
	b.listener.AppendLoginInfoCallback(b.onLogin)
	b.listener.AppendLogoutInfoCallback(b.onLogout)
	b.logger = &utils.MultipleLogger{
		Loggers: []defines.LineDst{
			b.frame.GetBackendDisplay(),
			b.frame.GetLogger("login_out.log"),
		},
	}
}

func (b *Bonjour) Activate() {
	b.BasicComponent.Activate()
	existingPlayers := make([]string, 0)
	for _, p := range b.frame.GetUQHolder().PlayersByEntityID {
		b.ctrl.GetPlayerKit(p.Username).GetViolatedStorage()["login_time"] = time.Now()
		existingPlayers = append(existingPlayers, p.Username)
	}
	b.logger.Write(fmt.Sprintf("当前已经在线玩家: %v", existingPlayers))
}

func (b *Bonjour) onLogin(entry protocol.PlayerListEntry) {
	//fmt.Println(entry)
	b.logger.Write(fmt.Sprintf("登入  %v %v", entry.Username, entry.UUID.String()))
	name := utils.ToPlainName(entry.Username)
	b.ctrl.GetPlayerKit(entry.Username).GetViolatedStorage()["login_time"] = time.Now()
	go func() {
		t := time.NewTimer(time.Duration(b.Delay) * time.Second)
		<-t.C
		for _, cmd := range b.LoginCmds {
			s := strings.ReplaceAll(cmd, "[target_player]", name)
			b.ctrl.SendCmd(s)
		}
	}()
}

func (b *Bonjour) onLogout(entry protocol.PlayerListEntry) {
	//fmt.Println(entry)
	player := b.ctrl.GetPlayerKitByUUID(entry.UUID)
	if player == nil {
		b.logger.Write(fmt.Sprintf("登出 (name not found) %v %v", entry, entry.UUID.String()))
		return
	}
	playTime := time.Now().Sub(player.GetViolatedStorage()["login_time"].(time.Time)).Minutes()
	b.logger.Write(fmt.Sprintf("logout %v %v (%.1fm)", player.GetRelatedUQ().Username, entry.UUID.String(), playTime))
	name := utils.ToPlainName(player.GetRelatedUQ().Username)

	for _, cmd := range b.LogoutCmds {
		s := strings.ReplaceAll(cmd, "[target_player]", name)
		b.ctrl.SendCmd(s)
	}
}
