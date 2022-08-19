package guild

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

type Yscore struct {
	StrongHold              map[string]*StrongHoldData `json:"yscore专属据点"`
	YscoreMenuTitle         string                     `json:"yscore菜单显示"`
	StrongholdDataTitle     string                     `json:"据点详细信息模板"`
	StrongHoldMenuTitle     string                     `json:"据点菜单"`
	OccupiedStrongHoldTitle string                     `json:"占领据点时广播"`
	FreeStrongHoldTitle     string                     `json:"据点释放时广播"`
	TitleOfScore            string                     `json:"据点内分数显示"`
	AllyMenuTitle           string                     `json:"盟友菜单"`
	HolyRelicsMenu          string                     `json:"圣遗物菜单"`
	YscoreBuff              map[string]*YscoreBuffData `json:"圣遗物与对应buff"` //圣遗物 对应buff map[圣遗物][buff][等级]
	TpMenu                  string                     `json:"传送菜单"`
}
type YscoreBuffData struct {
	Buff  map[string]int `json:"buff"`
	Price string         `json:"价格"`
}
type StrongHoldData struct {
	Pos        []int               `json:"坐标"`
	Range      []int               `json:"范围"`
	Buff       map[string]int      `json:"占领后buff与等级"`
	OpenTime   *StrongHoldOpenTime `json:"开启时间"`
	IsOccupied bool                //是否被占领
	GuildName  string              //占领公会为
	Score      int
}
type PlayerInStrongHold struct {
	Member    map[string]string //成员名字 以及所属公会
	GuildName map[string]string //公会名字
}
type StrongHoldOpenTime struct {
	WeekDate  []string `json:"星期几"`
	StartTime int      `json:"当天开启时间(24以内)"`
	EndTime   int      `json:"当天结束时间(24以内)"`
}
type Buff struct {
	Buff map[string]int
}

// yscore主菜单
func (b *Guild) YscoreMenu(name string) {
	b.sayto(name, b.YsCoreDefines.YscoreMenuTitle)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.StrongHoldMenu(name)
			case "1":
				b.AllyMenu(name)
			case "2":
				b.ExchangeGuildScore(name)
			case "3":
				b.SetInfoOfGuild(name)
			case "4":
				b.HolyRelics(name)
			case "5":
				b.SetTpPosOfYscore(name)
			case "6":
				b.TpBackToYscore(name)
			case "7":
				//展示公会列表 并加入
			}
		}
		return true
	})
}

// 检查是否在据点范围内返回bool和所在据点名字
func (b *Guild) CheckPlayerInStrongHold(pos []int) (ok bool, strongholdName string) {
	for k, v := range b.YsCoreDefines.StrongHold {
		//fmt.Println("pos", pos)
		if pos[0] >= v.Pos[0] && pos[1] >= v.Pos[1] && pos[2] >= v.Pos[2] && pos[0] <= (v.Pos[0]+v.Range[0]) && pos[1] <= (v.Pos[1]+v.Range[1]) && pos[2] <= (v.Pos[2]+v.Range[2]) {
			//fmt.Println("满足条件")
			return true, k
		}
	}
	return false, ""
}

// 检查所有人坐标是否在据点范围内如果存在则返回存在列表 member[据点]玩家名字 guildDic[公会名字]随便啥
func (b *Guild) GetPlayerInStrongHold(list map[string][]int) (member map[string]string, guildDic map[string]string) {
	//fmt.Println(list)
	_member := make(map[string]string)
	_guildDic := make(map[string]string)
	for k, v := range list {
		if ok, strongHoldName := b.CheckPlayerInStrongHold(v); ok {
			_member[strongHoldName] = k
			if guildname, isok, _ := b.CheckInGuild(k); isok {
				_guildDic[guildname] = "这里有个公会"
			}

		}
	}

	return _member, _guildDic
}

