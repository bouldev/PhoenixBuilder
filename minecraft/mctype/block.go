package mctype


type Module struct {
	Block  Block
	Entity Entity
	Point  Position
}

type Block struct {
	Name string
	Data int
}

type DoubleModule struct {
	Begin           Position
	End             Position
	Block, OldBlock Block
	Entity          Entity
}