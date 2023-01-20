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
	Flag    uint32
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

// 返回的tag前后带有颜色符号, 这里负责将它们去除
func cleanTags(tags []string) []string {
	for k, v := range tags {
		tags[k] = strings.TrimSuffix(strings.TrimPrefix(v, "§a"), "§r")
	}
	return tags
}

// 查询两个切片是否有相同的元素, 并返回这个元素
func hasElementInOtherSlice(a []string, b []string) (string, bool) {
	for _, v1 := range a {
		for _, v2 := range b {
			if v1 == v2 {
				return v1, true
			}
		}
	}
	return "", false
}

// 这里根据给定的参数来变更能力, 且会向玩家发送对应的提示; 返回值代表是否进行了变更, 后续会根据变更与否来决定是否发送数据包
func (o *AbilityManage) switchAbility(tags []string, pd *AbilityDetails, as []AbilitySettings, name string) bool {
	if tag, has := hasElementInOtherSlice(tags, pd.Tags); pd.Enable && (pd.Always || has) {
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
			for i, v := range as {
				if (*v.Type&v.Flag != 0) != (has != v.Reverse) {
					*v.Type = *v.Type ^ v.Flag
					result = true
					// 为了避免多次向玩家发送提示
					if i == 0 {
						if v.Reverse {
							sendFeedbackMsg(*v.Type&v.Flag == 0)
						} else {
							sendFeedbackMsg(*v.Type&v.Flag != 0)
						}
					}
				}
			}
			return result
		}
		if changeAbility() {
			// 非持续生效, 移除tag
			if !pd.Always {
				o.Frame.GetGameControl().SendWOCmd(fmt.Sprintf("tag \"%s\" remove %s", name, tag))
			}
			return true
		}
	}
	return false
}

// 正常情况下, commandPermissionLevel 与 permissionLevel 应该由 actionPermissions 决定的
func (o *AbilityManage) getPermissionLevel(actionPermissions uint32) (commandPermissionLevel, permissionLevel uint32) {
	if actionPermissions&packet.ActionPermissionOperator != 0 {
		commandPermissionLevel = packet.CommandPermissionLevelHost
	}
	switch actionPermissions {
	case 447:
		permissionLevel = packet.PermissionLevelOperator
	case 287:
		permissionLevel = packet.PermissionLevelMember
	default:
		if actionPermissions != 0 {
			permissionLevel = packet.PermissionLevelCustom
		}
	}
	return commandPermissionLevel, permissionLevel
}

func (o *AbilityManage) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Duration))
	for {
		for _, v := range o.Frame.GetUQHolder().PlayersByEntityID {
			// 跳过Bot, 以防止一些情况的出现
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

					// 开启飞行, 需要同时设置以下两个能力, 否则生存模式会无法飞行
					isChanged := o.switchAbility(tags, o.AllowFlight, []AbilitySettings{
						{Type: &pf, Flag: packet.AdventureFlagFlying, Reverse: false},
						{Type: &pf, Flag: packet.AdventureFlagAllowFlight, Reverse: false},
					}, name)

					// 关闭碰撞
					isChanged = o.switchAbility(tags, o.NoClip, []AbilitySettings{
						{Type: &pf, Flag: packet.AdventureFlagNoClip, Reverse: false},
					}, name) || isChanged

					// 禁止发言
					isChanged = o.switchAbility(tags, o.Mute, []AbilitySettings{
						{Type: &pf, Flag: packet.AdventureFlagMuted, Reverse: false},
					}, name) || isChanged

					// 禁止破坏方块, 这里使用 Reverse, 因为默认情况下是允许破坏方块的
					isChanged = o.switchAbility(tags, o.NoMine, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionMine, Reverse: true},
					}, name) || isChanged

					// 禁止使用门与开关
					isChanged = o.switchAbility(tags, o.NoDoorsAndSwitches, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionDoorsAndSwitches, Reverse: true},
					}, name) || isChanged

					// 禁止打开容器
					isChanged = o.switchAbility(tags, o.NoOpenContainers, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionOpenContainers, Reverse: true},
					}, name) || isChanged

					// 禁止攻击玩家
					isChanged = o.switchAbility(tags, o.NoAttackPlayers, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionAttackPlayers, Reverse: true},
					}, name) || isChanged

					// 禁止攻击生物
					isChanged = o.switchAbility(tags, o.NoAttackMobs, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionAttackMobs, Reverse: true},
					}, name) || isChanged

					// 操作员命令
					isChanged = o.switchAbility(tags, o.Operator, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionOperator, Reverse: false},
					}, name) || isChanged

					// 禁止使用传送
					isChanged = o.switchAbility(tags, o.NoTeleport, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionTeleport, Reverse: true},
					}, name) || isChanged

					// 禁止放置方块
					isChanged = o.switchAbility(tags, o.NoBuild, []AbilitySettings{
						{Type: &ap, Flag: packet.ActionPermissionBuild, Reverse: true},
					}, name) || isChanged

					// 如果调用现有api, 可能会发送多个数据包来完成相同的任务, 这里选择将它们放在一个数据包里发送
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
		<-t.C
	}
}
