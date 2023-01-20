package sunlife

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

type Equip struct {
	*defines.BasicComponent
	Data                map[string]*Data
	PassWord            string                  `json:"激活密码"`
	Menu                map[string]string       `json:"菜单"`
	Score               map[string]string       `json:"所需计分板"`
	TitleWord           map[string]string       `json:"提示词"`
	Equips              map[string]*Equips      `json:"装备"`
	Triggers            []string                `json:"触发词"`
	Usage               string                  `json:"描述"`
	Types               map[string]string       `json:"类型对应"`
	SynchronizeTheScore map[string]*Synchronize `json:"合并计分板"`
}
type Synchronize struct {
	OneScore  string `json:"需要相加的计分板1"`
	SencScore string `json:"需要相加的计分板2"`
}
type Equips struct {
	//装备类型
	TypeOfEquips string `json:"类型"`
	//属性
	Attribute map[string]string `json:"属性"`
	//价格
	Price int `json:"价格"`
}

// 玩家数据
type Data struct {
	//头部
	Heads map[string]string
	//饰品
	Jewelrys map[string]string
	//褪
	Legs map[string]string
	//脚部
	Foots map[string]string
	//胸甲
	Cuirass    map[string]string
	TheCuirass string
	Head       string
	Jewelry    string
	Leg        string
	Foot       string
}

func (b *Equip) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	if b.Data == nil {
		b.Data = make(map[string]*Data)
	}
	if cfg.Version == "0.0.1" {
		delete(cfg.Configs, "激活密码")
		cfg.Description = "免费的组件"
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	//b.GuildData = make(map[string]*GuildDatas)
}
func (b *Equip) InitData() {
	go func() {
		for k, _ := range b.Score {
			b.CmdSender(fmt.Sprintf("scoreboard objectives add %v dummy", k))
		}
		list := <-GetPlayerName(b.Frame, "@a")
		for _, v := range list {
			if _, ok := b.Data[v]; !ok {
				b.Data[v] = &Data{
					Heads:    make(map[string]string),
					Jewelrys: make(map[string]string),
					Legs:     make(map[string]string),
					Foots:    make(map[string]string),
					Cuirass:  make(map[string]string),
				}
			}
		}
	}()

}
func (b *Equip) Inject(frame defines.MainFrame) {
	b.Frame = frame
	//注入frame等东西
	/*
		b.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddItemActor, func(p packet.Packet) {
			fmt.Print("凋落物的包:", p, "\n")
		})
	*/

	b.BasicComponent.Inject(frame)
	//b.Frame.GetJsonData("新公会信息.json", &b.GuildData)
	b.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     b.Triggers,
			ArgumentHint: " ",
			FinalTrigger: false,
			Usage:        b.Usage,
		},
		OptionalOnTriggerFn: b.Center,
	})
	b.Frame.GetJsonData("装备系统.json", &b.Data)
	b.InitData()
	b.Listener.AppendLoginInfoCallback(b.onLogin)

}
func (b *Equip) Activate() {
	go func() {
		for {
			time.Sleep(time.Second * 3)
			go func() {
				playerList := <-GetPlayerName(b.Frame, "@a")
				for _, v := range playerList {
					if _, ok := b.Data[v]; ok {
						theLevelList := b.GetBackScore(v)
						for scoreName, scoreNum := range theLevelList {
							b.CmdSender(fmt.Sprintf("scoreboard players set @a[name=\"%v\"] %v %v", v, scoreName, scoreNum))
						}
					}
				}
			}()
			//同步分数
			go func() {
				ScoreList := <-GetScore(b.Frame)
				//循环所有人的分数()
				for k, v := range ScoreList {
					for EndScore, data := range b.SynchronizeTheScore {
						if _, ok := v[EndScore]; ok {
							if fscore, isok := v[data.OneScore]; isok {
								if sscore, isok := v[data.SencScore]; isok {
									cmd := fmt.Sprintf("scoreboard players set @a[name=\"%v\"] %v %v", k, EndScore, (fscore + sscore))
									b.CmdSender(cmd)
								}

							}
						}
					}
				}

			}()
		}
	}()
}

// 检查类型
func (b *Equip) CheckType(equip string) (string, bool) {
	for k, v := range b.Types {
		if v == equip {
			return k, true
		}
	}
	return "", false
}

