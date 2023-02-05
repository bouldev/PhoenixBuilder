package sunlife

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

type King struct {
	*defines.BasicComponent
	//玩家信息
	PlayerData map[string]*KingData
	Tirgger    []string          `json:"触发词"`
	Usage      string            `json:"描述"`
	Words      map[string]string `json:"提示词"`
	Menu       map[string]string `json:"菜单"`
	Content    KingContent       `json:"国家相关信息"`
}
type KingData struct {
	Master  string
	Member  map[string]string
	Level   int
	TerrPos map[string][]int
	TpPos   map[string][]int
	//保护持续有效时间
	ColdTime int
	//申请列表
	Application map[string]string
}
type KingContent struct {
	//领地等级可拥有领地数
	NumOfTerr       map[string]int `json:"领地等级可拥有领地数"`
	NumOfpopulation map[string]int `json:"领地等级可拥有人数"`
	//领地范围
	Extent []int `json:"领地范围"`
	//纳税系统
	RatepaySysteam RatePay           `json:"纳税系统"`
	BuySysteam     map[string]string `json:"购买系统(必须执行成功)"`
	//ProtectPaySysteam ProtectPay          `json:"保护费系统"`
	ProtectSysteam map[string][]string `json:"保护系统"`
}
type RatePay struct {
	ColdTime int    `json:"领地纳税限制时间(秒)"`
	Cmds     string `json:"纳税时执行指令"`
}
type ProtectPay struct {
	ColdTime int      `json:"保护费限制时间(秒)"`
	Cmds     []string `json:"缴纳保护费执行指令"`
}

func (b *King) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	b.PlayerData = make(map[string]*KingData)

}
func (b *King) Inject(frame defines.MainFrame) {
	b.Frame = frame
	//注入frame等东西
	/*
		b.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddItemActor, func(p packet.Packet) {
			fmt.Print("凋落物的包:", p, "\n")
		})
	*/

	b.BasicComponent.Inject(frame)
	b.Frame.GetJsonData("王国信息.json", &b.PlayerData)
	b.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     b.Tirgger,
			ArgumentHint: " ",
			FinalTrigger: false,
			Usage:        b.Usage,
		},
		OptionalOnTriggerFn: b.Center,
	})

}
func (b *King) Center(chat *defines.GameChat) bool {
	Sayto(b.Frame, chat.Name, b.Menu["主菜单"])
	//b.Frame.GetGameControl().PlaceCommandBlock()
	if b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(NewChat *defines.GameChat) (catch bool) {
		if len(NewChat.Msg) > 0 {
			switch NewChat.Msg[0] {
			case "0":
				//购买国家
				b.BuyKing(chat.Name)
			case "1":
				//在有国家情况下购买领地
				b.BuyTerr(chat.Name)
			case "2":
				//返回详细信息
				b.GetKingData(chat.Name)
			case "3":
				//返回设定好的传送点
				b.TpBack(chat.Name)
			case "4":
				b.Pay(chat.Name)
			case "5":
				b.JoinKing(chat.Name)
			case "6":
				//管理菜单
				b.Manager(chat.Name)
				//b.SetTpPos(chat.Name)
			case "7":
				b.QuitKing(chat.Name)

			}

		}

		return true
	}) == nil {

	}
	return false
}
func (b *King) QuitKing(name string) {
	kingdom, isin := b.FindkingdomByName(name)
	_, isMaster := b.FindkingdomByMasterName(name)
	if !isin {
		Sayto(b.Frame, name, b.Words["没有王国提示"])
		return
	}
	if isMaster {
		Sayto(b.Frame, name, b.Words["国王不能退出"])
		return
	}
	delete(b.PlayerData[kingdom].Member, name)
	Sayto(b.Frame, name, b.Words["退出成功"])

}

