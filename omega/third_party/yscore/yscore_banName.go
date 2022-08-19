package yscore

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	"regexp"
	"time"
)

type BanName struct {
	*defines.BasicComponent
	Namelist       []string `json:"违规名字正则表达式"`
	DelayTime      int      `json:"延迟检测时间(秒)"`
	Title          string   `json:"踢出时提示话语"`
	BanList        map[string]string
	ComponentsName string
	Username       collaborate.STRING_FB_USERNAME
}

func (b *BanName) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	b.ComponentsName = cfg.Name
	b.BanList = make(map[string]string)
	//b. = make(map[string]*GuildDatas)

}
func (b *BanName) Inject(frame defines.MainFrame) {
	b.Frame = frame
	b.BasicComponent.Inject(frame)
	b.Listener.AppendLoginInfoCallback(b.onLogin)
	b.Frame.GetJsonData("自定义封禁违规名字.json", &b.BanList)
	CreateNameHash(b.Frame)
	//fmt.Println("-------", b.SnowsMenuTitle)
}
func (b *BanName) Activate() {

}
func (b *BanName) onLogin(entry protocol.PlayerListEntry) {
	if len(b.BanList) > 0 {
		if _, ok := b.BanList[entry.Username]; ok {

			b.Kick(entry.Username, b.Title)
			//b.Frame.GetGameControl().SendCmd(fmt.Sprintf("kick @a[name=\"%v\"] %v", entry.Username, b.Title))
		}

	}
	go func() {
		time.Sleep(time.Millisecond * time.Duration(b.DelayTime))
		for _, v := range b.Namelist {
			if b.checkName(entry.Username, v) {
				b.Kick(entry.Username, b.Title)
				b.BanList[entry.Username] = v
				fmt.Println("封禁玩家] " + entry.Username + " 因为违规词:" + v)
				b.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("[封禁玩家] %v 因为名字含违规词 %v", entry.Username, v))
			}
		}
	}()

}
func (b *BanName) checkName(name string, str string) bool {
	ok, _ := regexp.MatchString(str, name)
	return ok
}
func (b *BanName) Kick(name string, msg string) {
	b.Frame.GetGameControl().SendCmd(fmt.Sprintf("kick @a[name=\"%v\"] %v", name, msg))
}
