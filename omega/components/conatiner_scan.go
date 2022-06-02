package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"

	"github.com/pterm/pterm"
)

type ContainerScan struct {
	*BasicComponent
	EnableK32Detect bool                   `json:"启用32容器检测"`
	K32Threshold    int                    `json:"32k物品附魔等级阈值"`
	k32Response     []defines.Cmd          `json:"32k容器反制"`
	RegexCheckers   []*ContainerRegexCheck `json:"使用以下正则表达式检查"`
}

type ContainerRegexCheck struct {
	Enabled               bool        `json:"启用"`
	Description           string      `json:"检测说明"`
	Debug                 bool        `json:"调试模式"`
	Tag                   string      `json:"匹配标签名"`
	RegexString           string      `json:"使用正则表达式匹配标签值"`
	Allow                 bool        `json:"匹配标签值成功时true为放行false为作弊"`
	ExtraCommandIn        interface{} `json:"附加指令"`
	extraCommands         []defines.Cmd
	compiledItemNameRegex regexp.Regexp
	compiledValueRegex    regexp.Regexp
}

func (o *ContainerScan) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.k32Response, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"32k容器反制"})
	if err != nil {
		panic(err)
	}
	for _, rc := range o.RegexCheckers {
		rc.compiledValueRegex = *regexp.MustCompile(rc.RegexString)
		if rc.ExtraCommandIn == nil {
			rc.extraCommands = make([]defines.Cmd, 0)
		} else {
			if rc.extraCommands, err = utils.ParseAdaptiveCmd(rc.ExtraCommandIn); err != nil {
				panic(err)
			}
		}
	}
}

func (o *ContainerScan) regexNbtDetect(nbt map[string]interface{}, x, y, z int) (has32K bool, reason string) {
	for _, regexCheck := range o.RegexCheckers {
		if has32K {
			break
		}
		if !regexCheck.Enabled {
			continue
		}
		debug := regexCheck.Debug
		if debug {
			pterm.Info.Printfln("正在调试正则表达式检查器\"%v\":检测nbt方块中tag \"%v\" 对应值是否符合 \"%v\" (调试模式)",
				regexCheck.Description,
				regexCheck.Tag, regexCheck.RegexString)
			s, err := json.Marshal(nbt)
			if err == nil {
				pterm.Info.Println("完整nbt为" + string(s))
			} else {
				pterm.Info.Println("完整nbt获取失败" + err.Error())
			}
		}
		reason := ""
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
					if len(regexCheck.extraCommands) > 0 {
						mapping := map[string]interface{}{
							"[x]": x,
							"[y]": y,
							"[z]": z,
						}
						for i := 0; i < 4; i++ {
							mapping[fmt.Sprintf("[x+%v]", i)] = x + i
							mapping[fmt.Sprintf("[y+%v]", i)] = y + i
							mapping[fmt.Sprintf("[z+%v]", i)] = z + i
						}
						for i := -3; i < 0; i++ {
							mapping[fmt.Sprintf("[x%v]", i)] = x + i
							mapping[fmt.Sprintf("[y%v]", i)] = y + i
							mapping[fmt.Sprintf("[z%v]", i)] = z + i
						}
						utils.LaunchCmdsArray(o.Frame.GetGameControl(), regexCheck.extraCommands, mapping, o.Frame.GetBackendDisplay())
					}
					return true
				}
			} else {
				if debug {
					pterm.Success.Printfln("该方块不是作弊方块")
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

func (o *ContainerScan) checkNbt(x, y, z int, nbt map[string]interface{}, getStr func() string) {
	has32K := false
	reason := ""
	if o.EnableK32Detect {
		findK("lvl", nbt, func(v interface{}) {
			if level, success := v.(int16); success {
				if int(level) > o.K32Threshold {
					has32K = true
					reason = fmt.Sprintf("32k 方块：%v > %v", int(level), o.K32Threshold)
				} else if int(level) < 0 {
					has32K = true
					reason = fmt.Sprintf("32k 方块：%v < 0", int(level))
				}
			}
		})
	}
	if !has32K {
		has32K, reason = o.regexNbtDetect(nbt, x, y, z)
	}
	if has32K {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("位于 %v %v %v 的方块:"+reason, x, y, z))
		utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.k32Response, map[string]interface{}{
			"[x]": x,
			"[y]": y,
			"[z]": z,
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *ContainerScan) onLevelChunk(cd *mirror.ChunkData) {
	for _, nbt := range cd.BlockNbts {
		if x, y, z, success := define.GetPosFromNBT(nbt); success {
			o.checkNbt(int(x), int(y), int(z), nbt, func() string {
				marshal, _ := json.Marshal(nbt)
				return string(marshal)
			})
		}
	}
}

func (o *ContainerScan) onBlockActorData(pk *packet.BlockActorData) {
	nbt := pk.NBTData
	x, y, z := pk.Position.X(), pk.Position.Y(), pk.Position.Z()
	o.checkNbt(int(x), int(y), int(z), nbt, func() string {
		marshal, _ := json.Marshal(nbt)
		return string(marshal)
	})
}

func (o *ContainerScan) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDBlockActorData, func(p packet.Packet) {
		o.onBlockActorData(p.(*packet.BlockActorData))
	})
	o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.onLevelChunk)
}