// 显示积分对应据点
func (b *Guild) TitleOfScores(StrongHold string) {
	X := b.YsCoreDefines.StrongHold[StrongHold].Pos[0]
	Y := b.YsCoreDefines.StrongHold[StrongHold].Pos[1]
	Z := b.YsCoreDefines.StrongHold[StrongHold].Pos[2]
	Dx := b.YsCoreDefines.StrongHold[StrongHold].Range[0]
	Dy := b.YsCoreDefines.StrongHold[StrongHold].Range[1]
	Dz := b.YsCoreDefines.StrongHold[StrongHold].Range[2]
	//fmt.Println("test______________")
	msg := b.YsCoreDefines.TitleOfScore

	msg = b.FormateMsg(msg, "分数", strconv.Itoa(b.YsCoreDefines.StrongHold[StrongHold].Score))
	if b.YsCoreDefines.StrongHold[StrongHold].IsOccupied {
		msg = b.FormateMsg(msg, "isok", "占领中")
	} else {
		msg = b.FormateMsg(msg, "isok", "空置中")
	}
	msg = b.FormateMsg(msg, "据点", StrongHold)
	b.CmdSender(fmt.Sprintf("title @a[x=%v,y=%v,z=%v,dy=%v,dz=%v,dx=%v] actionbar %v", X, Y, Z, Dy, Dz, Dx, msg))
}

// 初始化
func (b *Guild) InitStrongHold() {
	for _, v := range b.YsCoreDefines.StrongHold {
		v.IsOccupied = false
		v.GuildName = ""
		v.Score = 0
	}
	//b.CmdSender(fmt.Sprintf("scoreboard objectives add %v dummy", b.YsCoreDefines.YscoreScore))
	//初始化圣遗物

}

// 检查据点是否开启
func (b *Guild) CheckStrongHoldIsOpen(strongholdName string) bool {

	weekName := time.Now().Weekday().String()
	for _, v := range b.YsCoreDefines.StrongHold[strongholdName].OpenTime.WeekDate {
		//查看当前时间是否符合
		if v == weekName && time.Now().Hour() >= b.YsCoreDefines.StrongHold[strongholdName].OpenTime.StartTime && time.Now().Hour() <= b.YsCoreDefines.StrongHold[strongholdName].OpenTime.EndTime {
			return true
		}
	}
	return false
}

// 将buff分数加入在内
func (b *Guild) AddBuffInBuffList(buffname string, level int, buffList map[string]*Buff, guildname string) map[string]*Buff {
	//fmt.Println(buffList)
	for k, _ := range buffList {
		if buffList[k].Buff == nil {
			fmt.Println("为空")
			buffList[k].Buff = make(map[string]int)
		}
	}
	//查询是否公会buff 没有则创建新的 有则对比
	if socre, ok := buffList[guildname]; ok && socre.Buff[buffname] < level {
		buffList[guildname].Buff[buffname] = level
	} else if !ok {
		buffList[guildname] = &Buff{
			Buff: map[string]int{
				buffname: level,
			},
		}
	}
	return buffList
}

// 规范的加上buff
func (b *Guild) SendBuff(bufflist map[string]*Buff, playerlist map[string][]int) {
	for k, v := range bufflist {
		for buff, level := range v.Buff {
			//每次buff发送间隙停顿50毫秒
			time.Sleep(time.Millisecond * 50)
			b.CmdSender(fmt.Sprintf("effect @a[scores={%v=%v}] %v %v %v true", b.DictScore["公会同步积分计分板"], strconv.Itoa(b.GuildData[k].GuildRankings), buff, "20", strconv.Itoa(level)))
		}
	}
}

