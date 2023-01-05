package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/items"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"

	"github.com/pterm/pterm"
)

type RegexCheck struct {
	Enabled               bool   `json:"启用"`
	Description           string `json:"检测说明"`
	Debug                 bool   `json:"调试模式"`
	Item                  string `json:"使用正则表达式匹配物品名"`
	Tag                   string `json:"匹配标签名"`
	RegexString           string `json:"使用正则表达式匹配标签值"`
	Allow                 bool   `json:"匹配标签值成功时true为放行false为作弊"`
	compiledItemNameRegex regexp.Regexp
	compiledValueRegex    regexp.Regexp
}

type IntrusionDetectSystem struct {
	*defines.BasicComponent
	EnableK32Detect bool `json:"启用32k手持物品检测"`
	K32Threshold    int  `json:"32k手持物品附魔等级阈值"`
	k32Response     []defines.Cmd
	K32ResponseIn   interface{}   `json:"32k手持物品反制"`
	RegexCheckers   []*RegexCheck `json:"使用以下正则表达式检查"`
}

func findK(key string, val interface{}, onKey func(interface{})) {
	switch value := val.(type) {
	case map[string]interface{}:
		for k, v := range value {
			if k == key {
				onKey(v)
			} else {
				findK(key, v, onKey)
			}
		}
	case []interface{}:
		for _, v := range value {
			findK(key, v, onKey)
		}
	case int32:
	}
}

func (o *IntrusionDetectSystem) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.k32Response, err = utils.ParseAdaptiveCmd(o.K32ResponseIn)
	if err != nil {
		panic(err)
	}
	for _, rc := range o.RegexCheckers {
		rc.compiledItemNameRegex = *regexp.MustCompile(rc.Item)
		rc.compiledValueRegex = *regexp.MustCompile(rc.RegexString)
	}
}

func (o *IntrusionDetectSystem) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddPlayer, func(p packet.Packet) {
		o.onSeePlayer(p.(*packet.AddPlayer))
	})
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDMobEquipment, func(p packet.Packet) {
		o.onSeeMobItem(p.(*packet.MobEquipment))
	})
}

