package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"regexp"
	"strconv"
	"time"
)

type TerritoryTest struct {
	*defines.BasicComponent
	ProtectCmds      []string          `json:"地皮保护时执行的指令"`
	KeyWordMenu      map[string]string `json:"菜单显示项目"`
	KeyWord          map[string]string `json:"各种提示词"`
	IsCustom         bool              `json:"是否开启自定义地皮大小"`
	PriceOfoneblock  int               `json:"每格价格"`
	PriceOfAllblock  int               `json:"非自定义时单个地皮价格"`
	RangeOfTerritory []int             `json:"非自定义时地皮的范围"`
	Score            string            `json:"扣除货币的积分板"`
	Triggers         []string          `json:"触发词"`
	Usage            string            `json:"提示信息"`
	DelayTime        int64             `json:"执行地皮保护指令延迟(毫秒)"`
	BuyCmds          []string          `json:"购买地皮时执行的指令"`
	fileChange       bool
	Data             map[string]*TerritoryData
}

type TerritoryData struct {
	Pos    []int
	Range  []int
	Member []string
}

func (b *TerritoryTest) Init(cfg *defines.ComponentConfig) {
	//get config of setting
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	b.Data = make(map[string]*TerritoryData)
	//检查信息是否完整 如果不完整则自动填充默认的

	if _, isok := b.KeyWord["地皮主人选择器"]; isok {

	} else {
		b.KeyWord["地皮主人选择器"] = "@a[name=\"[地皮主人]\"]"
	}
	if _, isok := b.KeyWord["无地皮提示词"]; isok {
	} else {
		b.KeyWord["无地皮提示词"] = "[无地皮对象] 你没有地皮 请购买后再尝试"
	}
	if _, isok := b.KeyWord["购买地皮时检测指令"]; isok {
	} else {
		b.KeyWord["购买地皮时检测指令"] = "testfor @a[name=\"[购买地皮的人]\",scores={[货币计分板]=[地皮价格]..}]"
	}

}
func (b *TerritoryTest) Inject(frame defines.MainFrame) {
	//
	b.Frame = frame
	//注入frame等东西
	b.Frame.GetJsonData("地皮信息.json", &b.Data)
	b.BasicComponent.Inject(frame)
	//b.Listener.SetGameChatInterceptor(b.ProcessingCenter)
	//获取信息
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
func (b *TerritoryTest) Activate() {
	b.Frame.GetGameControl().SayTo("@a", "地皮组件已开启")
	go b.protect()
}

func (o *TerritoryTest) TerritoryTest(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP("地皮信息.json", ".ckpt", o.Data)
		}
	}
	return nil
}

func (b *TerritoryTest) Stop() error {
	fmt.Print("开始保存地皮插件信息")
	return b.Frame.WriteJsonDataWithTMP("地皮信息.json", ".final", b.Data)
}

