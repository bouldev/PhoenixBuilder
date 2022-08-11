package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Guild struct {
	*defines.BasicComponent
	FistCmds               []string              `json:"一级保护指令"`
	ScendCmds              []string              `json:"二级保护指令"`
	MemberName             map[string]string     `json:"成员等级称呼"`
	Guild                  map[string]string     `json:"公会等级称呼"`
	ThePermissionsOfGuild  map[string]int        `json:"最低公会等级可开启的功能"`
	ThePermissionsOfMember map[string]int        `json:"最低权限可使用功能"`
	Triggers               []string              `json:"触发词"`
	Usage                  string                `json:"提示词"`
	MenuGuild              map[string]string     `json:"公会菜单显示"`
	TartgetBuy             string                `json:"购买领地时限制器"`
	DictScore              map[string]string     `json:"各种公会所需计分板"`
	Price                  string                `json:"公会价格"`
	GuildRange             map[string][]int      `json:"公会保护范围"`
	DelayTime              int                   `json:"保护延迟时间[秒]"`
	KeyTitle               map[string]string     `json:"各种提示词"`
	StarGuilds             map[string]*Commodity `json:"公会商店"`
	GuildFristPower        int                   `json:"公会初始等级"`
	TargetOfSetGuildLb     string                `json:"可设置公会权限的选择器"`
	TriggersOfSetGuidb     string                `json:"设置公会权限触发词"`
	PersonScoreTitle       map[string]string     `json:"显示个人信息所需计分板"`
	NoGuild                [][]int               `json:"禁止设置公会坐标"`
	TriggersOfOp           string                `json:"隐藏菜单触发词"`

	GuildData map[string]*GuildDatas
}
type Commodity struct {
	name      string   `json:"商品名字"`
	IdName    string   `json:"商品英文"`
	Price     string   `json:"商品价格"`
	Score     string   `json:"使用的货币"`
	Cmds      []string `json:"购买时执行指令"`
	CheckCmds string   `json:"购买时检测指令"`
}
type GuildDatas struct {
	Master       string
	Member       map[string]*GuildDtails //记得初始化（）
	SpPlace      map[string][]int        //[起点x 起点y 起点z dx dy dz]
	Range        []int
	announcement []string
	Pos          []int
	CenterPos    []int
	Power        int
}
type User struct {
	Name []string `json:"victim"`
}
type GuildDtails struct {
	Permistion string
	title      []string
}

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
	b.Frame.GetJsonData("公会信息.json", &b.GuildData)
	b.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     b.Triggers,
			ArgumentHint: "",
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
	go func() {
		for {
			time.Sleep(time.Duration(b.DelayTime) * time.Second)
			b.ProtectGuildCentry()
		}
	}()

}
func (b *Guild) Stop() error {
	fmt.Print("开始保存公会插件信息")
	return b.Frame.WriteJsonData("公会信息.json", b.GuildData)

}
func (b *Guild) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return b.Frame.WriteJsonDataWithTMP("公会信息.json", ".ckpt", &b.GuildData)
	}
	return nil
}

// -----------------------分界线-----------------------------------
func (b *Guild) Center(chat *defines.GameChat) bool {
	fmt.Print(b.MenuGuild, "\n")
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.FormateMenu(b.Usage, b.MenuGuild))
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
				b.MasterMenu(NewChat.Name)
			case "4":
				b.sayto(chat.Name, "抱歉尚未做好")
				//b.BackGuild(NewChat.Name)
			case "5":
				b.GetPerson(NewChat.Name)
			case b.TriggersOfSetGuidb:

				go func() {
					list, ok := <-b.GetPlayerName(b.TargetOfSetGuildLb)
					if ok {
						for _, j := range list {
							if j == NewChat.Name {

								b.setGuildPower(NewChat.Name)
							}
						}

					}
				}()
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

// 更方便输出
func (b *Guild) sayto(name string, str string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), str)
}

// 规范输出公会名字菜单 name 为发送对象 str为格式 theGuilMap为列表 [i]为数字 [公会名字]]
func (b *Guild) formateGuildNameMenu(name string, str string, theGuildMap map[string]string) {
	menu := ""
	for k, v := range theGuildMap {
		msg := b.FormateMsg(str, "i", k)
		msg = b.FormateMsg(msg, "公会名字", v)
		menu = menu + msg + "\n"
	}
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), menu)
}

