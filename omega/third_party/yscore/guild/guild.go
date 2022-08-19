package guild

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"time"

	"github.com/pterm/pterm"
)

func (b *Guild) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	b.GuildData = make(map[string]*GuildDatas)

}
func (b *Guild) Inject(frame defines.MainFrame) {
	b.Frame = frame
	//注入frame等东西
	/*
		b.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddItemActor, func(p packet.Packet) {
			fmt.Print("凋落物的包:", p, "\n")
		})
	*/

	b.BasicComponent.Inject(frame)
	b.Frame.GetJsonData("新公会信息.json", &b.GuildData)
	b.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     b.Triggers,
			ArgumentHint: " ",
			FinalTrigger: false,
			Usage:        b.Usage,
		},
		OptionalOnTriggerFn: b.Center,
	})

}
func (b *Guild) Activate() {
	//fmt.Print(b.GetPlayerPos(), "\n")
	b.Frame.GetGameControl().SendCmd("scoreboard objectives add " + b.DictScore["权限计分板"] + "dummy omega权限计分板")
	b.Frame.GetGameControl().SendCmd("scoreboard objectives add " + b.DictScore["购买公会计分板"] + " dummy 公会购买货币")
	b.Frame.GetGameControl().SendCmd("scoreboard objectives add " + b.DictScore["公会同步积分计分板"] + " dummy 公会计分板")
	b.Frame.GetGameControl().SendCmd("scoreboard objectives add " + b.DictScore["兑换贡献计分板"] + " dummy 公会计分板")
	b.Frame.GetGameControl().SendCmd("scoreboard objectives add " + b.DictScore["贡献计分板"] + " dummy 公会计分板")
	if b.IsNeedTerr == false {
		pterm.Info.Println(fmt.Sprintf("注意你现在的模式为默认不挟带地皮 如果你需要地皮的话请在配置文件中把 {是否携带地皮} 后面的false改为true \n本组件的地皮与地皮组件有冲突请不要同时开启"))

	}
	b.GetPos("@a")
	//开启yscore专属组件
	if b.IsYsCore {
		pterm.Info.Println(fmt.Sprintf("注意 你选择了打开yscore组件 那么系统会自动检测你是否为yscore用户 如果不是则会跳过yscore公会专属组件 但是不会影响正常使用"))
		b.CreateNameHashOfGuild()
		go b.GoYscore()
	}
	go func() {
		//初始化公会的分数
		for _, v := range b.GuildData {
			if v.GuildRankings == 0 {
				v.GuildRankings = len(b.GuildData) + 1
			}
		}
		pterm.Info.Println(fmt.Sprintf("提示当前 公会保护延迟为:%d,配置文件中的一级保护指令或者二级保护指令均只会在别人在领地且满足对应选择器时内时触发\n比如你写了一级保护指令触发为:@a[m=0]那么当别人(非公会成员)处于领地且满足为生存模式则会触发", b.DelayTime))
		for {
			time.Sleep(time.Duration(b.DelayTime) * time.Second)
			go b.RefreshTheScore()
			b.ProtectGuildCentry()
		}
	}()

}
func (b *Guild) Stop() error {
	fmt.Print("开始保存公会插件信息")
	return b.Frame.WriteJsonData("新公会信息.json", b.GuildData)

}
func (b *Guild) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return b.Frame.WriteJsonDataWithTMP("新公会信息.json", ".ckpt", &b.GuildData)
	}
	return nil
}

/*
"公会列表上公会名字中加个公会等级"
*/
// -----------------------分界线-----------------------------------
func (b *Guild) Center(chat *defines.GameChat) bool {
	//fmt.Print(b.MenuGuild, "\n")

	//b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.FormateMenu(b.Usage, b.MenuGuild))
	b.sendMenu(chat.Name, b.KeyTitle["公会菜单显示"])
	//b.Frame.GetGameControl().PlaceCommandBlock()
	if b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(NewChat *defines.GameChat) (catch bool) {
		if len(NewChat.Msg) > 0 {
			switch NewChat.Msg[0] {
			case "0":
				b.BuyGuild(NewChat.Name)
			case "1":
				b.TpBack(NewChat.Name)
			case "2":
				b.StarGuild(NewChat.Name)
			case "3":
				//b.sayto(chat.Name, "抱歉尚未做好")
				b.QuitTheGuild(NewChat.Name)
				//b.BackGuild(NewChat.Name)
			case "4":
				b.bourseMenu(NewChat.Name)
			case "5":
				b.MasterMenu(NewChat.Name)
			case "6":
				b.GetPerson(NewChat.Name)
			case "7":
				b.GetGuildDataMenu(NewChat.Name)
			case "8":
				if b.IsYsCore {
					b.YscoreMenu(NewChat.Name)
				} else {
					b.sayto(NewChat.Name, "[权限不足]")
				}
			case b.TriggersOfOp:
				fmt.Print("执行")
				b.setOpMenu(NewChat.Name)
			}

		}

		return true
	}) == nil {

	}
	return false
}