// 判断当前领地状态 并做对应处理
func (b *Guild) Judgment(member map[string]string, guildDic map[string]string) {
	//遍历所有有人的据点
	//fmt.Println(member, "\n", guildDic)
	//fmt.Println("test")
	for k, _ := range member {
		score := b.YsCoreDefines.StrongHold[k].Score
		IsOc := b.YsCoreDefines.StrongHold[k].IsOccupied
		//fmt.Println(score, "\n", IsOc)
		if score >= 0 && score < 100 && IsOc == false {
			b.YsCoreDefines.StrongHold[k].Score++
			//fmt.Println(fmt.Sprintf("%v 分数:%v/100", k, score))
			b.TitleOfScores(k)
			//分数为0-100之间 且占领状态为false
		} else if score == 0 && IsOc == true {
			//分数为零且占领状态则不扣分然后重置状态
			b.YsCoreDefines.StrongHold[k].IsOccupied = false
			b.YsCoreDefines.StrongHold[k].GuildName = ""
			msg := b.YsCoreDefines.FreeStrongHoldTitle
			msg = b.FormateMsg(msg, "据点名字", k)
			b.Frame.GetGameControl().SayTo("@a", msg)
		} else if score == 100 && IsOc == false {
			b.YsCoreDefines.StrongHold[k].IsOccupied = true
			for guildname, _ := range guildDic {
				b.YsCoreDefines.StrongHold[k].GuildName = guildname
				msg := b.YsCoreDefines.OccupiedStrongHoldTitle
				msg = b.FormateMsg(msg, "公会名字", guildname)
				msg = b.FormateMsg(msg, "据点名字", k)
				b.Frame.GetGameControl().SayTo("@a", msg)
				break
			}

		} else if score > 0 && IsOc == true {
			//积分满分 占领状态为true
			for guildname, _ := range guildDic {
				if b.YsCoreDefines.StrongHold[k].GuildName != guildname {
					//自减少并通报
					b.YsCoreDefines.StrongHold[k].Score--
					//fmt.Println(fmt.Sprintf("%v 分数:%v/100", k, score))
					b.TitleOfScores(k)
				}
				break
			}

		}
	}
}

// yscore相关的guild特殊执行项
func (b *Guild) GoYscore() {
	//初始化
	fmt.Println("开启yscore会员组件")
	b.InitStrongHold()

	for {
		time.Sleep(time.Second * 1)
		//开启携程
		//创立buff列表
		//map[buff]公会名字
		buffList := make(map[string]*Buff)
		go func() {
			list := <-b.GetPos("@a")
			member, guildDic := b.GetPlayerInStrongHold(list)
			//循环所有的据点

			//然后遍历所有有人据点
			//fmt.Println(guildDic)
			for k, v := range b.YsCoreDefines.StrongHold {
				//如果说据点和有人据点重合则再执行加分减分判断需求
				if player, ok := member[k]; ok {

					if b.CheckStrongHoldIsOpen(k) {
						if len(guildDic) == 1 {
							//判断并
							//fmt.Println("判定为没人")
							b.Judgment(member, guildDic)
						}
					} else {
						b.CmdSender(fmt.Sprintf("title @a[name=\"%v\"] actionbar 未在规定时间内", player))
					}

				}
				//被占领则执行buff加成(防止没人时不加buff)
				if v.IsOccupied {
					for buffname, level := range v.Buff {
						buffList = b.AddBuffInBuffList(buffname, level, buffList, v.GuildName)
					}
				}
				//将圣遗物加入名单
				for k, v := range b.GuildData {
					//查看圣遗物是否存在
					if holyrelics, ok := b.YsCoreDefines.YscoreBuff[v.HolyRelics]; ok {
						for buffName, Level := range holyrelics.Buff {
							buffList = b.AddBuffInBuffList(buffName, Level, buffList, k)
						}
					}
				}
				//判断公告
				for k, v := range b.GuildData {
					if v.Member == nil {
						b.GuildData[k].Member = make(map[string]*GuildDtails)
					}
					for memberName, j := range v.Member {
						if j.Announcement != "" {
							b.sayto(memberName, j.Announcement)
							b.GuildData[k].Member[memberName].Announcement = ""
						}
					}
				}

			}
			//发送buff
			b.SendBuff(buffList, list)
		}()

	}

}

