package mctype

type MainConfig struct {
	Execute               string
	Block, OldBlock       Block
	Begin, End, Position  Position
	Radius                int
	Length, Width, Height int
	Method, OldMethod     string
	Facing, Path, Shape   string
}
