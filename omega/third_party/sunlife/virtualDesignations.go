package sunlife

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"sort"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

type Designation struct {
	*defines.BasicComponent
	Triggers            []string                          `json:"触发词"`
	Usage               string                            `json:"描述"`
	Menu                map[string]string                 `json:"菜单显示"`
	TitleWord           map[string]string                 `json:"提示词显示"`
	Score               map[string]string                 `json:"所需计分板"`
	RankingDesignations map[string]*RankingDesignationMap `json:"排名称号"`
	Occupation          map[string]string                 `json:"职业称号(前分数后称号)"`
	StoreDesignation    map[string]*StoreDesignationMap   `json:"商店"`
	PassWord            string                            `json:"激活密码"`
	Data                map[string]*PlayerData
}
type PlayerData struct {
	StoreDesignations   map[string]string // 商店称号
	RankingDesignations map[string]string // 排名称号
	StoreDesignation    string            //商店称号装备
	RankeingDesignation string            //排名称号装备
	Occupation          string            // 职业称号
	Score               map[string]int    //玩家计分板分数
}
type StoreDesignationMap struct {
	Score int `json:"对应分数"`
	Price int `json:"价格"`
}
type RankingDesignationMap struct {
	One   string `json:"第一名称号"`
	Tow   string `json:"第二名称号"`
	Three string `json:"第三名称号"`
}

func (b *Designation) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	if b.Data == nil {
		b.Data = make(map[string]*PlayerData)
	}
	if cfg.Version == "0.0.1" {
		delete(cfg.Configs, "激活密码")
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	//b.GuildData = make(map[string]*GuildDatas)
}
func (b *Designation) InitData() {
	go func() {
		list := <-GetPlayerName(b.Frame, "@a")
		for _, v := range list {
			if _, ok := b.Data[v]; !ok {
				b.Data[v] = &PlayerData{
					StoreDesignations:   map[string]string{},
					RankingDesignations: map[string]string{},
					Occupation:          "",
					RankeingDesignation: "",
					StoreDesignation:    "",
					Score:               map[string]int{},
				}
			}
		}
	}()

}
func (b *Designation) Inject(frame defines.MainFrame) {
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
	b.Frame.GetJsonData("称号系统.json", &b.Data)
	b.InitData()
	b.Listener.AppendLoginInfoCallback(b.onLogin)
	b.Listener.AppendLoginInfoCallback(b.onLoginOfDes)

}

// 保存数据
func (b *Designation) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return b.Frame.WriteJsonDataWithTMP("称号系统.json", ".ckpt", &b.Data)
	}
	return nil
}

// 登录时全服提示
func (b *Designation) onLoginOfDes(entry protocol.PlayerListEntry) {
	if data, ok := b.Data[entry.Username]; !ok {
		b.Data[entry.Username] = &PlayerData{
			StoreDesignations:   map[string]string{},
			RankingDesignations: map[string]string{},
			Occupation:          "",
			RankeingDesignation: "",
			StoreDesignation:    "",
		}
	} else {
		if _, isok := data.StoreDesignations[data.StoreDesignation]; !isok && data.StoreDesignation != "" {
			b.Data[entry.Username].StoreDesignation = ""
			Sayto(b.Frame, entry.Username, "[错误并未找到你的商店称号在你的商店称号仓库中]\n已自动删除")
		}
		if _, isok := data.RankingDesignations[data.RankeingDesignation]; !isok && data.RankeingDesignation != "" {
			b.Data[entry.Username].RankeingDesignation = ""
			Sayto(b.Frame, entry.Username, "[错误并未找到你的排名称号在你的排名称号仓库中]\n已自动删除")
		}

		//有任意一个称号都会触发
		if data.StoreDesignation != "" || data.RankeingDesignation != "" || data.Occupation != "" {
			msg := b.TitleWord["进服提示"]
			relist := map[string]interface{}{
				"职业":     data.Occupation,
				"商店称号":   data.StoreDesignation,
				"排名称号":   data.RankeingDesignation,
				"player": entry.Username,
			}
			msg = FormateMsg(b.Frame, relist, msg)
			b.Frame.GetGameControl().SayTo("@a", msg)
		}
	}

}

