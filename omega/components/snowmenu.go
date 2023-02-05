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

type SnowMenu struct {
	*defines.BasicComponent
	Snowscore              string              `json:"雪球菜单计分板"`
	SnowsMenuTitle         map[string]string   `json:"雪球菜单菜单显示以及对应积分"`
	SnowsMenuActive        map[string][]string `json:"雪球菜单对应积分执行指令"`
	SnowMenuScoreCirculate map[string]string   `json:"当分数等于前者时分数时自动跳转后者"`
	SnowMenuTitleTarget    string              `json:"雪球菜单显示对应选择器"`
	SnowMenuActiveTarget   string              `json:"雪球菜单触发时的选择器"`
	TimeDelay              int                 `json:"系统检测周期(毫秒)"`
}

func (b *SnowMenu) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, b)
	if err != nil {
		panic(err)
	}
	//b. = make(map[string]*GuildDatas)

}
func (b *SnowMenu) Inject(frame defines.MainFrame) {
	b.Frame = frame

	b.BasicComponent.Inject(frame)
	//fmt.Println("-------", b.SnowsMenuTitle)
}
func (b *SnowMenu) Activate() {
	fmt.Println("[提示] 当前周期为:", b.TimeDelay)
	b.Frame.GetGameControl().SendCmd("scoreboard objectives add " + b.Snowscore + " dummy")

	go func() {
		for {
			//fmt.Println("timedelay :", b.TimeDelay)
			time.Sleep(time.Millisecond * time.Duration(b.TimeDelay)) //time.Duration(b.timeDelay))
			//雪球功能实现
			// fmt.Println("---------------------------\ntest")

			go func() {
				list := <-b.GetScore("@a")
				b.MenuTitle(list)
			}()
			go func() {
				rxlist := <-b.GetScore("@a[rx=-88]")
				b.snowActive(rxlist)
			}()

			//list := <-

			//rxlist := <-b.GetScore()

		}
	}()

}

// 雪球实现部分
func (b *SnowMenu) snowActive(rxlist map[string]map[string]int) {
	fmt.Println("snowactive")
	for k, v := range rxlist {
		if cmd, ok := b.SnowsMenuActive[strconv.Itoa(v[b.Snowscore])]; ok {
			for i, j := range cmd {
				j = b.FormateMsg(j, "雪球菜单计分板", b.Snowscore)
				b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(b.FormateMsg(j, "执行对象名字", k), func(output *packet.CommandOutput) {
					if output.SuccessCount > 0 {
					} else {
						fmt.Printf("[错误] 雪球优化菜单组件功能实现分数:%v第%v出现错误 报错信息如下%v\n", strconv.Itoa(v[b.Snowscore]), strconv.Itoa(i), output.DataSet)
					}
				})

			}
		}
	}
}

// 雪球显示部分
func (b *SnowMenu) MenuTitle(list map[string]map[string]int) {
	//fmt.Print("menutitle")
	//fmt.Println("list", list)
	for k, v := range list {
		//如果分数达到对应则显示对应分数
		//fmt.Println("k:", k, "\nv:", v)
		//fmt.Println("snowsscores", v[b.Snowscore])
		if t, ok := b.SnowsMenuTitle[strconv.Itoa(v[b.Snowscore])]; ok {
			//fmt.Println("ok")
			b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("title @a[name=\"%v\"] actionbar %v", k, t), func(output *packet.CommandOutput) {
				fmt.Println(output.OutputMessages)
			})
		}
	}
}

/*
// 雪球加分

	func (b *SnowMenu) addScore(target string) {
		for {
			time.Sleep(time.Millisecond * 500)
			//实现雪球加分
			b.Frame.GetGameControl().SendCmd("scoreboard players rest @a[rxm=88] " + b.Snowscore + " ")

			b.Frame.GetGameControl().SendCmdAndInvokeOnResponse("execute @e[type=snowball] ~~~ give @p[r=5] snowball", func(output *packet.CommandOutput) {
				if output.SuccessCount > 0 {
					b.Circulate()
					b.Frame.GetGameControl().SendCmd("execute @e[type=snowball] ~~~ scoreboard players add @p " + b.Snowscore + " 1")
					b.Frame.GetGameControl().SendCmd("kill @e[type=snowball]")

				}
			})
			//让分数循环

		}
	}
*/
func (b *SnowMenu) sayto(name string, str string) {
	b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), str)
}

// 获取所有人的积分 返回通道
func (b *SnowMenu) GetScore(target string) (PlayerScoreList chan map[string]map[string]int) {

	cmd := "scoreboard players list " + target
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

// 获取指定限制器的玩家名字 返回通道值 key 为玩家名字 v为号数()
func (b *SnowMenu) GetPlayerName(name string) (listChan chan map[string]string) {
	type User struct {
		Name []string `json:"victim"`
	}
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
				mapName[v] = strconv.Itoa(k)
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

// 格式化信息
func (b *SnowMenu) FormateMsg(str string, re string, afterstr string) (newstr string) {

	res := regexp.MustCompile("\\[" + re + "\\]")
	return res.ReplaceAllString(str, afterstr)

}