// 返回公会详细信息
func (b *Guild) getGuildDataD(name string, guildname string, data GuildDatas) {
	msg := b.KeyTitle["op隐藏菜单公会详细信息显示"]
	msg = b.FormateMsg(msg, "公会名字", guildname)
	msg = b.FormateMsg(msg, "会长", data.Master)
	memberstr := ""
	if len(data.Member) > 0 {
		//n为成员名字
		for n, _ := range data.Member {
			memberstr = memberstr + n + ","
		}
	}
	msg = b.FormateMsg(msg, "成员", memberstr)

	msg = b.FormateMsg(msg, "权限", strconv.Itoa(data.Power))
	msg = b.FormateMsg(msg, "中心坐标", fmt.Sprintf("%v %v %v", strconv.Itoa(data.CenterPos[0]), strconv.Itoa(data.CenterPos[1]), strconv.Itoa(data.CenterPos[2])))
	msg = b.FormateMsg(msg, "起始坐标", fmt.Sprintf("%v %v %v", strconv.Itoa(data.Pos[0]), "-65", strconv.Itoa(data.Pos[2])))
	Range := b.GuildRange[strconv.Itoa(data.Power)]
	msg = b.FormateMsg(msg, "终点坐标", fmt.Sprintf("%v %v %v", strconv.Itoa(data.Pos[0]+Range[0]), "400", strconv.Itoa(data.Pos[2]+Range[1])))
	b.sayto(name, msg)
}

// 查询公会信息
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

// 改变公会的数据
func (b *Guild) chargeGuildData(name string) {
	b.sayto(name, "[输入想要修改的公会名字]")
	theGuildMap := b.getGuildMap()
	b.formateGuildNameMenu(name, b.KeyTitle["设置公会权限菜单模板"], theGuildMap)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if guildname, ok := theGuildMap[chat.Msg[0]]; ok {
				b.sayto(name, "[输入想要修改的权限]\n0 :删除领地\n1 :修改领地权限")
				b.Frame.GetGameControl().SetOnParamMsg(name, func(newchat *defines.GameChat) (catch bool) {
					if len(newchat.Msg) > 0 {
						switch newchat.Msg[0] {
						case "0":
							delete(b.GuildData, guildname)
							b.sayto(name, "[删除成功]")
						case "1":
							b.sayto(name, "[输入权限等级]")
							b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
								if len(Newchat.Msg) > 0 {
									if newchat.Msg[0] == "1" || newchat.Msg[0] == "2" || newchat.Msg[0] == "3" || newchat.Msg[0] == "4" {
										b.GuildData[guildname].Power, _ = strconv.Atoi(Newchat.Msg[0])
										b.sayto(name, "[修改成功]")
									} else {
										b.sayto(name, "【输入有效数字】")
									}
								}
								return true
							})
						}

					}
					return true
				})
			} else {
				b.sayto(name, "[请输入有效数字]")
			}
		}
		return true
	})
}

// 设置op权限菜单
func (b *Guild) setOpMenu(name string) {
	go func() {
		n := <-b.GetPlayerName(b.TargetOfSetGuildLb)
		fmt.Println("n", n)
		for _, v := range n {
			if v == name {
				b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["op隐藏菜单"])
				b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
					if len(chat.Msg) > 0 {
						switch chat.Msg[0] {
						case "0":
							b.findGuildData(chat.Name)
						case "1":
							b.chargeGuildData(chat.Name)
						}

					}

					return true
				})

			}

		}

	}()

}

// 获取公会名字以及对应的号数key为号 v为公会
func (b *Guild) getGuildMap() (List map[string]string) {
	list := make(map[string]string)
	num := 0
	for k, _ := range b.GuildData {
		list[strconv.Itoa(num)] = k
		num++
	}
	return list
}

// 设置公会权限
func (b *Guild) setGuildPower(name string) {
	list := make(map[string]string)
	msg := "[输入对应的数字进行操纵]"
	if len(b.GuildData) > 0 {
		//获取公会名字并成列表给用户并让它选择

		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				//k为公会名字
				if k, ok := list[chat.Msg[0]]; ok {
					b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("[提示] 你选择的是%v公会 请输入更改权限大小(1/2/3/4)", k))
					b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
						if len(Newchat.Msg) > 0 {

							if Newchat.Msg[0] == "1" || Newchat.Msg[0] == "2" || Newchat.Msg[0] == "3" || Newchat.Msg[0] == "4" {
								num, _ := strconv.Atoi(Newchat.Msg[0])
								b.GuildData[k].Power = num
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("[修改成功] %v 更改权限为%v", k, strconv.Itoa(b.GuildData[k].Power)))

							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("请输入有效数字"))

							}

						}
						return true
					})
				}

			}
			return true
		})

	} else {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["暂无公会存在时提示词"])
	}
}