// 据点菜单
func (b *Guild) StrongHoldMenu(name string) {
	b.sayto(name, b.YsCoreDefines.StrongHoldMenuTitle)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				for k, v := range b.YsCoreDefines.StrongHold {
					msg := b.YsCoreDefines.StrongholdDataTitle
					msg = b.FormateMsg(msg, "据点", k)
					week := ""
					for _, _week := range v.OpenTime.WeekDate {
						week = week + _week + ","
					}
					msg = b.FormateMsg(msg, "星期", week)
					msg = b.FormateMsg(msg, "开始时间", strconv.Itoa(v.OpenTime.StartTime))
					msg = b.FormateMsg(msg, "结束时间", strconv.Itoa(v.OpenTime.EndTime))
					buff := ""
					for _buff, _ := range v.Buff {
						buff = buff + " " + _buff
					}

					msg = b.FormateMsg(msg, "buff", buff)
					b.sayto(name, msg)
				}
			case "1":
				guildname, ok, _ := b.CheckInGuild(name)
				if ok {
					msg := "你所属公会 " + guildname + " 含有据点:\n"
					for k, v := range b.YsCoreDefines.StrongHold {

						if v.GuildName == guildname {
							msg = msg + k + " "
						}
						b.sayto(name, msg)
					}
				} else {
					b.sayto(name, "你没有公会")
				}
			}
		}

		return true
	})

}
func (b *Guild) getBackAlly(name string) {
	nameGuildName, ok, _ := b.CheckInGuild(name)
	if ok {
		msg := ""
		if b.GuildData[nameGuildName].AllyData == nil {
			b.GuildData[nameGuildName].AllyData = make(map[string]string)
		}
		if !(len(b.GuildData[nameGuildName].AllyData) > 0) {
			b.sayto(name, "[你没有盟友]")
		} else {
			for k, _ := range b.GuildData[nameGuildName].AllyData {
				msg = msg + fmt.Sprintf("\n%v", k)
			}
			b.sayto(name, msg)
		}

	} else {
		b.sayto(name, b.KeyTitle["无公会提示词"])
	}

}

func (b *Guild) delectAlly(name string) {
	guildname, ok, _ := b.CheckInGuild(name)
	if ok {
		b.sayto(name, "[输入对应的盟友数字]")
		msg := ""
		num := 0
		list := make(map[string]string)
		for k, _ := range b.GuildData[guildname].AllyData {
			list[strconv.Itoa(num)] = k
		}

		for k, v := range list {
			msg = msg + fmt.Sprintf("\n[%v] %v", k, v)
		}
		b.sayto(name, msg)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				if youGuildName, ok := list[chat.Msg[0]]; ok {
					delete(b.GuildData[guildname].AllyData, youGuildName)
					delete(b.GuildData[youGuildName].AllyData, guildname)
					b.sayto(name, "[删除成功]")
				}
			}
			return true
		})
	} else {
		b.sayto(name, b.KeyTitle["无工会提示词"])
	}

}

