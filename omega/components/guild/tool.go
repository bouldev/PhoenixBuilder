package guild

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

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

// 检查字典的值是否有吻合这个的
func (b *Guild) FindmapKey(list map[string]string, key string) bool {
	for _, v := range list {
		if v == key {
			return true
		}
	}
	return false
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

// 正则表达检查字符串是否为数字
func (b *Guild) CheckIsNum(str string) bool {
	ok, _ := regexp.MatchString("^\\+?[1-9][0-9]*$", str)
	return ok
}

// 检查地皮是否重合
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

// 格式化输出
func (b *Guild) FromatePrint(msg string, T string) {
	fmt.Print(fmt.Sprintf("----------------\n[%v]  :   %v\n-------------------\n", T, msg))
}
func (b *Guild) QuitTheGuild(name string) {
	guildname, ok, PowerNum := b.CheckInGuild(name)
	if ok && PowerNum < 4 {
		delete(b.GuildData[guildname].Member, name)
		b.sayto(name, b.KeyTitle["退出公会成功提示词"])

	} else if ok && PowerNum == 4 {
		b.sayto(name, b.KeyTitle["会长退出公会提示词"])
	} else if ok == false {
		b.sayto(name, b.KeyTitle["无公会提示词"])
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

// 检查结构体是否含有某个元素
func (b *Guild) IsHash(ex interface{}, name string) bool {
	t := reflect.TypeOf(ex)
	//修改开始
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if _, ok := t.FieldByName(name); ok {
		return true
	}
	return false
}

// 让指定的name返回领地
func (b *Guild) TpBack(name string) bool {
	guildname, ok, _ := b.CheckInGuild(name)
	if ok && len(b.GuildData[guildname].CenterPos) > 0 {
		msg := b.FormateMsg(b.KeyTitle["返回地皮时提示词"], "公会名字", guildname)
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
		fmt.Print("centerpos", b.GuildData[guildname].CenterPos)
		fmt.Print(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", name, b.GuildData[guildname].CenterPos[0], b.GuildData[guildname].CenterPos[1], b.GuildData[guildname].CenterPos[2]))
		b.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", name, b.GuildData[guildname].CenterPos[0], b.GuildData[guildname].CenterPos[1], b.GuildData[guildname].CenterPos[2]))
		return true

	} else if ok == false {
		b.sayto(name, b.KeyTitle["无地皮时提示"])
	}

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

// 获取个人信息
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

// 设置特殊区域
func (b *Guild) setSpPlace(name string, guildName string) {

	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		go func() {
			_starPosSp := <-b.Frame.GetGameControl().GetPlayerKit(name).GetPos("@a[name=[player]]")
			starPosSp := []int{}
			if _starPosSp != nil {
				starPosSp = []int{_starPosSp.X(), _starPosSp.Y(), _starPosSp.Z()}
			}
			fmt.Println("starPosSp:", starPosSp)
			if b.CheckInGuildPlace(guildName, starPosSp) {
				b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("[确定坐标] %v %v %v \n[终点坐标确定] 在终点输入任意字", strconv.Itoa(starPosSp[0]), strconv.Itoa(starPosSp[1]), strconv.Itoa(starPosSp[2])))
				b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
					fmt.Print(Newchat.Name, "\n")
					go func() {
						_EndPos := <-b.Frame.GetGameControl().GetPlayerKit(Newchat.Name).GetPos("@a[name=[player]]")
						EndPos := []int{}
						if _EndPos != nil {
							EndPos = []int{_EndPos.X(), _EndPos.Y(), _EndPos.Z()}
						}
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

// 踢出成员
func (b *Guild) kickMember(guildName string, source string) {
	if b.IsAllowKick {
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
	} else {
		b.sayto(source, "不允许使用kick")
	}

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

// 获取指定限制器的玩家名字 返回通道值 key 为数字 v为玩家
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
	if data.IsTerr {
		msg = b.FormateMsg(msg, "中心坐标", fmt.Sprintf("%v %v %v", strconv.Itoa(data.CenterPos[0]), strconv.Itoa(data.CenterPos[1]), strconv.Itoa(data.CenterPos[2])))
		msg = b.FormateMsg(msg, "起始坐标", fmt.Sprintf("%v %v %v", strconv.Itoa(data.Pos[0]), "-65", strconv.Itoa(data.Pos[2])))
		Range := b.GuildRange[strconv.Itoa(data.Power)]
		msg = b.FormateMsg(msg, "终点坐标", fmt.Sprintf("%v %v %v", strconv.Itoa(data.Pos[0]+Range[0]), "400", strconv.Itoa(data.Pos[2]+Range[1])))

	}
	b.sayto(name, msg)
}

// 写入公会信息不带领地
func (b *Guild) RegisteredWithoutTerr(name string, guildName string) {
	b.GuildData[guildName] = &GuildDatas{
		Master: name,
		Member: make(map[string]*GuildDtails),
		Power:  b.GuildFristPower,
		IsTerr: false,
	}
	fmt.Println("[提示] ", name, " ", "创建成功", b.GuildData[guildName])
	b.sayto(name, "[购买成功]")
	//b.sayto(name,)
}

// 写入公会信息
func (b *Guild) Registered(name string, guildName string) {
	go func() bool {
		//获取坐标
		_PosOf := <-b.Frame.GetGameControl().GetPlayerKit(name).GetPos("@a[name=[player]]")
		PosOf := []int{}
		if _PosOf != nil {
			PosOf = []int{_PosOf.X(), _PosOf.Y(), _PosOf.Z()}
		}

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
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["退出公会成功提示词"])
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
						IsTerr:    true,
					}

					b.Frame.GetGameControl().SendCmd(fmt.Sprintf("scoreboard players set @a[name=\"%v\"] %v 4", name, b.DictScore["权限计分板"]))
					b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "§b§l[购买成功] ")
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
}

// 写入公会信息
func (b *Guild) WriteGuildData(name string, guildName string) {
	k, v := b.CheckIsMaster(name)
	if k == false {
		if b.IsNeedTerr {
			fmt.Println("[判断] ", name, "为需要带地皮")
			b.Registered(name, guildName)
		} else {
			b.RegisteredWithoutTerr(name, guildName)
		}

	} else if k == true {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§a[购买失败] %v你已经是§b%v§a的会长了", name, v))
	} else {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "意料之外的情况 稍后重试")
	}
}
func (b *Guild) setGuildPower(guildname string, num string) {

}

// 发送菜单
func (b *Guild) sendMenu(name, str string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), str)
}