// 商城
func (b *Guild) StarGuild(name string) {
	//判断对方公会是否达到对应级别
	if k, v, _ := b.CheckInGuild(name); v {
		if b.GuildData[k].Power >= b.ThePermissionsOfGuild["商店功能"] {
			//达到级别则弹出商店
			msg := "§b[公会商店]\n输入对应数字来购买对应商品\n"
			for i, j := range b.StarGuilds {
				msg = msg + b.FormateMsg(b.KeyTitle["公会菜单商品显示"], "商品", j.name)
				msg = b.FormateMsg(msg, "价格", j.Price)
				msg = b.FormateMsg(msg, "i", i)
				msg = msg + "\n"
			}
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				for _i, _j := range b.StarGuilds {
					i, j := _i, _j
					//fmt.Println(i, "i", "\nmsg:", chat.Msg[0])
					if len(chat.Msg) > 0 && i == chat.Msg[0] {
						//fmt.Println("触发")
						cmd := b.FormateMsg(j.CheckCmds, "购买者", chat.Name)
						cmd = b.FormateMsg(cmd, "计分板", j.Score)
						cmd = b.FormateMsg(cmd, "购买价格", j.Price)
						fmt.Println("cmd:", cmd)
						//防止错位而专门设置的新名字
						newName := name
						b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
							if output.SuccessCount > 0 {

								for _, c := range j.Cmds {
									c = b.FormateMsg(c, "购买者", newName)
									c = b.FormateMsg(c, "计分板", j.Score)
									c = b.FormateMsg(c, "购买价格", j.Price)
									//发送指令并查看是否正确
									fmt.Println("c:", c)
									b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(c, func(output *packet.CommandOutput) {
										fmt.Println("ouputmsg:", output.OutputMessages)
									})
								}
							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", newName), b.KeyTitle["商店购买失败提示"])
							}
						})
					}
				}

				return true
			})
		}
	} else {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["权限不足时提示"])
	}
}

// 刷新指定的选择器的公会权限分数
func (b *Guild) FlushedPower(target string) {

	go func() {
		NameList, isok := <-b.GetPlayerName(target)

		if isok {
			for _, v := range NameList {
				if GuildNames, ok, _ := b.CheckInGuild(v); ok {
					b.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreaboard players set @a[name=\"%v\"] %v %v", v, b.DictScore["权限计分板"], b.GuildData[GuildNames].Member[v].Permistion))
				}
			}
		}

	}()
}

// 设置他人权限
func (b *Guild) SetOtherPower(SourceName string, name string, guildName string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", SourceName), "[请输入赋予权限等级]")
	b.Frame.GetGameControl().SetOnParamMsg(SourceName, func(chat *defines.GameChat) (catch bool) {
		if b.CheckIsNum(chat.Msg[0]) {
			num, err := strconv.Atoi(chat.Msg[0])

			if err != nil {
				fmt.Print(err)
			} else {
				//检查是否在权限以下
				if _, ok, PowerNum := b.CheckInGuild(SourceName); ok {
					if num >= PowerNum {
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", SourceName), "[警告] 只能操纵权限以下的")
					} else if PowerNum > b.ThePermissionsOfMember["可以改变他人权限"] {

						b.GuildData[guildName].Member[name].Permistion = chat.Msg[0]
						b.FlushedPower(fmt.Sprintf("@a[name=\"%v\"]", name))
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", SourceName), "[授权成功] "+name+" 权限变更为:"+chat.Msg[0])
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "[权限变更] "+name+" 权限变更为:"+chat.Msg[0])

					} else {
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", SourceName), "[警告] 只能操纵权限以下的")
					}
				}
			}
		} else {

		}
		return true
	})

}

