package mctype

type MainConfig struct {
	Execute               string
	Block, OldBlock       *ConstBlock
	Begin, End, Position  Position
	Radius                int
	Length, Width, Height int
	Method, OldMethod     string
	Facing, Path, Shape   string
}