// 管理菜单
func (b *King) Manager(name string) {
	_, isok := b.FindkingdomByMasterName(name)
	if !isok {
		Sayto(b.Frame, name, b.Words["不能进入管理菜单提示"])
		return
	}
	Sayto(b.Frame, name, b.Menu["管理菜单"])
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) <= 0 {
			return
		}
		switch chat.Msg[0] {
		case "0":
			b.SetTpPos(name)
		case "1":
			b.LookApplication(name)
		case "2":
			b.KickMember(name)
		}

		return true
	})

}

// 踢出成员
func (b *King) KickMember(name string) {
	msg := ""
	list := make(map[string]string)
	num := 0
	kingdom, _ := b.FindkingdomByMasterName(name)
	for k, _ := range b.PlayerData[kingdom].Member {
		relist := map[string]interface{}{
			"i":      num,
			"player": name,
		}
		list[strconv.Itoa(num)] = k
		Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["踢出成员菜单格式"]))
		msg += FormateMsg(b.Frame, relist, b.Menu["踢出成员菜单格式"])
		num++
	}
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) <= 0 {
			return true
		}
		if playerName, isin := list[chat.Msg[0]]; isin {
			delete(b.PlayerData[kingdom].Member, playerName)
			Sayto(b.Frame, name, b.Words["踢出玩家成功"])
		}
		return true
	})
}

// 查看申请列表
func (b *King) LookApplication(name string) {
	msg := ""
	list := make(map[string]string)
	num := 0
	kingdom, _ := b.FindkingdomByMasterName(name)
	for k, _ := range b.PlayerData[kingdom].Application {
		relist := map[string]interface{}{
			"i":      num,
			"player": name,
		}
		list[strconv.Itoa(num)] = k
		Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["查看申请列表格式"]))
		msg += FormateMsg(b.Frame, relist, b.Menu["查看申请列表格式"])
		num++
	}
	fmt.Println(msg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) <= 0 {
			return true
		}

		if playerName, isin := list[chat.Msg[0]]; isin {
			_, wasOk := b.FindkingdomByName(playerName)
			if wasOk {
				Sayto(b.Frame, name, b.Words["申请被提前同意"])
				delete(b.PlayerData[kingdom].Application, playerName)
				return true
			}
			b.PlayerData[kingdom].Member[playerName] = "这是一个成员"
			Sayto(b.Frame, name, b.Words["同意申请成功"])
		}
		return true
	})

}

// 购买国家
func (b *King) BuyKing(name string) {
	//检查是否在国家内
	_, isok := b.FindkingdomByMasterName(name)
	if isok {
		Sayto(b.Frame, name, b.Words["已有王国提示"])
		return
	}
	go func() {
		relist := map[string]interface{}{
			"player": name,
		}
		cmd := FormateMsg(b.Frame, relist, b.Content.BuySysteam["购买王国执行指令"])
		b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
			if output.SuccessCount <= 0 {
				Sayto(b.Frame, name, b.Words["购买王国失败提示"])
				return
			}
			Sayto(b.Frame, name, b.Words["提示输入国家名字"])
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				//是否为有效输入
				if len(chat.Msg) <= 0 {
					return true
				}
				//检查是否国家名字重复
				if _, isRe := b.PlayerData[chat.Msg[0]]; isRe {
					Sayto(b.Frame, name, b.Words["国家名字重复"])
					return true
				}
				b.PlayerData[chat.Msg[0]] = &KingData{
					Master:      name,
					Member:      make(map[string]string),
					Level:       1,
					TerrPos:     make(map[string][]int),
					TpPos:       make(map[string][]int),
					ColdTime:    b.Content.RatepaySysteam.ColdTime,
					Application: make(map[string]string),
				}
				//从老国家中移除此人
				OldKing, isIn := b.FindkingdomByName(name)
				if isIn {
					delete(b.PlayerData[OldKing].Member, name)
					fmt.Println(fmt.Println("%v中移除玩家%v", OldKing, name))
				}
				pterm.Println("新的国家建立", b.PlayerData[chat.Msg[0]])
				Sayto(b.Frame, name, b.Words["国家建立成功"])

				return true
			})

		})
	}()

}