// 邀请他人
func (b *Guild) Invite(guildNmae string, name string, AllPLayer map[string]string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.FormateMsg(b.KeyTitle["邀请对方进入公会时提示"], "公会名字", guildNmae))
	b.Frame.GetGameControl().SetOnParamMsg(name, func(Newschat *defines.GameChat) (catch bool) {
		if len(Newschat.Msg) > 0 {
			if Newschat.Msg[0] == "y" {
				b.DeleteMember(name)
				b.GuildData[guildNmae].Member[name] = &GuildDtails{
					Permistion: "0",
				}
				//fmt.Print("成功")
				b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "[加入成功]")
			} else if Newschat.Msg[0] == "n" {
				b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "§e[拒绝成功]")
			}

		}
		return true
	})
}
func (b *Guild) kickMember(guildName string, source string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", source), "请输入要kick的名字")
	b.Frame.GetGameControl().SetOnParamMsg(source, func(chat *defines.GameChat) (catch bool) {
		if _, ok := b.GuildData[guildName].Member[chat.Msg[0]]; ok {
			//如果权限大于于对方则马上踢出
			_, _, PowerNum := b.CheckInGuild(source)
			MemberPowerNum, _ := strconv.Atoi(b.GuildData[guildName].Member[chat.Msg[0]].Permistion)
			if MemberPowerNum < PowerNum {
				b.Frame.GetGameControl().SendCmd(fmt.Sprintf("kick \"%v\"", chat.Msg[0]))
			} else {
				b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", source), "[错误] 无法踢出玩家")
			}
		} else {
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", source), "[错误] 无法在公会成员中找到该玩家")
		}
		return true
	})
}

// 设置特殊区域
func (b *Guild) setSpPlace(name string, guildName string) {
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		go func() {
			starPosSp := <-b.Frame.GetGameControl().GetPlayerKit(name).GetPos("@a[name=[player]]")
			fmt.Println("starPosSp:", starPosSp)
			if b.CheckInGuildPlace(guildName, starPosSp) {
				b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("[确定坐标] %v %v %v \n[终点坐标确定] 在终点输入任意字", strconv.Itoa(starPosSp[0]), strconv.Itoa(starPosSp[1]), strconv.Itoa(starPosSp[2])))
				b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
					fmt.Print(Newchat.Name, "\n")
					go func() {
						EndPos := <-b.Frame.GetGameControl().GetPlayerKit(Newchat.Name).GetPos("@a[name=[player]]")
						fmt.Println("endpos:", EndPos)

						if b.CheckInGuildPlace(guildName, EndPos) {
							b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", Newchat.Name), "请输入特殊区域名字")
							b.Frame.GetGameControl().SetOnParamMsg(Newchat.Name, func(Newchats *defines.GameChat) (catch bool) {
								b.GuildData[guildName].SpPlace = make(map[string][]int)
								b.GuildData[guildName].SpPlace[Newchats.Msg[0]] = []int{
									starPosSp[0],
									starPosSp[1],
									starPosSp[2],
									EndPos[0],
									EndPos[1],
									EndPos[2],
								}
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "[设置成功]")
								return true
							})
						}

					}()

					return true
				})
			} else {
				b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "请在领地范围内")
			}

		}()
		return true
	})
}

// 格式化输出
func (b *Guild) FromatePrint(msg string, T string) {
	fmt.Print(fmt.Sprintf("----------------\n[%v]  :   %v\n-------------------\n", T, msg))
}

