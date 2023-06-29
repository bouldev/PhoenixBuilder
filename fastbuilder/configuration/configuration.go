package configuration

import (
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/environment"
)

const (
	ConfigTypeMain   = 0
	ConfigTypeDelay  = 1
	ConfigTypeGlobal = 2
)

var AirBlock = &types.ConstBlock{Name: "air", Data: 0}
var IronBlock = &types.ConstBlock{Name: "iron_block", Data: 0}

type FullConfig map[byte]interface{}
var ForwardedBrokSender chan string

func ConcatFullConfig(mc *types.MainConfig, dc *types.DelayConfig) *FullConfig {
	mco:=*mc
	dco := *dc
	return &FullConfig {
		ConfigTypeMain: &mco,
		ConfigTypeDelay: &dco,
	}
}

func CreateFullConfig() *FullConfig {
	mConfig := types.MainConfig{
		Execute: "",
		Block: IronBlock,
		OldBlock: AirBlock,
		Position: types.Position{
			X: 200,
			Y: 100,
			Z: 200,
		},
		End: types.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		ResumeFrom: 0,
		Radius:    5,
		Length:    0,
		Width:     0,
		Height:    1,
		Method:    "",
		OldMethod: "",
		Facing:    "y",
		Path:      "",
		Shape:     "solid",
		MapX:      1,
		MapZ:      1,
		MapY:      0,
	}
	dConf := types.DelayConfig {
		Delay:     decideDelay(types.DelayModeContinuous),
		DelayMode: types.DelayModeContinuous,
		DelayThreshold:decideDelayThreshold(),
	}
	gConf := types.GlobalConfig {
		TaskCreationType: types.TaskTypeAsync,
		TaskDisplayMode:  types.TaskDisplayYes,
	}
	fc := make(FullConfig)
	fc[ConfigTypeMain]=&mConfig
	fc[ConfigTypeDelay]=&dConf
	fc[ConfigTypeGlobal]=&gConf
	return &fc
}


var IsOp bool
var SessionInitID int

var UserToken string

var globalFullConfig *FullConfig

func GlobalFullConfig(env *environment.PBEnvironment) *FullConfig {
	if env.GlobalFullConfig == nil {
		env.GlobalFullConfig = CreateFullConfig()
	}
	ret:=env.GlobalFullConfig.(*FullConfig)
	return ret
}

func (conf *FullConfig) Main() *types.MainConfig {
	mConf, _ :=(*conf)[ConfigTypeMain].(*types.MainConfig)
	return mConf
}

func (conf *FullConfig) Delay() *types.DelayConfig {
	dConf, _ :=(*conf)[ConfigTypeDelay].(*types.DelayConfig)
	return dConf
}

func (conf *FullConfig) Global() *types.GlobalConfig {
	gConf, _ :=(*conf)[ConfigTypeGlobal].(*types.GlobalConfig)
	return gConf
}

func decideDelay(delaytype byte) int64 {
	// Will add system check later,so don't merge into other functions.
	if delaytype==types.DelayModeContinuous {
		return 1000
	}else if delaytype==types.DelayModeDiscrete {
		return 15
	}else{
		return 0
	}
}

func decideDelayThreshold() int {
	// Will add system check later,so don't merge into other functions.
	return 20000
}