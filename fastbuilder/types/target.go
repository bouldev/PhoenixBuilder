package types


type Target string

const (
	AllPlayers    Target = "@a"
	AllEntities   Target = "@e"
	NearestPlayer Target = "@p"
	RandomPlayer  Target = "@r"
	Self          Target = "@s"
)