// 盟友系统
// 十分特别()
func (b *Guild) AllyMenu(name string) {
	nameGuildName, ok, _ := b.CheckInGuild(name)
	if ok {
		b.sayto(name, b.YsCoreDefines.AllyMenuTitle)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				switch chat.Msg[0] {
				case "0":
					list := b.getGuildMap()
					msg := ""
					for k, v := range list {
						msg = msg + fmt.Sprintf("\n[%v] %v\n", k, v)
					}
					b.sayto(name, msg)
					b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
						if len(chat.Msg) > 0 {
							//guidname为对方的 nameGuildName为自己的
							if guildname, ok := list[chat.Msg[0]]; ok {
								b.sayto(name, "[已发送请求] 等待对方同意")
								if b.GuildData[guildname].PendingAlly == nil {
									b.GuildData[guildname].PendingAlly = make(map[string]string)
								}
								if b.GuildData[guildname].AllyData == nil {
									b.GuildData[guildname].AllyData = make(map[string]string)
								}
								if _, ok := b.GuildData[guildname].AllyData[nameGuildName]; ok {
									b.sayto(name, "[已经有该盟友]")
									return true
								}
								if guildname == nameGuildName {
									b.sayto(name, "不能邀请自己")
									return true
								}

								b.GuildData[guildname].PendingAlly[nameGuildName] = "这是一个待处理"

							}
						}
						return true
					})
				case "1":
					b.getBackAlly(name)
				case "2":
					b.delectAlly(name)
				case "3":
					b.Agree(name)

				}
			}

			return true
		})
	} else {
		b.sayto(name, b.KeyTitle["无公会提示词"])

	}

}

// 同意
func (b *Guild) Agree(name string) {
	if guildname, ok, powernum := b.CheckInGuild(name); ok && powernum >= 3 {
		num := 0
		list := make(map[string]string)
		msg := "输入对应数字"
		for k, _ := range b.GuildData[guildname].PendingAlly {
			list[strconv.Itoa(num)] = k
			msg = msg + fmt.Sprintf("\n[%v] %v", num, k)
		}
		b.sayto(name, msg)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				if pendingGuildname, ok := list[chat.Msg[0]]; ok {
					if b.GuildData[guildname].AllyData == nil {
						b.GuildData[guildname].AllyData = make(map[string]string)
					}
					if b.GuildData[pendingGuildname].AllyData == nil {
						b.GuildData[pendingGuildname].AllyData = make(map[string]string)
					}

					b.GuildData[guildname].AllyData[pendingGuildname] = "这是一个盟友"
					b.GuildData[pendingGuildname].AllyData[guildname] = "这是一个盟友"
					delete(b.GuildData[guildname].PendingAlly, pendingGuildname)
					delete(b.GuildData[pendingGuildname].AllyData, guildname)
					b.sayto(name, "添加成功")
				} else {
					b.sayto(name, "输入正确数字")
				}
			}
			return true
		})
	} else if ok {
		b.sayto(name, b.KeyTitle["权限不足时提示"])
	} else if !ok {
		b.sayto(name, b.KeyTitle["无公会提示词"])
	}

}

// 兑换公会积分
func (b *Guild) ExchangeGuildScore(name string) {
	if guidname, ok, _ := b.CheckInGuild(name); ok {
		b.sayto(name, "输入要兑换的数额")
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {

			if len(chat.Msg) > 0 && b.CheckIsNum(chat.Msg[0]) {
				b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("scoreboard players test @a[name=\"%v\"] %v %v *", name, b.DictScore["兑换贡献计分板"], chat.Msg[0]), func(output *packet.CommandOutput) {
					if output.SuccessCount > 0 {
						b.CmdSender(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", name, b.DictScore["兑换贡献计分板"], chat.Msg[0]))
						num, _ := strconv.Atoi(chat.Msg[0])
						b.GuildData[guidname].YscoreScore += num
						b.sayto(name, fmt.Sprintf("成功兑换公会点:%v", num))
					} else {
						pterm.Info.Printfln(fmt.Sprintf("指令错误错误信息为：%v\n错误指令为:%v", output.OutputMessages, fmt.Sprintf("scoreboard players test @a[name=\"%v\"] %v %v *")))
					}
				})
			}

			return true
		})

	} else {
		b.sayto(name, b.KeyTitle["无公会提示词"])
	}
}

