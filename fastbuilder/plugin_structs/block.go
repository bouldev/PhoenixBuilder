package plugin_structs

type Entity string

type Module struct {
	Block  *Block
	CommandBlockData *CommandBlockData
	Entity *Entity
	Point  Position
}

type Block struct {
	Name *string
	Data uint16
}

type CommandBlockData struct {
	Mode uint32
	Command string
	CustomName string
	LastOutput string
	TickDelay int32
	ExecuteOnFirstTick bool //byte
	TrackOutput bool //byte
	Conditional bool
	NeedRedstone bool
}

type ConstBlock struct {
	Name string
	Data uint16
}

type DoubleModule struct {
	Begin           Position
	End             Position
	Block, OldBlock *Block
	Entity          *Entity
}

var takenBlocks map[*ConstBlock]*Block = make(map[*ConstBlock]*Block)

func CreateBlock(name string,data uint16) *Block {
	return &Block {
		Name:&name,
		Data:data,
	}
}

func (req *ConstBlock) Take() *Block {
	block, ok := takenBlocks[req]
	if ok {
		return block
	}
	block=&Block {
		Name:&req.Name, //ConstBlock shouldn't be destroyed
		Data:req.Data,
	}
	takenBlocks[req]=block
	return block
}