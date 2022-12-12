package liliya

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strings"
	"time"
)

type AbilityManage struct {
	*defines.BasicComponent
	Duration           int             `json:"检查周期"`
	AllowFlight        *AbilityDetails `json:"开启飞行"`
	NoClip             *AbilityDetails `json:"关闭碰撞"`
	Mute               *AbilityDetails `json:"禁止发言"`
	NoMine             *AbilityDetails `json:"禁止破坏方块"`
	NoDoorsAndSwitches *AbilityDetails `json:"禁止使用门与开关"`
	NoOpenContainers   *AbilityDetails `json:"禁止打开容器"`
	NoAttackPlayers    *AbilityDetails `json:"禁止攻击玩家"`
	NoAttackMobs       *AbilityDetails `json:"禁止攻击生物"`
	Operator           *AbilityDetails `json:"操作员命令"`
	NoTeleport         *AbilityDetails `json:"禁止使用传送"`
	NoBuild            *AbilityDetails `json:"禁止放置方块"`
}

type AbilityDetails struct {
	Enable bool     `json:"启用"`
	Always bool     `json:"持续生效"`
	Tags   []string `json:"标签"`
	Msg1   string   `json:"开启时提示"`
	Msg2   string   `json:"关闭时提示"`
}

type AbilitySettings struct {
	Type    *uint32
	Pos     int
	Reverse bool
}

func (o *AbilityManage) Init(cfg *defines.ComponentConfig) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
	if o.Duration < 1 {
		panic("检查周期不能小于1, 请更改配置文件")
	}
}

func (o *AbilityManage) Inject(frame defines.MainFrame) {
	o.Frame = frame
}

func cleanTags(tags []string) []string {
	for k, v := range tags {
		tags[k] = strings.TrimSuffix(strings.TrimPrefix(v, "§a"), "§r")
	}
	return tags
}

func hasElementInOtherSlice(a []string, b []string) (bool, string) {
	for _, v1 := range a {
		for _, v2 := range b {
			if v1 == v2 {
				return true, v1
			}
		}
	}
	return false, ""
}

func (o *AbilityManage) switchAbility(tags []string, pd *AbilityDetails, ps []AbilitySettings, name string) bool {
	if has, tag := hasElementInOtherSlice(tags, pd.Tags); pd.Enable && (pd.Always || has) {
		sendFeedbackMsg := func(status bool) {
			if status {
				if pd.Msg1 != "" {
					o.Frame.GetGameControl().SayTo(name, pd.Msg1)
				}
			} else {
				if pd.Msg2 != "" {
					o.Frame.GetGameControl().SayTo(name, pd.Msg2)
				}
			}
		}
		changeAbility := func() (result bool) {
			for i, v := range ps {
				if (*v.Type>>v.Pos%2 == 1) != (has != v.Reverse) {
					*v.Type = *v.Type ^ (1 << v.Pos)
					result = true
					if i == 0 {
						if v.Reverse {
							sendFeedbackMsg(!(*v.Type>>v.Pos%2 == 1))
						} else {
							sendFeedbackMsg(*v.Type>>v.Pos%2 == 1)
						}
					}
				}
			}
			return result
		}
		if changeAbility() {
			if !pd.Always {
				o.Frame.GetGameControl().SendWOCmd(fmt.Sprintf("tag \"%s\" remove %s", name, tag))
			}
			return true
		}
	}
	return false
}

func (o *AbilityManage) getPermissionLevel(ap uint32) (commandPermissionLevel, permissionLevel uint32) {
	if ap>>5%2 == 1 {
		commandPermissionLevel = packet.CommandPermissionLevelHost
	}
	switch ap {
	case 447:
		permissionLevel = packet.PermissionLevelOperator
	case 287:
		permissionLevel = packet.PermissionLevelMember
	default:
		if ap != 0 {
			permissionLevel = packet.PermissionLevelCustom
		}
	}
	return commandPermissionLevel, permissionLevel
}

func (o *AbilityManage) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Duration))
	for {
		<-t.C
		for _, v := range o.Frame.GetUQHolder().PlayersByEntityID {
			if v.Username == o.Frame.GetUQHolder().GetBotName() {
				continue
			}
			o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("tag \"%s\" list", v.Username), func(output *packet.CommandOutput) {
				if output.SuccessCount > 0 {
					var tags []string
					if output.OutputMessages[0].Message != "commands.tag.list.single.empty" {
						tags = cleanTags(strings.Split(output.OutputMessages[0].Parameters[2], ", "))
					}
					pl := o.Frame.GetUQHolder().GetPlayersByUUID(o.Frame.GetGameControl().GetPlayerKit(output.OutputMessages[0].Parameters[0]).GetRelatedUQ().UUID)
					pf, ap, name := pl.PropertiesFlag, pl.ActionPermissions, pl.Username
					isChanged := o.switchAbility(tags, o.AllowFlight, []AbilitySettings{{Type: &pf, Pos: 6, Reverse: false}, {Type: &pf, Pos: 9, Reverse: false}}, name)
					isChanged = o.switchAbility(tags, o.NoClip, []AbilitySettings{{Type: &pf, Pos: 7, Reverse: false}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.Mute, []AbilitySettings{{Type: &pf, Pos: 10, Reverse: false}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.NoMine, []AbilitySettings{{Type: &ap, Pos: 0, Reverse: true}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.NoDoorsAndSwitches, []AbilitySettings{{Type: &ap, Pos: 1, Reverse: true}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.NoOpenContainers, []AbilitySettings{{Type: &ap, Pos: 2, Reverse: true}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.NoAttackPlayers, []AbilitySettings{{Type: &ap, Pos: 3, Reverse: true}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.NoAttackMobs, []AbilitySettings{{Type: &ap, Pos: 4, Reverse: true}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.Operator, []AbilitySettings{{Type: &ap, Pos: 5, Reverse: false}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.NoTeleport, []AbilitySettings{{Type: &ap, Pos: 7, Reverse: true}}, name) || isChanged
					isChanged = o.switchAbility(tags, o.NoBuild, []AbilitySettings{{Type: &ap, Pos: 8, Reverse: true}}, name) || isChanged
					if isChanged {
						cpml, pml := o.getPermissionLevel(ap)
						o.Frame.GetGameControl().SendMCPacket(&packet.AdventureSettings{
							Flags:                  pf,
							CommandPermissionLevel: cpml,
							ActionPermissions:      ap,
							PermissionLevel:        pml,
							PlayerUniqueID:         pl.EntityUniqueID,
						})
					}
				}
			})
		}
	}
}