// 购买物品
func (b *Equip) Buy(name string, equip string) {
	cmd := fmt.Sprintf("scoreboard players remove @a[name=\"%v\",scores={%v=%v..}] %v %v", name, b.Score["购买所需计分板"], b.Equips[equip].Price, b.Score["购买所需计分板"], b.Equips[equip].Price)
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {

			data := b.Equips[equip]
			if TheName, isok := b.CheckType(data.TypeOfEquips); isok {
				switch TheName {
				case "头部":
					b.Data[name].Heads[equip] = "这是一个头部"
				case "脚":
					b.Data[name].Legs[equip] = "这是一个脚部"
				case "胸甲":
					b.Data[name].Cuirass[equip] = "这是一个胸甲"
				case "饰品":
					b.Data[name].Jewelrys[equip] = "这是一个饰品"
				case "腿":
					b.Data[name].Foots[equip] = "这是一个腿部"
				}
				Sayto(b.Frame, name, "购买成功")
			} else {
				Sayto(b.Frame, name, "抱歉管理员配置类型出错购买失败")
			}
		} else {
			pterm.Info.Printfln("错误指令 错误信息为:%v\n错误指令为:%v", output.OutputMessages, cmd)
		}
	})
}

// 商店菜单
func (b *Equip) Store(name string) {
	num := 0
	theMsg := ""
	list := make(map[string]string)
	for k, v := range b.Equips {
		msg := ""
		for i, j := range v.Attribute {
			thelist := map[string]interface{}{
				"计分板名": i,
				"提升":   j,
			}
			msg += FormateMsg(b.Frame, thelist, b.Menu["属性模板"]) + "/"
		}
		relist := map[string]interface{}{
			"i":  num,
			"装备": k,
			"类型": v.TypeOfEquips,
			"价格": v.Price,
			"属性": msg,
		}
		theMsg += FormateMsg(b.Frame, relist, b.Menu["商店模板"])
		list[strconv.Itoa(num)] = k
		num++
	}
	Sayto(b.Frame, name, theMsg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if equipName, ok := list[chat.Msg[0]]; ok {
				b.Buy(name, equipName)
			} else {
				Sayto(b.Frame, name, "输入有效数字")
			}
		}

		return true
	})

}

// 保存数据
func (b *Equip) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return b.Frame.WriteJsonDataWithTMP("装备系统.json", ".ckpt", &b.Data)
	}
	return nil
}

// 命令发送者
func (b *Equip) CmdSender(str string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(str, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {

		} else {
			pterm.Info.Printfln("指令错误 错误信息为:%v\n错误指令为:%v", output.OutputMessages, str)
		}
	})

}

// 仓库
func (b *Equip) WareHouse(name string) {
	msg := ""
	num := 0
	list := make(map[string]string)
	for k, _ := range b.Data[name].Jewelrys {
		relist := map[string]interface{}{
			"i":  num,
			"装备": k,
			"类型": b.Equips[k].TypeOfEquips,
		}
		list[strconv.Itoa(num)] = k
		msg += FormateMsg(b.Frame, relist, b.Menu["仓库模板"])
		num++
	}
	for k, _ := range b.Data[name].Foots {
		relist := map[string]interface{}{
			"i":  num,
			"装备": k,
			"类型": b.Equips[k].TypeOfEquips,
		}
		list[strconv.Itoa(num)] = k
		msg += FormateMsg(b.Frame, relist, b.Menu["仓库模板"])
		num++
	}
	for k, _ := range b.Data[name].Heads {
		relist := map[string]interface{}{
			"i":  num,
			"装备": k,
			"类型": b.Equips[k].TypeOfEquips,
		}
		list[strconv.Itoa(num)] = k
		msg += FormateMsg(b.Frame, relist, b.Menu["仓库模板"])
		num++
	}
	for k, _ := range b.Data[name].Cuirass {
		relist := map[string]interface{}{
			"i":  num,
			"装备": k,
			"类型": b.Equips[k].TypeOfEquips,
		}
		list[strconv.Itoa(num)] = k
		msg += FormateMsg(b.Frame, relist, b.Menu["仓库模板"])
		num++
	}
	for k, _ := range b.Data[name].Legs {
		relist := map[string]interface{}{
			"i":  num,
			"装备": k,
			"类型": b.Equips[k].TypeOfEquips,
		}
		list[strconv.Itoa(num)] = k
		msg += FormateMsg(b.Frame, relist, b.Menu["仓库模板"])
		num++
	}
	//fmt.Println("list", list)
	Sayto(b.Frame, name, msg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if equipName, ok := list[chat.Msg[0]]; ok {
				b.Apparel(name, equipName)
			} else {
				Sayto(b.Frame, name, "输入有效数字")
			}
		}

		return true
	})

}