// 购买领地
func (b *King) BuyTerr(name string) {
	Kingdom, isFind := b.FindkingdomByMasterName(name)

	if !isFind {
		Sayto(b.Frame, name, b.Words["没有领地提示"])
		return
	}
	//判断领地数量是否超载
	switch b.PlayerData[Kingdom].Level {
	case 1:
		if len(b.PlayerData[Kingdom].TerrPos) >= b.Content.NumOfTerr["一级"] {
			Sayto(b.Frame, name, b.Words["领地数过量"])
			return
		}
	case 2:
		if len(b.PlayerData[Kingdom].TerrPos) >= b.Content.NumOfTerr["二级"] {
			Sayto(b.Frame, name, b.Words["领地数过量"])
			return
		}
	case 3:
		if len(b.PlayerData[Kingdom].TerrPos) >= b.Content.NumOfTerr["三级"] {
			Sayto(b.Frame, name, b.Words["领地数过量"])
			return
		}
	case 4:
		if len(b.PlayerData[Kingdom].TerrPos) >= b.Content.NumOfTerr["四级"] {
			Sayto(b.Frame, name, b.Words["领地数过量"])
			return
		}
	}

	relist := map[string]interface{}{
		"player": name,
	}
	cmd := FormateMsg(b.Frame, relist, b.Content.BuySysteam["购买领地执行指令"])
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		//购买失败
		if output.SuccessCount <= 0 {
			Sayto(b.Frame, name, b.Words["购买领地失败提示"])
			return
		}
		Sayto(b.Frame, name, b.Words["提示输入领地名字"])
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) <= 0 {
				return true
			}
			//检查是否名字重复
			kingdom, _ := b.FindkingdomByMasterName(name)
			if _, isRe := b.PlayerData[kingdom].TerrPos[chat.Msg[0]]; isRe {
				Sayto(b.Frame, name, b.Words["领地名字重复"])
				return true
			}
			//名字没重复 则尝试创建
			go func() {
				playerPos := <-GetPos(b.Frame, fmt.Sprintf("@a[name=\"%v\"]", name))
				if b.CheckInTerr(playerPos[name]) {
					Sayto(b.Frame, name, b.Words["附近有领地"])
					return
				}
				b.PlayerData[kingdom].TerrPos[chat.Msg[0]] = playerPos[name]
				Sayto(b.Frame, name, b.Words["建立领地成功"])
			}()

			return true
		})

	})

}

// 返回详细信息
func (b *King) GetKingData(name string) {
	kingdom, isok := b.FindkingdomByName(name)
	if isok {
		data := b.PlayerData[kingdom]
		relist := map[string]interface{}{
			"king":   data.Master,
			"成员":     data.Member,
			"等级":     data.Level,
			"领地":     data.TerrPos,
			"领地有效时间": data.ColdTime,
		}
		Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Words["国家详细信息"]))
		//提前退出
		return
	}
	Sayto(b.Frame, name, b.Words["没有王国提示"])

}

// 返回已经设定好的传送点
func (b *King) TpBack(name string) {
	kingdom, isok := b.FindkingdomByName(name)
	if !isok {
		Sayto(b.Frame, name, b.Words["没有国家提示"])
		return
	}
	if len(b.PlayerData[kingdom].TpPos) == 0 {
		Sayto(b.Frame, name, b.Words["没有建立传送点"])
		return
	}
	msg := ""
	list := make(map[string]string)
	num := 0
	for k, _ := range b.PlayerData[kingdom].TpPos {
		relist := map[string]interface{}{
			"i":    num,
			"领地名字": k,
		}
		list[strconv.Itoa(num)] = k
		msg += FormateMsg(b.Frame, relist, b.Menu["返回领地菜单格式"])
		num++
	}
	Sayto(b.Frame, name, msg)
	b.Ctrl.SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) <= 0 {
			Sayto(b.Frame, name, "无效输入")
			return true
		}
		if tpName, isin := list[chat.Msg[0]]; isin {
			pos := b.PlayerData[kingdom].TpPos[tpName]
			cmd := fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", name, pos[0], pos[1], pos[2])
			b.Frame.GetGameControl().SendCmd(cmd)
			Sayto(b.Frame, name, b.Words["返回成功提示"])
		} else {
			Sayto(b.Frame, name, "无效输入")
		}

		return true
	})

}

