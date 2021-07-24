package mctype

const (
	DelayModeContinuous = 0
	DelayModeDiscrete   = 1
	DelayModeNone       = 2
	DelayModeInvalid    = 100
)

type MainConfig struct {
	Execute               string
	Block, OldBlock       *ConstBlock
	Begin, End, Position  Position
	Radius                int
	Length, Width, Height int
	Method, OldMethod     string
	Facing, Path, Shape   string
	Delay                 int64
	DelayMode             byte
	DelayThreshold        int
}

func ParseDelayMode(mode string) byte {
	if mode == "continuous" {
		return DelayModeContinuous
	}else if mode=="discrete" {
		return DelayModeDiscrete
	}else if mode=="none" {
		return DelayModeNone
	}
	return DelayModeInvalid
}

func StrDelayMode(mode byte) string {
	if mode==DelayModeContinuous {
		return "continuous"
	}else if mode==DelayModeDiscrete {
		return "discrete"
	}else if mode==DelayModeNone {
		return "none"
	}else{
		return "invalid"
	}
}