// 设置公会简介
func (b *Guild) SetInfoOfGuild(name string) {
	if guildname, ok, PowerNum := b.CheckInGuild(name); ok {
		if PowerNum >= 3 {
			b.sayto(name, "输入公告内容")
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				if len(chat.Msg) > 0 {
					Announcement := ""
					for _, v := range chat.Msg {
						Announcement = Announcement + v
					}
					if b.GuildData[guildname].Member == nil {
						b.GuildData[guildname].Member = make(map[string]*GuildDtails)
					}
					for _, v := range b.GuildData[guildname].Member {
						v.Announcement = Announcement
					}
					b.sayto(name, fmt.Sprintf("公告更新完成 会在公会玩家上线时提示，公会内容为:%v", Announcement))

				}
				return true
			})
		} else {
			b.sayto(name, "仅有权限为3以上时才可以设置公会公告")
		}
	} else {
		b.sayto(name, b.KeyTitle["无公会提示词"])
	}
}

// 返回圣遗物信息
func (b *Guild) GetBackHolyRelics(name string, guildname string) {
	msg := fmt.Sprintf("当前圣遗物:%v\nbuff效果:", b.GuildData[guildname].HolyRelics)
	if holyrelics, ok := b.YsCoreDefines.YscoreBuff[b.GuildData[guildname].HolyRelics]; ok {
		for k, v := range holyrelics.Buff {
			msg = msg + fmt.Sprintf("%v(%v)/", k, v)
		}
		b.sayto(name, msg)
	} else {
		b.sayto(name, "你所在公会没有圣遗物")
	}
}

// 购买公会圣遗物
func (b *Guild) HolyRelicsShop(name string, guildName string, powerNum int) {
	if powerNum >= 3 {
		Num := 0
		list := make(map[string]string)
		for k, v := range b.YsCoreDefines.YscoreBuff {
			msg := "[输入对应的号数]\n"
			list[strconv.Itoa(Num)] = k
			HolyRelicsName := k
			HolyRelicsPrice := v.Price
			msg = msg + fmt.Sprintf("[%v] %v 价格:%v buff:", Num, HolyRelicsName, HolyRelicsPrice)
			for buffname, level := range v.Buff {
				msg = msg + fmt.Sprintf("%v(%v)/", buffname, level)
			}
			msg = msg + "\n"
			Num++
			b.sayto(name, msg)
		}
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				if holyrelics, ok := list[chat.Msg[0]]; ok {
					price, _ := strconv.Atoi(b.YsCoreDefines.YscoreBuff[holyrelics].Price)
					if b.GuildData[guildName].YscoreScore >= price {
						b.GuildData[guildName].YscoreScore -= price
						b.GuildData[guildName].HolyRelics = holyrelics //设置圣遗物
						b.sayto(name, "购买成功")
					} else {
						b.sayto(name, "[余额不足]")
					}
				} else {
					b.sayto(name, "请输入有效数字")
				}
			}
			return true
		})

	} else {
		b.sayto(name, b.KeyTitle["权限不足时提示"])
	}
}

// 圣遗物菜单
func (b *Guild) HolyRelics(name string) {
	if guildname, ok, powerNum := b.CheckInGuild(name); ok {
		b.sayto(name, b.YsCoreDefines.HolyRelicsMenu)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				switch chat.Msg[0] {
				case "0":
					b.GetBackHolyRelics(chat.Name, guildname)
				case "1":
					b.HolyRelicsShop(chat.Name, guildname, powerNum)

				}
			}
			return true
		})
	} else {
		b.sayto(name, b.KeyTitle["无公会提示词"])
	}

}

// 返回公会传送点信息并决定是否删除
func (b *Guild) BackGuildTpPos(name string, guildname string) {
	if len(b.GuildData[guildname].TpPos) > 0 {
		num := 0
		list := make(map[string]string)
		msg := "输入对应号数即可删除"
		for k, v := range b.GuildData[guildname].TpPos {
			list[strconv.Itoa(num)] = k
			msg = msg + fmt.Sprintf("[%v] %v %v\n", num, k, fmt.Sprintf("%v %v %v", v[0], v[1], v[2]))
			num++
		}
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				if TpPlaceName, ok := list[chat.Msg[0]]; ok {
					b.sayto(name, "输入y则删除\n输入n则取消")
					b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
						if len(chat.Msg) > 0 {
							switch chat.Msg[0] {
							case "y":
								delete(b.GuildData[guildname].TpPos, TpPlaceName)
								b.sayto(name, "删除成功")
							case "n":
								b.sayto(name, "取消成功")
							}
						}
						return true
					})
				}
			}
			return true
		})
	} else {
		b.sayto(name, "公会暂时没有设置传送点")
	}

}