// 缴费
func (b *King) Pay(name string) {
	kingdom, isok := b.FindkingdomByName(name)
	if !isok {
		Sayto(b.Frame, name, b.Words["没有国王提示"])
		return
	}
	if len(b.PlayerData[kingdom].TerrPos) == 0 {
		Sayto(b.Frame, name, b.Words["没有领地提示"])
	}
	relist := map[string]interface{}{
		"player": name,
	}
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(FormateMsg(b.Frame, relist, b.Content.RatepaySysteam.Cmds), func(output *packet.CommandOutput) {
		if output.SuccessCount <= 0 {
			Sayto(b.Frame, name, b.Words["纳税失败"])
			return
		}
		b.PlayerData[kingdom].ColdTime += b.Content.RatepaySysteam.ColdTime
		relist := map[string]interface{}{
			"秒": b.PlayerData[kingdom].ColdTime,
		}
		Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Words["纳税成功提示"]))
		fmt.Println(FormateMsg(b.Frame, relist, b.Words["纳税成功提示"]))
	})

}

// 提出申请
func (b *King) JoinKing(name string) {
	_, isok := b.FindkingdomByMasterName(name)
	if isok {
		Sayto(b.Frame, name, b.Words["你已经有个王国了"])
		return
	}
	//
	msg := ""
	list := make(map[string]string)
	num := 0
	for k, _ := range b.PlayerData {
		relist := map[string]interface{}{
			"i":    num,
			"国家名字": k,
		}
		list[strconv.Itoa(num)] = k
		msg += FormateMsg(b.Frame, relist, b.Menu["申请加入国家格式"])
		num++
	}
	Sayto(b.Frame, name, msg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) <= 0 {
			return
		}
		if kingomName, isok := list[chat.Msg[0]]; isok {
			//申请加入国家
			b.SubmitApplication(kingomName, name)
		} else {
			Sayto(b.Frame, name, "无效输入")
		}

		return true
	})

}

// 设置传送点
func (b *King) SetTpPos(name string) {
	kingdom, _ := b.FindkingdomByMasterName(name)
	if len(b.PlayerData[kingdom].TerrPos) == 0 {
		Sayto(b.Frame, name, b.Words["没有领地提示"])
		return
	}
	go func() {
		playerPos := <-GetPos(b.Frame, "@a")
		MasterPos := playerPos[name]
		for _, v := range b.PlayerData[name].TerrPos {
			if MasterPos[0] >= v[0] && MasterPos[1] >= v[1] && MasterPos[2] >= v[2] && MasterPos[0] <= (v[0]+b.Content.Extent[0]) && MasterPos[1] <= (v[1]+b.Content.Extent[1]) && MasterPos[2] <= (v[2]+b.Content.Extent[2]) {
				b.DepositTpPos(name, MasterPos, kingdom)
			} else {
				Sayto(b.Frame, name, b.Words["不在领地范围内"])
			}
		}

	}()
}

