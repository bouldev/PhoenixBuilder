package types

const (
	DelayModeContinuous = 0
	DelayModeDiscrete   = 1
	DelayModeNone       = 2
	DelayModeInvalid    = 100
)

const (
	TaskTypeSync    = 0
	TaskTypeAsync   = 1
	TaskTypeInvalid = 100
)

const (
	TaskDisplayYes     = 1
	TaskDisplayNo      = 0
	TaskDisplayInvalid = 30
)

type MainConfig struct {
	Execute               string
	Block, OldBlock       *ConstBlock
	Entity                string
	End, Position         Position //Position=Begin
	ResumeFrom            float64
	Radius                int
	Length, Width, Height int
	MapX, MapZ, MapY      int
	Method, OldMethod     string
	Facing, Path, Shape   string
	AssignNBTData         bool
	ExcludeCommands       bool
	InvalidateCommands    bool
	Strict                bool
}

type DelayConfig struct {
	Delay          int64
	DelayMode      byte
	DelayThreshold int
}

type GlobalConfig struct {
	TaskCreationType byte
	TaskDisplayMode  byte
}

func ParseDelayMode(mode string) byte {
	if mode == "continuous" {
		return DelayModeContinuous
	} else if mode == "discrete" {
		return DelayModeDiscrete
	} else if mode == "none" {
		return DelayModeNone
	}
	return DelayModeInvalid
}

func StrDelayMode(mode byte) string {
	if mode == DelayModeContinuous {
		return "continuous"
	} else if mode == DelayModeDiscrete {
		return "discrete"
	} else if mode == DelayModeNone {
		return "none"
	} else {
		return "invalid"
	}
}

func ParseTaskType(mode string) byte {
	if mode == "sync" {
		return TaskTypeSync
	} else if mode == "async" {
		return TaskTypeAsync
	}
	return TaskTypeInvalid
}

func MakeTaskType(mode byte) string {
	if mode == TaskTypeSync {
		return "sync"
	} else if mode == TaskTypeAsync {
		return "async"
	}
	return "invalid"
}

func ParseTaskDisplayMode(mode string) byte {
	if mode == "true" {
		return TaskDisplayYes
	} else if mode == "false" {
		return TaskDisplayNo
	}
	return TaskDisplayInvalid
}

func MakeTaskDisplayMode(mode byte) string {
	if mode == TaskDisplayYes {
		return "true"
	} else if mode == TaskDisplayNo {
		return "false"
	}
	return "?"
}
