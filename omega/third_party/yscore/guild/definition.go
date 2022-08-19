package guild

import (
	"phoenixbuilder/omega/defines"
)

type Guild struct {
	*defines.BasicComponent
	FistCmds               []string              `json:"一级保护指令"`
	FistCmdTarget          string                `json:"一级保护触发选择器"`
	ThePermissionsOfGuild  map[string]int        `json:"最低公会等级可开启的功能"`
	ThePermissionsOfMember map[string]int        `json:"最低权限可使用功能"`
	Triggers               []string              `json:"触发词"`
	Usage                  string                `json:"提示词"`
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
	IsNeedTerr             bool                  `json:"是否需要自带领地"`
	UpgradePrice           map[string]string     `json:"每次公会升级为下一级所需贡献值"`
	IsAllowKick            bool                  `json:"是否允许公会可以kick"`
	IsYsCore               bool                  `json:"是否开启yscore专属公会"`
	YsCoreDefines          *Yscore               `json:"yscore会员配置"`

	BuffList  map[string]*Buff
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
	AllyData        map[string]string
	PendingAlly     map[string]string
	YscoreScore     int //公会点
	Master          string
	Member          map[string]*GuildDtails //记得初始化（）
	SpPlace         map[string][]int        //[起点x 起点y 起点z dx dy dz]
	Range           []int
	Pos             []int
	CenterPos       []int
	IsTerr          bool
	Power           int
	ApplicationList []string
	GuildRankings   int              //在计分板内分数（）
	HolyRelics      string           //圣遗物
	TpPos           map[string][]int //公会传送点
}
type User struct {
	Name []string `json:"victim"`
}
type GuildDtails struct {
	Announcement string
	Permistion   string
	title        []string
}
