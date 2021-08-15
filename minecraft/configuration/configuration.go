package configuration

import (
	"phoenixbuilder/minecraft/mctype"
	//"phoenixbuilder/minecraft/builder"
	"github.com/google/uuid"
)

const (
	ConfigTypeMain   = 0
	ConfigTypeDelay  = 1
	ConfigTypeGlobal = 2
)

var AirBlock = &mctype.ConstBlock{Name: "air", Data: 0}
var IronBlock = &mctype.ConstBlock{Name: "iron_block", Data: 0}

type FullConfig map[byte]interface{}
var ForwardedBrokSender chan string

func ConcatFullConfig(mc *mctype.MainConfig, dc *mctype.DelayConfig) *FullConfig {
	mco:=*mc
	dco := *dc
	return &FullConfig {
		ConfigTypeMain: &mco,
		ConfigTypeDelay: &dco,
	}
}

func CreateFullConfig() *FullConfig {
	mConfig := mctype.MainConfig{
		Execute: "",
		Block: IronBlock,
		OldBlock: AirBlock,
		/*Begin: mctype.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},*/
		End: mctype.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		Position: mctype.Position{
			X: 200,
			Y: 100,
			Z: 200,
		},
		Radius:    5,
		Length:    0,
		Width:     0,
		Height:    1,
		Method:    "",
		OldMethod: "",
		Facing:    "y",
		Path:      "",
		Shape:     "solid",
	}
	dConf := mctype.DelayConfig {
		Delay:     decideDelay(mctype.DelayModeContinuous),
		DelayMode: mctype.DelayModeContinuous,
		DelayThreshold:decideDelayThreshold(),
	}
	gConf := mctype.GlobalConfig {
		TaskCreationType: mctype.TaskTypeAsync,
		TaskDisplayMode:  mctype.TaskDisplayYes,
	}
	fc := make(FullConfig)
	fc[ConfigTypeMain]=&mConfig
	fc[ConfigTypeDelay]=&dConf
	fc[ConfigTypeGlobal]=&gConf
	return &fc
}

var RespondUser string
var ZeroId uuid.UUID
var OneId uuid.UUID

var UserToken string

var globalFullConfig *FullConfig

func GlobalFullConfig() *FullConfig {
	if globalFullConfig == nil {
		globalFullConfig = CreateFullConfig()
	}
	return globalFullConfig
}

func (conf *FullConfig) Main() *mctype.MainConfig {
	mConf, _ :=(*conf)[ConfigTypeMain].(*mctype.MainConfig)
	return mConf
}

func (conf *FullConfig) Delay() *mctype.DelayConfig {
	dConf, _ :=(*conf)[ConfigTypeDelay].(*mctype.DelayConfig)
	return dConf
}

func (conf *FullConfig) Global() *mctype.GlobalConfig {
	gConf, _ :=(*conf)[ConfigTypeGlobal].(*mctype.GlobalConfig)
	return gConf
}

func decideDelay(delaytype byte) int64 {
	// Will add system check later,so don't merge into other functions.
	if delaytype==mctype.DelayModeContinuous {
		return 1000
	}else if delaytype==mctype.DelayModeDiscrete {
		return 15
	}else{
		return 0
	}
}

func decideDelayThreshold() int {
	// Will add system check later,so don't merge into other functions.
	return 20000
}