func (b *TerritoryTest) Center(chat *defines.GameChat) bool {

	msg := "§b§l[输入对应数字]\n1:§e购买地皮\n§b2:§e返回地皮\n§b3:§e授予白名单\n§b4:§e删除白名单\n§b5:§e删除自己地皮"

	b.Frame.GetGameControl().SayTo("@a[name=\""+chat.Name+"\"]", msg)

	if b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(NewChat *defines.GameChat) (catch bool) {
		if NewChat.Msg[0] == "1" {
			//如果为1则是购买地皮
			b.BuyTerritorysCenter(NewChat.Name)
		} else if NewChat.Msg[0] == "2" {
			//如果是2则是返回地皮
			b.TpBackTerritorys(NewChat.Name)
		} else if NewChat.Msg[0] == "3" {
			//三是授予白名单
			b.giveMember(NewChat.Name)
		} else if NewChat.Msg[0] == "4" {
			//删除白名单
			b.delectMember(NewChat.Name)
		} else if NewChat.Msg[0] == "5" {
			//删除地皮
			b.delectTerr(NewChat.Name)
		}
		return true
	}) == nil {

	}
	return true

}
func (b *TerritoryTest) delectTerr(name string) {
	if _, ok := b.Data[name]; ok {
		msg := b.KeyWord["删除地皮成功提示"]
		delete(b.Data, name)
		b.fileChange = true
		b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", msg)
	} else {
		b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "§e[错误]§b你没有地皮")
	}
}
func (b *TerritoryTest) delectMember(name string) {
	if _, ok := b.Data[name]; ok {
		msg := "§e§c[白名单]\n"
		if len(b.Data[name].Member) >= 1 {
			for k, v := range b.Data[name].Member {
				msg = msg + "§b" + strconv.Itoa(k) + ":§e" + v + "\n"
			}
			msg = msg + "§e§l输入对应数字删除对应人"
			b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", msg)
			if b.Frame.GetGameControl().SetOnParamMsg(name, func(NewChat *defines.GameChat) (catch bool) {
				num, err := strconv.Atoi(NewChat.Msg[0])
				if err != nil {
					b.Frame.GetGameControl().SayTo("@a[name=\""+NewChat.Name+"\"]", "请输入有效数字")
				}
				if len(b.Data[NewChat.Name].Member) >= num {

					msgs := b.KeyWord["删除白名单成功提示"]
					msgs = b.FormateMsg(msgs, "删除对象", b.Data[NewChat.Name].Member[num])
					b.Data[NewChat.Name].Member = append(b.Data[NewChat.Name].Member[:num], b.Data[NewChat.Name].Member[num+1:]...)
					b.fileChange = true
					b.Frame.GetGameControl().SayTo("@a[name=\""+NewChat.Name+"\"]", msgs)
				} else {
					b.Frame.GetGameControl().SayTo("@a[name=\""+NewChat.Name+"\"]", "【警告】输入的数字超过了最大")
				}
				return true
			}) == nil {

			}

		} else {
			b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "§e没有白名单人员")
		}

	} else {
		b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "§e您没有地皮 请购买后再尝试")
	}

}
func (b *TerritoryTest) giveMember(name string) {
	if _, ok := b.Data[name]; ok {
		b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "§b§l请输入你要授权人的名字")
		if b.Frame.GetGameControl().SetOnParamMsg(name, func(NewChat *defines.GameChat) (catch bool) {
			b.Data[NewChat.Name].Member = append(b.Data[NewChat.Name].Member, NewChat.Msg[0])
			b.fileChange = true
			b.Frame.GetGameControl().SayTo("@a[name=\""+NewChat.Name+"\"]", "§a§l已对对方授权\n授权名字:"+NewChat.Msg[0])
			return true
		}) == nil {

		}
	} else {
		b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "§e您没有地皮 请购买后再尝试")
	}

}
func (b *TerritoryTest) protect() {
	for {
		time.Sleep(time.Duration(b.DelayTime) * time.Millisecond)
		if len(b.Data) == 0 {
			time.Sleep(time.Second * 3)
		}
		if len(b.Data) >= 1 {
			for _k, _v := range b.Data {
				k, v := _k, _v
				//延迟设置
				time.Sleep(time.Duration(b.DelayTime) * time.Millisecond)
				cmd := "testfor @a[[地皮范围]]"
				strPos := "x=" + strconv.Itoa(v.Pos[0]) + ",z=" + strconv.Itoa(v.Pos[2]) + ",y=-60" + ",dx=" + strconv.Itoa(v.Range[0]) + ",dy=400" + ",dz=" + strconv.Itoa(v.Range[1])
				cmd = b.FormateMsg(cmd, "地皮范围", strPos)
				b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
					for _, i := range b.ProtectCmds {
						MasterName := "name=!\"" + k + "\""
						if len(v.Member) >= 1 {
							for _, j := range v.Member {
								MasterName = MasterName + "," + "name=!\"" + j + "\""
							}
						}
						i = b.FormateMsg(i, "非地皮合法进入者", MasterName)

						//fmt.Print(strPos, "\n")
						i = b.FormateMsg(i, "地皮范围", strPos)
						i = b.FormateMsg(i, "地皮主人", k)
						i = b.FormateMsg(i, "地皮起点x", strconv.Itoa(v.Pos[0]))
						i = b.FormateMsg(i, "地皮起点z", strconv.Itoa(v.Pos[2]))
						i = b.FormateMsg(i, "地皮起点y", strconv.Itoa(v.Pos[1]))
						i = b.FormateMsg(i, "地皮dx", strconv.Itoa(v.Range[0]))
						i = b.FormateMsg(i, "地皮dz", strconv.Itoa(v.Range[1]))
						///fmt.Print(i, "\n")
						b.Frame.GetGameControl().SendCmd(i)

						//b.Frame.GetGameControl().SendCmd(i)
					}
				})

			}
		}
	}

}

func (b *TerritoryTest) TpBackTerritorys(name string) {
	//返回地皮
	if _, ok := b.Data[name]; ok {
		msg := "tp " + b.FormateMsg(b.KeyWord["地皮主人选择器"], "地皮主人", name) + " " + strconv.Itoa(b.Data[name].Pos[0]) + " " + strconv.Itoa(b.Data[name].Pos[1]) + " " + strconv.Itoa(b.Data[name].Pos[2])

		b.Frame.GetGameControl().SendCmd(msg)
	} else {
		msg := b.FormateMsg(b.KeyWord["无地皮提示词"], "无地皮对象", name)
		b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", msg)
	}

}

