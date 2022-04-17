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
	Delay      int      `json:"登录时延迟发送" yaml:"登录时延迟发送"`
	LoginCmds  []string `json:"登录时发送指令" yaml:"登录时发送指令"`
	LogoutCmds []string `json:"登出时发送指令" yaml:"登出时发送指令"`
	logger     defines.LineDst
	newCome    bool
}

func (b *Bonjour) Init(cfg *defines.ComponentConfig) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, b); err != nil {
		panic(err)
	}
}

func (b *Bonjour) Inject(frame defines.MainFrame) {
	b.BasicComponent.Inject(frame)
	b.Listener.AppendLoginInfoCallback(b.onLogin)
	b.Listener.AppendLogoutInfoCallback(b.onLogout)
	b.logger = &utils.MultipleLogger{
		Loggers: []defines.LineDst{
			b.Frame.GetBackendDisplay(),
			b.Frame.GetLogger("login_out.log"),
		},
	}
}

func (b *Bonjour) Activate() {
	b.BasicComponent.Activate()
	existingPlayers := make([]string, 0)
	for _, p := range b.Frame.GetUQHolder().PlayersByEntityID {
		b.Ctrl.GetPlayerKit(p.Username).GetViolatedStorage()["login_time"] = time.Now()
		existingPlayers = append(existingPlayers, p.Username)
	}
	b.logger.Write(fmt.Sprintf("当前已经在线玩家: %v", existingPlayers))
	go func() {
		time.Sleep(20)
		b.newCome = true
	}()

}

func (b *Bonjour) onLogin(entry protocol.PlayerListEntry) {
	if !b.newCome {
		return
	}
	//fmt.Println(entry)
	b.logger.Write(fmt.Sprintf("登入  %v %v", entry.Username, entry.UUID.String()))
	name := utils.ToPlainName(entry.Username)
	b.Ctrl.GetPlayerKit(entry.Username).GetViolatedStorage()["login_time"] = time.Now()
	go func() {
		t := time.NewTimer(time.Duration(b.Delay) * time.Second)
		<-t.C
		for _, cmd := range b.LoginCmds {
			s := strings.ReplaceAll(cmd, "[target_player]", name)
			b.Ctrl.SendCmd(s)
		}
	}()
}

func (b *Bonjour) onLogout(entry protocol.PlayerListEntry) {
	//fmt.Println(entry)
	player := b.Ctrl.GetPlayerKitByUUID(entry.UUID)
	if player == nil {
		b.logger.Write(fmt.Sprintf("登出 (name not found) %v %v", entry, entry.UUID.String()))
		return
	}
	playTime := time.Now().Sub(player.GetViolatedStorage()["login_time"].(time.Time)).Minutes()
	b.logger.Write(fmt.Sprintf("logout %v %v (%.1fm)", player.GetRelatedUQ().Username, entry.UUID.String(), playTime))
	name := utils.ToPlainName(player.GetRelatedUQ().Username)

	for _, cmd := range b.LogoutCmds {
		s := strings.ReplaceAll(cmd, "[target_player]", name)
		b.Ctrl.SendCmd(s)
	}
}