// 会长菜单&&管理员菜单
func (b *Guild) MasterMenu(name string) {
	msg := b.KeyTitle["管理菜单显示"]
	//fmt.Print(msg)
	//n,ok,PowerNum :=b.CheckInGuild(name)

	//必须是管理员级别才能菜单执行 否则不予理会
	if n, _, PowerNum := b.CheckInGuild(name); PowerNum >= 2 {
		//发送菜单等待回执
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				//fmt.Print("--test--320")

				go func() {
					k, isOK := <-b.GetPlayerName("@a")
					if isOK {
						//fmt.Print("--test--320")
						switch chat.Msg[0] {
						case "0":

							//fmt.Print("name:", k)
							msgOfPlayer := "§e[请输入对应数字邀请对方]\n"

							for i, PlayerName := range k {
								msgOfPlayer = msgOfPlayer + fmt.Sprintf("[%v] %v", i, PlayerName)
							}
							b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msgOfPlayer)
							b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
								if len(Newchat.Msg) > 0 {
									if _, isok := k[Newchat.Msg[0]]; isok {

										if yes, _ := b.CheckIsMaster(k[Newchat.Msg[0]]); yes {
											//fmt.Print(yes)
											b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "[错误] 对方为会长级别")
										} else {
											//fmt.Print("yes:", yes)
											b.Invite(n, k[Newchat.Msg[0]], k)
										}

									}
								} else {
									b.FromatePrint("输入错误", "提示信息")
								}

								return true
							})

						case "1":
							//达到权限才能使用否则警告
							if PowerNum >= b.ThePermissionsOfMember["可以改变他人权限"] {
								MenuMsg := "§b[设置他人权限]"
								ListMap := make(map[string]string)
								num := 0
								for i, _ := range b.GuildData[n].Member {
									MenuMsg = MenuMsg + fmt.Sprintf("[%v] %v", num, i)
									ListMap[strconv.Itoa(num)] = i
									num++
								}
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), MenuMsg)
								b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
									if len(Newchat.Msg) > 0 {
										if _, isok := ListMap[Newchat.Msg[0]]; isok {
											b.SetOtherPower(Newchat.Name, ListMap[Newchat.Msg[0]], n)
										} else {
											b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", Newchat.Name), "[请输入有效数字]")
										}
									}

									return true
								})
							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])
							}
						case "2":
							if PowerNum >= b.ThePermissionsOfMember["可以发布公告"] {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), "[请输入公告内容]")
								b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(Newchats *defines.GameChat) (catch bool) {
									b.GuildData[n].announcement = append(b.GuildData[n].announcement, Newchats.Msg[0])
									b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", Newchats.Name), fmt.Sprintf("[发布公告] 内容为:%v", Newchats.Msg[0]))
									return true
								})
							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])
							}
						case "3":
							if b.GuildData[n].Power < b.ThePermissionsOfGuild["设置二级保护区域"] {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])
							} else if PowerNum >= b.ThePermissionsOfMember["可以设置二级保护区"] {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), "[请站在起点] 输入任意文字")
								fmt.Println("公会名字为:", n)
								b.setSpPlace(chat.Name, n)

							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])

							}
						case "4":
							b.kickMember(n, chat.Name)
						case "5":
							delete(b.GuildData, n)
							b.sayto(name, "[删除成功]")
						case "6":
							b.getGuildDataD(name, n, *b.GuildData[n])
						case "7":

							b.DeleteGuildMember(name, *b.GuildData[n])
						}

					} else {
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "[错误] 没有找到目标")
					}
				}()

			}

			return true
		})
	}

}

// 删除指定成员
func (b *Guild) DeleteGuildMember(master string, data GuildDatas) {
	if len(data.Member) > 0 {
		b.sayto(master, "[删除指定成员]")
		listOfMember := make(map[string]string)
		num := 0
		for k, _ := range data.Member {
			listOfMember[strconv.Itoa(num)] = k
		}
		b.formateGuildNameMenu(master, b.KeyTitle["设置公会权限菜单模板"], listOfMember)
		b.Frame.GetGameControl().SetOnParamMsg(master, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				if n, ok := listOfMember[chat.Msg[0]]; ok {
					delete(data.Member, n)
					b.sayto(master, "[删除成功]")
				} else {
					b.sayto(master, "[输入有效数字]")
				}
			}
			return true
		})
	} else {
		b.sayto(master, "[无成员]")
	}

}

// 获取公会排名
func (b *Guild) BackGuild(name string) {

}

//获取个人信息

func (b *Guild) GetPerson(name string) {
	//获取公会名字
	//fmt.Print("开始测试")
	if guildname, ok, PowerNum := b.CheckInGuild(name); ok {
		msg := b.FormateMsg(b.KeyTitle["显示个人信息菜单开头"], "公会名字", guildname)
		msg = b.FormateMsg(msg, "公会权限", strconv.Itoa(b.GuildData[guildname].Power))
		msg = b.FormateMsg(msg, "本人权限", strconv.Itoa(PowerNum))
		//k为计分板 v 为显示名字
		go func() {
			list := <-b.GetScore()
			fmt.Println("scoresMap", list)
			for k, v := range b.PersonScoreTitle {
				//fmt.Println("k", k, "v:", v)
				if score, isok := list[name][k]; isok {
					//fmt.Println("成功")
					msg = msg + b.FormateMsg(b.KeyTitle["显示个人信息模板"], "计分板名字", v)
					msg = b.FormateMsg(msg, "计分板分数", strconv.Itoa(score))
				} else {
					b.FromatePrint(fmt.Sprintf("抱歉该计分板并不存在 %v", k), "报错")
				}

			}
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
			//msg = b.FormateMsg(b.KeyTitle["显示个人信息模板"])
		}()

	} else {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["无地皮时提示"])
	}
}