// 设置传送地点
func (b *Guild) setGuildTpPos(name string, guildname string) {
	b.sayto(name, "温馨提示：\n公会等级与可设置传送点对等")
	if b.GuildData[guildname].Power >= len(b.GuildData[guildname].TpPos) {
		b.sayto(name, "输入传送点名字")
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				TpName := chat.Msg[0]
				fmt.Println(b.GuildData[guildname].TpPos[TpName])
				if _, ok := b.GuildData[guildname].TpPos[TpName]; ok {
					b.sayto(name, "已经存在该传送点名字")
				} else {
					go func() {
						_pos := <-b.Frame.GetGameControl().GetPlayerKit(name).GetPos("@a[name=[player]]")
						pos := []int{
							_pos.X(),
							_pos.Y(),
							_pos.Z(),
						}

						if b.CheckInGuildPlace(guildname, pos) {
							if b.GuildData[guildname].TpPos == nil {
								b.GuildData[guildname].TpPos = make(map[string][]int)
							}
							b.GuildData[guildname].TpPos[TpName] = pos
							b.sayto(name, "创建成功")
						} else {
							b.sayto(name, "请站在公会领地内")
						}

					}()

				}
			}
			return true
		})
	} else {
		b.sayto(name, "你的传送点已满")
	}
}

// 设置传送菜单
func (b *Guild) SetTpPosOfYscore(name string) {
	if guildname, ok, powerNum := b.CheckInGuild(name); ok && powerNum >= 3 {
		if b.GuildData[guildname].IsTerr {
			fmt.Println("tpmenu", b.YsCoreDefines.TpMenu)
			b.sayto(name, b.YsCoreDefines.TpMenu)
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				if len(chat.Msg) > 0 {
					switch chat.Msg[0] {
					case "0":
						b.BackGuildTpPos(name, guildname)
					case "1":
						b.setGuildTpPos(name, guildname)
					}
				}

				return true
			})
		} else {
			b.sayto(name, "公会不自带地皮不能使用该功能")
		}

	} else if ok {
		b.sayto(name, b.KeyTitle["权限不足时提示"])
	} else if !ok {
		b.sayto(name, b.KeyTitle["无公会提示词"])
	}
}

// 返回公会传送点菜单
func (b *Guild) TpBackToYscore(name string) {
	if guildname, ok, _ := b.CheckInGuild(name); ok {
		if len(b.GuildData[guildname].TpPos) > 0 {
			num := 0
			list := make(map[string]string)
			msg := "输入对应号数传送"
			for k, v := range b.GuildData[guildname].TpPos {
				list[strconv.Itoa(num)] = k
				msg = msg + fmt.Sprintf("[%v] %v %v\n", num, k, fmt.Sprintf("%v %v %v", v[0], v[1], v[2]))
				num++
			}
			b.sayto(name, msg)
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				if len(chat.Msg) > 0 {
					if TpPlaceName, ok := list[chat.Msg[0]]; ok {
						pos := b.GuildData[guildname].TpPos[TpPlaceName]
						b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", name, pos[0], pos[1], pos[2]))
						b.sayto(name, "欢迎回到"+TpPlaceName)
					}
				}
				return true
			})
		} else {
			b.sayto(name, "公会暂时未设置传送点")
		}
	} else {
		b.sayto(name, b.KeyTitle["无公会提示词"])
	}
}
