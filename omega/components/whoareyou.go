package components

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"

	"gopkg.in/square/go-jose.v2/json"
)

type WhoAreYou struct {
	*defines.BasicComponent
	react         []defines.Cmd
	Tag           string `json:"添加标签"`
	checkTag      string
	checkRngMark1 string
	checkRngMark2 string
	onScan        bool
	Schedule      int `json:"定时重新扫描"`
	Delay         int `json:"登入检查延迟"`
}

func (o *WhoAreYou) Init(cfg *defines.ComponentConfig) {

	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.react, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"反制"})
	if err != nil {
		panic(err)
	}
	o.checkTag = "_omega_system_name_check_1"
	o.checkRngMark1 = "_omega_system_name_check_r1"
	o.checkRngMark2 = "_omega_system_name_check_r2"
}

func (o *WhoAreYou) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().AppendLoginInfoCallback(o.onLogin)
}

func (o *WhoAreYou) onLogin(entry protocol.PlayerListEntry) {
	//fmt.Println("新登入")
	go func() {
		<-time.NewTimer(time.Second * time.Duration(o.Delay)).C
		//fmt.Println("登入时扫描")
		o.scan()
	}()

}

func (o *WhoAreYou) handleCheckResult(name string) {
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.react, map[string]interface{}{
		"[player]": "\"" + name + "\"",
		"[tag]":    o.Tag,
	}, o.Frame.GetBackendDisplay())
}

func (o *WhoAreYou) Activate() {
	//fmt.Println("激活时扫描")
	go func() {
		<-time.NewTimer(time.Second * time.Duration(o.Delay) * 2).C
		o.scan()
		t := time.NewTicker(time.Duration(o.Schedule) * time.Second)
		for {
			<-t.C
			//fmt.Println("定时扫描")
			o.scan()
		}
	}()

}

func (o *WhoAreYou) scan() {
	if o.onScan {
		//fmt.Println("跳过扫描")
		return
	}
	o.onScan = true
	cmd := o.Frame.GetGameControl().SendCmd
	cmd("tag @a add " + o.checkTag)
	cmd("tag @a add " + o.checkRngMark1)
	cmd("tag @a add " + o.checkRngMark2)
	allName := []string{}
	for _, player := range o.Frame.GetUQHolder().PlayersByEntityID {
		allName = append(allName, player.Username)
	}
	cmd("tag @s remove " + o.checkTag)
	go func() {
		for _, name := range allName {
			cmd(fmt.Sprintf("tag \"%v\" remove "+o.checkTag, name))
			cmd(fmt.Sprintf("tag @a[name=\"%v\"] remove "+o.checkTag, name))
		}
		<-time.NewTimer(time.Second / 5).C
		cmd("tag @a remove " + o.checkRngMark2)
		for _, name := range allName {
			cmd(fmt.Sprintf("tag \"%v\" remove "+o.checkTag, name))
			cmd(fmt.Sprintf("tag @a[name=\"%v\"] remove "+o.checkTag, name))
		}
		cmd("tag @s remove " + o.checkTag)
		<-time.NewTimer(time.Second / 5).C
		for _, name := range allName {
			cmd(fmt.Sprintf("tag \"%v\" remove "+o.checkTag, name))
			cmd(fmt.Sprintf("tag @a[name=\"%v\"] remove "+o.checkTag, name))
		}
		cmd("tag @s remove " + o.checkTag)
		illegal_names := []string{}
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("testfor @a[tag="+o.checkTag+",tag="+o.checkRngMark1+",tag=!"+o.checkRngMark2+"]", func(output *packet.CommandOutput) {
			if output.SuccessCount < 1 {
				o.onScan = false
				cmd("tag @a remove " + o.checkRngMark1)
				cmd("tag @a remove " + o.checkRngMark2)
				cmd("tag @a remove " + o.checkTag)
				return
			}
			for _, msg := range output.OutputMessages {
				if len(msg.Parameters) == 1 {
					illegal_names = append(illegal_names, msg.Parameters[0])
				}
			}
			for _, name := range allName {
				cmd(fmt.Sprintf("tag \"%v\" remove "+o.checkTag, name))
				cmd(fmt.Sprintf("tag @a[name=\"%v\"] remove "+o.checkTag, name))
			}
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("发现违规昵称: %v,添加tag: @a[tag=%v]", illegal_names, o.Tag))
			cmd(fmt.Sprintf("tag @a[tag="+o.checkTag+",tag="+o.checkRngMark1+",tag=!"+o.checkRngMark2+"] add %v", o.Tag))
			for _, name := range illegal_names {
				o.handleCheckResult(name)
			}
			cmd("tag @a remove " + o.checkRngMark1)
			cmd("tag @a remove " + o.checkRngMark2)
			cmd("tag @a remove " + o.checkTag)
			o.onScan = false
		})
	}()
}