// 初始化分数
func (b *Designation) onLogin(entry protocol.PlayerListEntry) {
	if _, ok := b.Data[entry.Username]; !ok {
		pterm.Info.Println("初始化成功", entry.Username)
		b.Data[entry.Username] = &PlayerData{
			StoreDesignations:   map[string]string{},
			RankingDesignations: map[string]string{},
			Occupation:          "",
			RankeingDesignation: "",
			StoreDesignation:    "",
			Score:               make(map[string]int),
		}
	}
}

// 命令发送者
func (b *Designation) CmdSender(str string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(str, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {

		} else {
			pterm.Info.Printfln("指令错误 错误信息为:%v\n错误指令为:%v", output.OutputMessages, str)
		}
	})

}

// 是否符合
func (b *Designation) CheckNumOfS(num int) (isok bool, TheName string) {
	for k, v := range b.StoreDesignation {
		if v.Score == num {
			return true, k
		}
	}
	return false, ""
}
func (b *Designation) Activate() {
	//初始化计分板
	go func() {
		for {
			time.Sleep(time.Second * 5)
			AllScore := <-GetScore(b.Frame)

			for k, Score := range AllScore {
				//检查商店称号绑定
				if ok, TheName := b.CheckNumOfS(Score[b.Score["商店称号计分板"]]); !ok {
					b.Data[k].StoreDesignation = ""
				} else {
					b.Data[k].StoreDesignation = TheName
				}
				//先是职业称号
				for num, OccupationName := range b.Occupation {
					scoreNum, _ := strconv.Atoi(num)
					//如果分数与盗贼分数一致则同步
					//fmt.Println(Score)
					if scoreNum == Score[b.Score["职业称号计分板"]] {
						b.Data[k].Occupation = OccupationName
					}
				}
				//然后是排名 先更新已有分数
				b.Data[k].Score = Score
			}
			//去除一次称号
			for _, v := range b.Data {
				v.RankingDesignations = map[string]string{}
			}
			//遍历排名称号好给称号 找寻每个计分板的第1、2、3名
			for scoreName, rankName := range b.RankingDesignations {
				scoreList := []int{}
				//遍历所有人讲分数存入scoreList
				for _, j := range b.Data {
					if num, ok := j.Score[scoreName]; ok {

						scoreList = append(scoreList, num)
					}
				}
				sort.Sort(sort.Reverse(sort.IntSlice(scoreList)))
				//pterm.Info.Printfln("%v", scoreList)
				if len(scoreList) >= 1 {
					FirstNum := scoreList[0]
					for k, v := range b.Data {
						//如果是第一名则给仓库添加称号
						if v.Score[scoreName] == FirstNum {
							//pterm.Info.Printfln("第一名为:%v", k)
							b.Data[k].RankingDesignations[rankName.One] = "这是一个排名称号"
						}
					}
				}
				if len(scoreList) >= 2 {
					SecondNum := scoreList[1]
					for k, v := range b.Data {
						//如果是第一名则给仓库添加称号
						if v.Score[scoreName] == SecondNum {
							//pterm.Info.Printfln("第二名为:%v", k)
							b.Data[k].RankingDesignations[rankName.Tow] = "这是一个排名称号"
						}
					}
				}
				if len(scoreList) >= 3 {
					ThreeNum := scoreList[2]
					for k, v := range b.Data {
						//如果是第一名则给仓库添加称号
						if v.Score[scoreName] == ThreeNum {
							//pterm.Info.Printfln("第三名为:%v", k)
							b.Data[k].RankingDesignations[rankName.Three] = "这是一个排名称号"
						}
					}
				}

			}

		}
	}()
	for k, v := range b.Score {
		b.CmdSender(fmt.Sprintf("scoreboard objectives add %v dummy %v", v, k))
	}

}