// 穿戴
func (b *Equip) Apparel(name string, equip string) {
	fmt.Println("test", equip)
	typeName, ok := b.CheckType(b.Equips[equip].TypeOfEquips)
	if ok {
		switch typeName {
		case "头部":
			b.Data[name].Head = equip
		case "脚":
			b.Data[name].Leg = equip
		case "胸甲":
			b.Data[name].TheCuirass = equip
		case "饰品":

			b.Data[name].Jewelry = equip
			fmt.Println("jewelry", b.Data[name].Jewelry)
		case "腿":
			b.Data[name].Foot = equip
		}
		Sayto(b.Frame, name, "装备成功")
	} else {
		Sayto(b.Frame, name, "类型错误 请管理员修改类型")
	}

	/*
		for k, v := range b.Equips[equip].Attribute {
			cmd := fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", name, k, v)
			b.CmdSender(cmd)
		}
	*/

}
func (b *Equip) GetBackScore(name string) map[string]int {
	list := make(map[string]int)
	if b.Data[name].Jewelry != "" {
		for attributeName, num := range b.Equips[b.Data[name].Jewelry].Attribute {
			intnum, _ := strconv.Atoi(num)
			if _, ok := list[attributeName]; ok {
				list[attributeName] += intnum
			} else {
				list[attributeName] = intnum
			}
		}

	}
	if b.Data[name].Foot != "" {
		for attributeName, num := range b.Equips[b.Data[name].Foot].Attribute {
			intnum, _ := strconv.Atoi(num)
			if _, ok := list[attributeName]; ok {
				list[attributeName] += intnum
			} else {
				list[attributeName] = intnum
			}
		}
	}
	if b.Data[name].Head != "" {
		for attributeName, num := range b.Equips[b.Data[name].Head].Attribute {
			intnum, _ := strconv.Atoi(num)
			if _, ok := list[attributeName]; ok {
				list[attributeName] += intnum
			} else {
				list[attributeName] = intnum
			}
		}
	}
	if b.Data[name].TheCuirass != "" {
		for attributeName, num := range b.Equips[b.Data[name].TheCuirass].Attribute {
			intnum, _ := strconv.Atoi(num)
			if _, ok := list[attributeName]; ok {
				list[attributeName] += intnum
			} else {
				list[attributeName] = intnum
			}
		}
	}
	if b.Data[name].Leg != "" {
		for attributeName, num := range b.Equips[b.Data[name].Leg].Attribute {
			intnum, _ := strconv.Atoi(num)
			if _, ok := list[attributeName]; ok {
				list[attributeName] += intnum
			} else {
				list[attributeName] = intnum
			}
		}
	}
	return list
}
func (b *Equip) GetBackData(name string) {
	msg := ""
	for k, v := range b.GetBackScore(name) {
		list := map[string]interface{}{
			"计分板名": k,
			"提升":   v,
		}
		msg += FormateMsg(b.Frame, list, b.Menu["属性模板"])
	}
	//fmt.Println("test", b.Data[name].Jewelry)
	fmt.Println(b.Data[name])
	headName := b.Data[name].Head
	if headName == "" {
		headName = "未装备"
	}
	CuirassName := b.Data[name].TheCuirass
	if CuirassName == "" {
		CuirassName = "未装备"
	}
	legName := b.Data[name].Leg
	if legName == "" {
		legName = "未装备"
	}
	jewelryName := b.Data[name].Jewelry
	if jewelryName == "" {
		jewelryName = "未装备"
	}
	FootName := b.Data[name].Foot
	if FootName == "" {
		FootName = "未装备"
	}
	relist := map[string]interface{}{
		"player": name,
		"头部":     headName,
		"胸甲":     CuirassName,
		"腿部":     legName,
		"武器":     jewelryName,
		"脚部":     FootName,
		"属性":     msg,
	}

	Sayto(b.Frame, name, b.Menu["个人装备栏菜单"])
	Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["个人装备栏"]))
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.Data[name].Head = ""
			case "1":
				b.Data[name].TheCuirass = ""
			case "2":
				b.Data[name].Leg = ""
			case "3":
				b.Data[name].Jewelry = ""
			case "4":
				b.Data[name].Foot = ""

			}
			Sayto(b.Frame, name, "脱下成功")
		}

		return true
	})

}
func (b *Equip) Center(chat *defines.GameChat) bool {
	Sayto(b.Frame, chat.Name, b.Menu["主菜单"])
	b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.Store(chat.Name)
			case "1":
				b.WareHouse(chat.Name)
			case "2":
				b.GetBackData(chat.Name)
			}
		}

		return true
	})
	return true
}

// 初始化分数
func (b *Equip) onLogin(entry protocol.PlayerListEntry) {
	if _, ok := b.Data[entry.Username]; !ok {
		pterm.Info.Println("初始化成功", entry.Username)
		b.Data[entry.Username] = &Data{
			Legs:     make(map[string]string),
			Leg:      "",
			Jewelrys: make(map[string]string),
			Jewelry:  "",
			Foots:    make(map[string]string),
			Foot:     "",
			Heads:    make(map[string]string),
			Head:     "",
			Cuirass:  make(map[string]string),
		}
	}
}