// 存入传送坐标
func (b *King) DepositTpPos(name string, pos []int, kingdomName string) {
	Sayto(b.Frame, name, b.Words["设置传送地点名字"])
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) <= 0 {
			return true
		}
		//如果重复
		if _, isRe := b.PlayerData[kingdomName].TpPos[chat.Msg[0]]; isRe {
			Sayto(b.Frame, name, b.Words["传送地点名字重复"])
			return true
		}
		b.PlayerData[kingdomName].TpPos[chat.Msg[0]] = pos
		Sayto(b.Frame, name, b.Words["设置传送地点成功"])
		fmt.Println(b.Words["设置传送地点成功"], b.PlayerData[kingdomName])
		return true
	})
}

// 提交申请
func (b *King) SubmitApplication(KingdomName string, name string) {
	//查看国家申请列表中是否含有这个人了
	for k, _ := range b.PlayerData[KingdomName].Application {
		if k == name {
			Sayto(b.Frame, name, "不能重复申请 同一对象")
			return
		}
	}
	b.PlayerData[KingdomName].Application[name] = "这是一个申请"
	Sayto(b.Frame, name, b.Words["申请成功"])
}

// 根据主人名字找国家
func (b *King) FindkingdomByMasterName(name string) (string, bool) {
	for k, v := range b.PlayerData {
		//检查是否为主人
		if v.Master == name {
			return k, true
		}
	}
	return "", false
}

// 根据人名来找国家
func (b *King) FindkingdomByName(name string) (string, bool) {
	for k, v := range b.PlayerData {
		//检查是否为主人
		if v.Master == name {
			return k, true
		}
		//检查人是否为组员
		for memberName, _ := range v.Member {
			if name == memberName {
				return k, true
			}
		}
	}
	return "", false
}

// 检查地皮是否重合
func (b *King) CheckInTerr(pos []int) bool {
	for _, v := range b.PlayerData {
		for _, newPos := range v.TerrPos {
			if b.CheckIsoverlap(pos, b.Content.Extent, newPos, b.Content.Extent) {
				return true
			}
		}
	}
	return false
}

// 检查数组是否重合
func (b *King) CheckIsoverlap(pos []int, Epos []int, Spos []int, SEpos []int) (IsOverlap bool) {
	//检查地皮是否重合
	x1 := pos[0]
	y1 := pos[2]
	x2 := pos[0] + Epos[0]
	y2 := pos[2] + Epos[1]
	x3 := Spos[0]
	y3 := Spos[2]
	x4 := Spos[0] + SEpos[0]
	y4 := Spos[2] + SEpos[1]

	if x1 <= x4 && x3 <= x2 && y1 <= y4 && y3 <= y2 {
		return true
	}
	return false
}

// 循环命令发送器
func (b *King) ReCmdSender(cmds []string, num int, relist map[string]interface{}) {
	if len(cmds)-1 >= num {
		cmd := FormateMsg(b.Frame, relist, cmds[num])
		b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {

			}
		})
	} else {
		return
	}
}

// 来不及了 屎山代码
func (b *King) Activate() {
	go func() {
		time.Sleep(time.Second * 1)
		playerPos := <-GetPos(b.Frame, "@a")
		for k, v := range playerPos {
			playerKingdom, isin := b.FindkingdomByName(k)
			for kingdom, data := range b.PlayerData {
				//免执行
				if isin && kingdom == playerKingdom {
					continue
				}
				//检查是否在内
				for _, terrPos := range data.TerrPos {
					if v[0] >= terrPos[0] && v[1] >= terrPos[1] && v[2] >= terrPos[2] && v[0] <= (terrPos[0]+b.Content.Extent[0]) && v[1] <= (terrPos[1]+b.Content.Extent[1]) && v[2] <= (terrPos[2]+b.Content.Extent[2]) {
						for _, cmd := range b.Content.ProtectSysteam["闯入领地范围执行指令"] {
							relist := map[string]interface{}{
								"player": k,
							}
							b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(FormateMsg(b.Frame, relist, cmd), func(output *packet.CommandOutput) {
								if output.SuccessCount < 0 {
									fmt.Println("指令执行失败", output.OutputMessages)
								}
							})
						}
					}
				}
			}
		}
	}()
}