func (b *Guild) ProtectGuildCentry() {
	//time.Sleep(time.Duration(b.DelayTime) * time.Second)
	//这里发送指令是所有公会的两条指令在1秒内发送
	if len(b.GuildData) > 0 {

		for _k, _v := range b.GuildData {
			//间隔200毫秒每个字典之间
			time.Sleep(time.Millisecond * time.Duration(200))
			k, v := _k, _v
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
			}
			//第二指令保护

			for _, j := range b.ScendCmds {
				/*

					//j = b.FormateMsg(j,"领地范围内非会员",str)
					//只有有权限的才会启动
				*/
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
						b.Frame.GetGameControl().SendCmd(j)
					}

					//fmt.Println("j:", j)

				}

			}

		}
	}
}

// 让指定的name返回领地
func (b *Guild) TpBack(name string) bool {
	for k, v := range b.GuildData {
		if v.Master == name {
			msg := b.FormateMsg(b.KeyTitle["返回地皮时提示词"], "公会名字", k)
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
			fmt.Print("centerpos", v.CenterPos)
			fmt.Print(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", name, v.CenterPos[0], v.CenterPos[1], v.CenterPos[2]))
			b.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", name, v.CenterPos[0], v.CenterPos[1], v.CenterPos[2]))
			return true
		}
		if len(v.Member) > 0 {
			for i, _ := range v.Member {
				if i == name {
					msg := b.FormateMsg(b.KeyTitle["返回地皮时提示词"], "公会名字", k)
					fmt.Print("centerpos", v.CenterPos)
					b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
					b.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", name, v.CenterPos[0], v.CenterPos[1], v.CenterPos[2]))
					return true
				}
			}
		}

	}
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["无地皮时提示"])
	return false
}

// 删除公会成员
func (b *Guild) DeleteMember(name string) {
	if len(b.GuildData) > 0 {
		for _, v := range b.GuildData {
			if _, ok := v.Member[name]; ok {
				delete(v.Member, name)
			}
		}
	}

}

// 检测是否
// 写入公会信息
func (b *Guild) WriteGuildData(name string, guildName string) {
	k, v := b.CheckIsMaster(name)
	if k == false {

		go func() bool {
			//获取坐标
			PosOf := <-b.Frame.GetGameControl().GetPlayerKit(name).GetPos("@a[name=[player]]")
			//fmt.Println("posO:", PosOf)
			if len(PosOf) > 0 {
				//fmt.Println("b.guildrange:", b.GuildRange)
				//fmt.Println("强制转换:", int(b.GuildRange["4"][0]/2))
				StarPos := []int{
					PosOf[0] - int(b.GuildRange[strconv.Itoa(b.GuildFristPower)][0]/2),
					PosOf[1],
					PosOf[2] - int(b.GuildRange[strconv.Itoa(b.GuildFristPower)][1]/2),
				}
				fmt.Println("satrPos:", StarPos)
				if len(b.GuildData) > 0 {
					//遍历所有公会

					for _, j := range b.GuildData {
						//如果在成员列表里面就删除
						if _, ok := j.Member[name]; ok {
							delete(j.Member, name)
							b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "§e§l[成功删除你的原有公会]")
						}
						//地皮重合就退出
						//posOf为地皮正中心位

						fmt.Println("satrPos:", StarPos)
						//遍历地皮的中心坐标
						JstarPos := []int{
							j.Pos[0] - int(b.GuildRange["4"][0]),
							j.Pos[1],
							j.Pos[2] - int(b.GuildRange["4"][1]),
						}
						//fmt.Println("jsatrPos:", JstarPos)
						if b.CheckIsoverlap(StarPos, b.GuildRange["4"], JstarPos, b.GuildRange["4"]) {

							b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "§c§l[错误] §e附近有他人公会")
							return false
						} else if len(b.NoGuild) > 0 {
							for _, nopos := range b.NoGuild {
								if b.CheckIsoverlap(StarPos, b.GuildRange["4"], nopos, b.GuildRange["4"]) {
									return false
								}
							}

						}

					}

				}
				//b.CheckIsoverlap(PosOf,b.GuildRange["0"],)
				//正式购买
				b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", name, b.DictScore["购买公会计分板"], b.Price), func(output *packet.CommandOutput) {
					if output.SuccessCount > 0 {
						//fmt.Print("---test--137\n")
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§b[扣除成功]§e 消费§l§a%v %v", b.DictScore["购买计分板名字"], b.Price))
						//初始化各种信息
						Ranges := b.GuildRange[strconv.Itoa(b.GuildFristPower)]
						b.GuildData[guildName] = &GuildDatas{
							Master:    name,
							Member:    make(map[string]*GuildDtails),
							Pos:       StarPos,
							CenterPos: PosOf,
							Range:     Ranges,
							Power:     b.GuildFristPower,
						}

						b.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreboard players set @a[name=\"%v\"] %v 4", name, b.DictScore["权限计分板"]))
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "§b§l[购买成功] 获取公会")
						msg := b.FormateMsg(b.KeyTitle["购买领地后提示"], "公会名字", guildName)
						msg = b.FormateMsg(msg, "起点坐标", strconv.Itoa(StarPos[0])+" -70 "+strconv.Itoa(StarPos[2]))
						msg = b.FormateMsg(msg, "终点坐标", strconv.Itoa(StarPos[0]+Ranges[0])+" 400 "+strconv.Itoa(StarPos[2]+Ranges[1]))
						msg = b.FormateMsg(msg, "中心坐标", strconv.Itoa(PosOf[0])+" "+strconv.Itoa(PosOf[1])+" "+strconv.Itoa(PosOf[2]))
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
						fmt.Println(msg)
						//fmt.Printf("信息:会长:%v\n成员:%v\n坐标:%v\n范围:%v\n", b.GuildData[guildName].Master, b.GuildData[guildName].Member, b.GuildData[guildName].Pos, b.GuildData[guildName].Range)
					} else {
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§c[扣除失败] 原因未知 请让腐竹检查配置文件是否正确"))
					}
				})
			}
			return true

		}()

	} else if k == true {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§a[购买失败] %v你已经是§b%v§a的会长了", name, v))
	} else {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "意料之外的情况 稍后重试")
	}
}