//写入用户信息
func (b *TerritoryTest) WriteJsonDatas(name string, Range []int, price string) (suss bool) {

	Pos := <-b.Frame.GetGameControl().GetPlayerKit(name).GetPos("@a[name=[player]]")
	if len(b.Data) >= 1 {
		for _, v := range b.Data {
			if b.CheckIsoverlap(Pos, Range, v.Pos, v.Range) {
				b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "[购买失败] 附近有其他地皮")
				return false
			}
		}
	}
	b.Data[name] = &TerritoryData{
		Pos:    Pos,
		Range:  Range,
		Member: make([]string, 1),
	}
	b.fileChange = true
	//fmt.Print(b.Data[name], "\n")
	fmt.Print(b.BuyCmds)
	for _, v := range b.BuyCmds {

		newMsg := b.FormateMsg(v, "购买对象名字", name)
		newMsg = b.FormateMsg(newMsg, "货币计分板", b.Score)
		newMsg = b.FormateMsg(newMsg, "价格", price)
		fmt.Print(newMsg, "--251--\n")
		b.Frame.GetGameControl().SendCmd(newMsg)

	}

	b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "购买成功 ")

	return true
}

func (b *TerritoryTest) BuyTerritory(name string, price int, Range []int) {

	cmd := b.FormateMsg(b.KeyWord["购买地皮时检测指令"], "购买地皮的人", name)
	cmd = b.FormateMsg(cmd, "货币计分板", b.Score)
	cmd = b.FormateMsg(cmd, "地皮价格", strconv.Itoa(price))

	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		//fmt.Print("--发送命令--", "\n")
		if output.SuccessCount > 0 {

			go b.WriteJsonDatas(name, Range, strconv.Itoa(price))

		} else {
			msg := b.FormateMsg(b.KeyWord["购买地皮指令检测失败后提示"], "购买地皮的人", name)
			msg = b.FormateMsg(msg, "货币计分板", b.Score)
			b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", msg)
		}
	})
}
func (b *TerritoryTest) BuyTerritorysCenter(name string) {
	if b.IsCustom {
		//开启自定义
		msg := b.FormateMsg(b.KeyWord["自定义购买时提示输入地皮大小"], "购买对象", name)
		msg = b.FormateMsg(msg, "每格价格", strconv.Itoa(b.PriceOfoneblock))
		fmt.Print(msg, "自定义购买时提示话语格式化\n")
		b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", msg)
		if b.Frame.GetGameControl().SetOnParamMsg(name, func(NewChat *defines.GameChat) (catch bool) {

			x := NewChat.Msg[0]
			y := NewChat.Msg[1]
			okx, _ := regexp.MatchString("^\\+?[1-9][0-9]*$", x)
			okz, _ := regexp.MatchString("^\\+?[1-9][0-9]*$", y)
			if okx && okz {
				//Range :=[2]int{strconv.Atoi(x),strconv.Atoi(y)}
				msg = b.FormateMsg(b.KeyWord["提示确认地皮大小提示词"], "地皮大小", x+","+y)
				xint, err := strconv.Atoi(x)
				if err != nil {
					fmt.Errorf(err.Error())
				}
				yint, err := strconv.Atoi(y)
				if err != nil {
					fmt.Errorf(err.Error())
				}
				price := xint * b.PriceOfoneblock * yint

				msg = b.FormateMsg(msg, "价格", strconv.Itoa(price))
				msg = b.FormateMsg(msg, "购买地皮计分板", b.Score)
				fmt.Print(msg, "--324--\n")
				b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", msg)
				if b.Frame.GetGameControl().SetOnParamMsg(name, func(NewChats *defines.GameChat) (catch bool) {
					if NewChats.Msg[0] == "y" {
						var list []int

						list = append(list, xint)
						list = append(list, yint)
						//print("--247--", list, "\n")
						b.BuyTerritory(NewChats.Name, price, list)
					} else {
						fmt.Print(NewChats, "购买失败\n")
						return false
					}
					return true
				}) == nil {
					fmt.Print("有人挤占")
				}
			} else {
				b.Frame.GetGameControl().SayTo("@a[name=\""+name+"\"]", "请输入有效数字")
			}

			return true
		}) == nil {

		}

	} else {
		TestName := name
		b.BuyTerritory(TestName, b.PriceOfAllblock, b.RangeOfTerritory)
	}
}
func (b *TerritoryTest) CheckIsoverlap(pos []int, Epos []int, Spos []int, SEpos []int) (IsOverlap bool) {
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

//检查某个东西是否在数组里面
func (b *TerritoryTest) CheckArr(arr []string, str string) (IsIn bool) {
	if len(arr) == 0 {
		fmt.Print("数组为空")
		return false
	} else {
		var set map[string]struct{}
		set = make(map[string]struct{})
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
func (b *TerritoryTest) FormateMsg(str string, re string, afterstr string) (newstr string) {

	res := regexp.MustCompile("\\[" + re + "\\]")
	return res.ReplaceAllString(str, afterstr)

}