// 购买商品
func (b *Designation) BuyCommodity(name string, commodity string) {
	cmd := fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", name, b.Score["货币计分板"], b.StoreDesignation[commodity].Price)
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			b.Data[name].StoreDesignations[commodity] = "这是一个商店称号"

			Sayto(b.Frame, name, b.TitleWord["购买成功提示词"])
		} else {
			pterm.Info.Printfln("错误执行 错误信息为:%v\n错误的指令为:%v", output.OutputMessages, cmd)
		}
	})
}

// 商店菜单
func (b *Designation) Store(name string) {
	num := 0
	list := make(map[string]string, len(b.StoreDesignation))
	for k, v := range b.StoreDesignation {
		msg := b.TitleWord["商店模板"]
		ReList := map[string]interface{}{
			"i":  num,
			"称号": k,
			"价格": v.Price,
		}
		list[strconv.Itoa(num)] = k
		msg = FormateMsg(b.Frame, ReList, msg)
		Sayto(b.Frame, name, msg)
		num++
	}
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if commodity, ok := list[chat.Msg[0]]; ok {
				b.BuyCommodity(name, commodity)
			} else {
				Sayto(b.Frame, name, b.TitleWord["无效输入提示"])
			}
		}

		return true
	})
}

// 返回商店称号
func (b *Designation) GetBackStoreDes(name string) {
	num := 0
	list := make(map[string]string)
	for k, _ := range b.Data[name].StoreDesignations {
		msg := b.TitleWord["商店称号仓库模板"]
		list[strconv.Itoa(num)] = k
		relist := map[string]interface{}{
			"称号": k,
			"i":  num,
		}
		msg = FormateMsg(b.Frame, relist, msg) + "\n"
		Sayto(b.Frame, name, msg)
		num++
	}
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if StoreName, ok := list[chat.Msg[0]]; ok {
				b.Data[name].StoreDesignation = StoreName
				b.CmdSender(fmt.Sprintf("scoreboard players set @a[name=\"%v\"] %v %v", name, b.Score["商店称号计分板"], b.StoreDesignation[StoreName].Score))
				Sayto(b.Frame, name, b.TitleWord["装备成功称号提示词"])
			} else {
				Sayto(b.Frame, name, "输入有效数字")
			}
		}

		return true
	})
}
func (b *Designation) GetBackRankDes(name string) {
	num := 0
	list := make(map[string]string)
	for k, _ := range b.Data[name].RankingDesignations {
		msg := b.TitleWord["排名称号仓库模板"]
		list[strconv.Itoa(num)] = k
		relist := map[string]interface{}{
			"称号": k,
			"i":  num,
		}
		msg = FormateMsg(b.Frame, relist, msg) + "\n"
		Sayto(b.Frame, name, msg)
		num++
	}
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if StoreName, ok := list[chat.Msg[0]]; ok {
				b.Data[name].RankeingDesignation = StoreName
				Sayto(b.Frame, name, b.TitleWord["装备成功称号提示词"])
			} else {
				Sayto(b.Frame, name, "输入有效数字")
			}
		}

		return true
	})
}

// 仓库
func (b *Designation) Warehouse(name string) {
	relist := map[string]interface{}{
		"职业": b.Data[name].Occupation,
	}
	Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["仓库商店"]))
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.GetBackStoreDes(name)
			case "1":
				b.GetBackRankDes(name)

			}
		}

		return true
	})

}

// 返回数据
func (b *Designation) GetBackData(name string) {
	list := map[string]interface{}{
		"player": name,
		"职业称号":   b.Data[name].Occupation,
		"排名称号":   b.Data[name].RankeingDesignation,
		"商店称号":   b.Data[name].StoreDesignation,
	}
	Sayto(b.Frame, name, FormateMsg(b.Frame, list, b.TitleWord["个人信息展示"]))
}
func (b *Designation) Center(chat *defines.GameChat) bool {
	Sayto(b.Frame, chat.Name, b.Menu["主菜单显示"])

	b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.Store(chat.Name)
			case "1":
				b.Warehouse(chat.Name)
			case "2":
				b.GetBackData(chat.Name)
			}
		}

		return true
	})

	return true
}