// 检查坐标是否在某个公会的领地范围内
func (b *Guild) CheckInGuildPlace(guildName string, pos []int) bool {
	fmt.Println("pos:", pos)
	fmt.Println("guildpos:", b.GuildData[guildName].Pos)
	x := b.GuildData[guildName].Pos[0]
	z := b.GuildData[guildName].Pos[2]
	dx := b.GuildData[guildName].Pos[0] + b.GuildData[guildName].Range[0]
	dz := b.GuildData[guildName].Pos[2] + b.GuildData[guildName].Range[1]
	if pos[0] <= dx && pos[0] >= x && pos[2] <= dz && pos[2] >= z {
		return true
	}
	return false
}

// 正则表达检查字符串是否为数字
func (b *Guild) CheckIsNum(str string) bool {
	ok, _ := regexp.MatchString("^\\+?[1-9][0-9]*$", str)
	return ok
}

// 检查是否在公会内 如果存在则返回对应权限与公会名字 true 不存在则返回"" false -1
func (b *Guild) CheckInGuild(name string) (guildname string, ok bool, Power int) {
	for k, v := range b.GuildData {
		//如果是会长
		if v.Master == name {
			return k, true, 4

		} else if _, isok := v.Member[name]; isok {
			num, _ := strconv.Atoi(v.Member[name].Permistion)
			return k, true, num
		}

	}
	return "", false, -1

}

// 检查name是否为某个公会的会长 如果是则返回true guildname
func (b *Guild) CheckIsMaster(name string) (yes bool, GuidName string) {
	//fmt.Println("启动")
	//defer fmt.Print("结束")
	for k, v := range b.GuildData {
		fmt.Print("v:", v.Master)
		if v.Master == name {
			return true, k
		}
	}
	return false, ""
}

// 检查公会名字是否重合
func (b *Guild) ChecklGuildName(name string) (yes bool) {
	if len(b.GuildData) > 0 {
		for k, _ := range b.GuildData {
			if k == name {
				return true
			}
		}

	}
	return false
}