func (o *IntrusionDetectSystem) doNbtCheck(rtid int32, nbt map[string]interface{}, getPlayerName func() string, getPacketString func() string) {
	has32K := false
	reason := ""
	if !has32K {
		has32K, reason = o.k32NbtDetect(nbt)
	}
	if !has32K {
		has32K, reason = o.regexNbtDetect(rtid, nbt)
	}
	if has32K {
		player := getPlayerName()
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("发现持有非法物资玩家 %v: %v", player, reason))
		o.Frame.GetBackendDisplay().Write(getPacketString())
		go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.k32Response, map[string]interface{}{
			"[player]": "\"" + player + "\"",
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *IntrusionDetectSystem) onSeeMobItem(pk *packet.MobEquipment) {
	if pk.EntityRuntimeID < 2 {
		// do not check bot
		return
	}
	rtid := pk.NewItem.Stack.NetworkID
	nbt := pk.NewItem.Stack.NBTData
	o.doNbtCheck(rtid, nbt, func() string {
		playerName := "未知玩家"
		for _, p := range o.Frame.GetUQHolder().PlayersByEntityID {
			if p.Entity != nil && p.Entity.RuntimeID == pk.EntityRuntimeID {
				playerName = p.Username
			}
		}
		return playerName
	}, func() string {
		marshal, err := json.Marshal(pk)
		if err != nil {
			return err.Error()
		} else {
			return string(marshal)
		}
	})
}

func findAndPrintK(key string, val interface{}, debug bool, onKey func(string)) {
	switch value := val.(type) {
	case map[string]interface{}:
		for k, v := range value {
			if debug {
				pterm.Info.Printfln("debug: current key=%v", k)
			}
			if k == key {
				vs := fmt.Sprintf("%v", v)
				onKey(vs)
			} else {
				findAndPrintK(key, v, debug, onKey)
			}
		}
	case []interface{}:
		for _, v := range value {
			findAndPrintK(key, v, debug, onKey)
		}
	case int32:
	}
}

func (o *IntrusionDetectSystem) k32NbtDetect(nbt map[string]interface{}) (has32K bool, reason string) {
	has32K, reason = false, ""
	if o.EnableK32Detect {
		findK("lvl", nbt, func(v interface{}) {
			level := int(v.(int16))
			if level > o.K32Threshold || level < 0 {
				has32K = true
				reason = fmt.Sprintf("持有32k(%v)", level)
			}

		})
	}
	return has32K, reason
}

func (o *IntrusionDetectSystem) regexNbtDetect(rtid int32, nbt map[string]interface{}) (has32K bool, reason string) {
	defer func() {
		r := recover()
		if r != nil {
			pterm.Error.Println(r)
		}
	}()
	itemName := items.ItemRuntimeIDToNameMapping(rtid)
	// fmt.Println(rtid, itemName)
	for _, regexCheck := range o.RegexCheckers {
		if has32K {
			break
		}
		if !regexCheck.Enabled {
			continue
		}
		debug := regexCheck.Debug
		if debug {
			pterm.Info.Printfln("正在调试正则表达式检查器\"%v\":检测符合\"%v\"的手持物品的nbt中tag \"%v\" 对应值是否符合 \"%v\" (调试模式)",
				regexCheck.Description,
				regexCheck.Item,
				regexCheck.Tag, regexCheck.RegexString)
		}
		matchName := regexCheck.compiledItemNameRegex.Find([]byte(itemName))
		reason := ""
		if matchName == nil {
			if debug {
				pterm.Info.Printfln("物品名\"%v\"不匹配指定的正则表达式\"%v\"", itemName, regexCheck.Item)
			}
			continue
		} else {
			if debug {
				pterm.Warning.Printfln("物品名\"%v\"匹配指定的正则表达式\"%v\",匹配项为\"%v\"", itemName, regexCheck.Item, string(matchName))
				s, err := json.Marshal(nbt)
				if err == nil {
					pterm.Info.Println("完整nbt为" + string(s))
				} else {
					pterm.Info.Println("完整nbt获取失败" + err.Error())
				}
			}
			reason = fmt.Sprintf("物品名\"%v\"匹配指定的正则表达式\"%v\",匹配项为\"%v\" ", itemName, regexCheck.Item, string(matchName))
		}
		tag := regexCheck.Tag
		doMatch := func(s string) (has32K bool) {
			if debug {
				pterm.Info.Printfln("key: \"%v\" value: \"%v\" => 检测是否匹配 \"%v\"", regexCheck.Tag, s, regexCheck.RegexString)
			}
			match := regexCheck.compiledValueRegex.Find([]byte(s))
			reason += ",且对于tag:\"" + tag + "\","
			if match == nil {
				if debug {
					pterm.Warning.Println("没有匹配项\n")
				}
				reason += "没有匹配项"
				if regexCheck.Allow == true {
					has32K = true
					reason += "，模式设置为匹配成功时放行，现在匹配不成功，因此认为作弊"
				}
			} else {
				if debug {
					pterm.Warning.Printfln("发现匹配项目:\"%v\"", string(match))
				}
				reason += fmt.Sprintf("中:\"%v\"命中正则匹配式:\"%v\"(\"%v\")", string(match), regexCheck.Description, regexCheck.RegexString)
				if regexCheck.Allow == false {
					has32K = true
					reason += "，模式设置为匹配成功时作弊，现在匹配成功了，因此认为作弊"
				}
			}
			if has32K {
				if debug {
					pterm.Error.Printfln("发现32k，具体判断理由为：%v，当前处于调试模式，因此不会实际执行反制指令", reason)
					return false
				} else {
					return true
				}
			} else {
				if debug {
					pterm.Success.Printfln("该物品不是作弊物品")
				}
				return false
			}

		}
		if tag == "" {
			s, err := json.Marshal(nbt)
			if err == nil {
				if doMatch(string(s)) {
					return true, reason
				}
			} else {
				fmt.Println(err)
			}
		} else {
			findAndPrintK(regexCheck.Tag, nbt, debug, func(s string) {
				if !has32K {
					has32K = doMatch(string(s))
				}
			})
		}
	}
	return has32K, reason
}

func (o *IntrusionDetectSystem) onSeePlayer(pk *packet.AddPlayer) {
	//name := pk.Username
	nbt := pk.HeldItem.Stack.NBTData
	rtid := pk.HeldItem.Stack.NetworkID
	o.doNbtCheck(rtid, nbt, func() string {
		return pk.Username
	}, func() string {
		marshal, err := json.Marshal(pk)
		if err != nil {
			return err.Error()
		} else {
			return string(marshal)
		}
	})
}