// 查询公会信息菜单
func (b *Guild) findGuildData(name string) {

	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "[输入要寻找的公会名字]")
	theGuildMap := b.getGuildMap()
	b.formateGuildNameMenu(name, b.KeyTitle["设置公会权限菜单模板"], theGuildMap)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(newchat *defines.GameChat) (catch bool) {
		if len(newchat.Msg) > 0 {
			if guildname, ok := theGuildMap[newchat.Msg[0]]; ok {
				data := b.GuildData[guildname]
				//发送
				b.getGuildDataD(name, guildname, *data)
			} else {
				b.sayto(name, "[请输入有效数字]")
			}
		}
		return true
	})
}

// 保护中心
func (b *Guild) ProtectGuildCentry() {
	//如果大于零则执行保护指令

	if len(b.GuildData) > 0 {
		go func() {
			PosOfAll := <-b.GetPos(b.FistCmdTarget)
			for guildname, v := range b.GuildData {

				//该地皮有地皮才执行保护指令
				if v.IsTerr {
					for user, pos := range PosOfAll {

						if b.CheckInGuildPlace(guildname, pos) {
							b.sendFirsetCmds(user, guildname)
						}
					}

				}
			}
		}()

	}

	//time.Sleep(time.Duration(b.DelayTime) * time.Second)
	//这里发送指令是所有公会的两条指令在1秒内发送
	/*
		if len(b.GuildData) > 0 {

			for _k, _v := range b.GuildData {
				//间隔200毫秒每个字典之间

				time.Sleep(time.Millisecond * time.Duration(200))
				k, v := _k, _v

				if v.IsTerr {
					X, Y, Z := v.Pos[0], v.Pos[1], v.Pos[2]
					GuildName := k
					//获取反取
					str := fmt.Sprintf("name=!%v", v.Master)
					if len(v.Member) > 0 {
						for member, _ := range v.Member {
							str = str + fmt.Sprintf(",name=!%v", member)
						}
					}
					//第一指令

					for _, j := range b.FistCmds {
						j = b.FormateMsg(j, "x", strconv.Itoa(X))
						j = b.FormateMsg(j, "z", strconv.Itoa(Z))
						j = b.FormateMsg(j, "y", strconv.Itoa(Y))
						j = b.FormateMsg(j, "公会名字", GuildName)
						j = b.FormateMsg(j, "领地范围内非会员", str)
						//fmt.Println("j:", j)

						Dx := b.GuildRange[strconv.Itoa(v.Power)][0]
						Dz := b.GuildRange[strconv.Itoa(v.Power)][1]
						j = b.FormateMsg(j, "Dx", strconv.Itoa(Dx))
						j = b.FormateMsg(j, "Dz", strconv.Itoa(Dz))

						//fmt.Println("141 指令:", j)
						b.Frame.GetGameControl().SendCmd(j)
						//b.CmdSender(j)
					}
					//第二指令保护

					for _, j := range b.ScendCmds {


							//j = b.FormateMsg(j,"领地范围内非会员",str)
							//只有有权限的才会启动

						//只有有权限的才会启动
						if v.Power >= b.ThePermissionsOfGuild["二级保护"] {
							for _, o := range v.SpPlace {
								j = b.FormateMsg(j, "x", strconv.Itoa(o[0]))
								j = b.FormateMsg(j, "z", strconv.Itoa(o[2]))
								j = b.FormateMsg(j, "y", strconv.Itoa(o[1]))
								j = b.FormateMsg(j, "Dx", strconv.Itoa(o[3]-o[0]))
								j = b.FormateMsg(j, "Dz", strconv.Itoa(o[4]-o[2]))
								j = b.FormateMsg(j, "Dy", strconv.Itoa(o[5]-o[1]))
								j = b.FormateMsg(j, "公会名字", GuildName)
								j = b.FormateMsg(j, "权限计分板", b.DictScore["权限计分板"])
								j = b.FormateMsg(j, "最低权限", strconv.Itoa(b.ThePermissionsOfMember["可以进入二级保护区域"])) //)
								//fmt.Println("j--521--", j)
								//b.CmdSender(j)
								b.Frame.GetGameControl().SendCmd(j)
							}

							//fmt.Println("j:", j)

						}

					}
				}

			}
		}
	*/
}

// 检测是否