// 购买公会
func (b *Guild) BuyGuild(name string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(b.FormateMsg(b.FormateMsg(b.FormateMsg(fmt.Sprintf("testfor %v", b.TartgetBuy), "player", name), "购买计分板", b.DictScore["购买公会计分板"]), "价格", b.Price), func(output *packet.CommandOutput) {
		fmt.Println("outputmessage : ", output.OutputMessages)
		if output.SuccessCount > 0 {
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§a[输入公会名字]:"))
			if b.Frame.GetGameControl().SetOnParamMsg(name, func(NewChat *defines.GameChat) (catch bool) {
				GuildName := NewChat.Msg[0]
				//检查名字重合与坐标重合与否（）
				if b.ChecklGuildName(GuildName) == false { //&& b.CheckIsoverlap() {
					b.WriteGuildData(name, GuildName)

				} else {
					b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§c[名字已经被注册了] 请重新购买"))
				}

				return true
			}) == nil {

			}

		} else {
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§c§l[错误]§a 无法购买请满足选择器:§b\n%v\n", b.FormateMsg(b.FormateMsg(b.FormateMsg(fmt.Sprintf("%v", b.TartgetBuy), "player", name), "购买计分板", b.DictScore["购买公会计分板"]), "价格", b.Price)))
		}
	})
}

// 格式化菜单 第一个参数是菜单的名字 第二个是菜单内容所对应的格式用map来存储
func (b *Guild) FormateMenu(name string, list map[string]string) (str string) {
	//fmt.Print(list, "\n")
	msg := name + "\n"
	if len(list) > 0 {
		num := 0
		//for k, v := range list {
		//fmt.Print("--test--", list)
		for k, _ := range list {
			fmt.Print(k)
			msg = msg + b.FormateMsg(list[strconv.Itoa(num)], "i", strconv.Itoa(num)) + "\n"
			num++
		}

		//}
	}
	return msg

}

// 替换指定的信息
func (b *Guild) FormateMsg(str string, re string, afterstr string) (newstr string) {

	res := regexp.MustCompile("\\[" + re + "\\]")
	return res.ReplaceAllString(str, afterstr)

}

// 获取所有人的积分 返回通道
func (b *Guild) GetScore() (PlayerScoreList chan map[string]map[string]int) {

	cmd := "scoreboard players list @a"
	GetScoreChan := make(chan map[string]map[string]int, 2)
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount >= 0 {
			List := make(map[string]map[string]int)
			gamePlayer := ""
			for _, i := range output.OutputMessages {
				//fmt.Println(i)
				if len(i.Parameters) == 2 {
					//fmt.Println("判定为人")
					gamePlayer = strings.Trim(i.Parameters[1], "%")
					List[gamePlayer] = make(map[string]int)
				} else if len(i.Parameters) == 3 {
					//fmt.Println("判定为分数")
					key, _ := strconv.Atoi(i.Parameters[0])
					List[gamePlayer][i.Parameters[1]] = key
				} else {
					continue
				}
			}
			if gamePlayer != "" && len(List) >= 1 {
				GetScoreChan <- List
			}
		}
	})
	return GetScoreChan

}

// 检查一个数组是否有某个元素
func (b *Guild) CheckArr(arr []interface{}, str interface{}) (IsIn bool) {
	if len(arr) == 0 {
		fmt.Print("数组为空")
		return false
	} else {
		var set map[interface{}]struct{}
		set = make(map[interface{}]struct{})
		for _, value := range arr {
			set[value] = struct{}{}
		}
		// 检查元素是否在map
		if _, ok := set[str]; ok {
			return true
		} else {
			return false
		}

	}

}

// 获取指定限制器的玩家名字 返回通道值
func (b *Guild) GetPlayerName(name string) (listChan chan map[string]string) {

	var Users User
	//var UsersListChan chan []string
	UsersListChan := make(chan map[string]string, 2)
	//OkChan := make(chan bool, 2)
	//fmt.Print("test")
	//isok := false
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse("testfor "+name, func(output *packet.CommandOutput) {
		//fmt.Print(",,,,,,,,,,,,,,,,,,")
		//fmt.Print(output.DataSet)
		if output.SuccessCount > 0 {
			json.Unmarshal([]byte(output.DataSet), &Users)

			//var mapName map[string]string
			//fmt.Print("Users:", Users)
			mapName := make(map[string]string, 40)
			for k, v := range Users.Name {
				mapName[strconv.Itoa(k)] = v
			}

			//isok = true
			//fmt.Print("isok:", isok)
			UsersListChan <- mapName
			//OkChan <- true
		}

	})

	//fmt.Print("isok:", isok)
	return UsersListChan
}
func (b *Guild) CheckIsoverlap(pos []int, Epos []int, Spos []int, SEpos []int) (IsOverlap bool) {
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
