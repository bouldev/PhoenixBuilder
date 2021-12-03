package types

type Position struct {
	X, Y, Z int
}

func (p Position) FromInt(arr []int) {
	p.X = arr[0]
	p.Y = arr[1]
	p.Z = arr[2]
}

type FloatPosition struct {
	X, Y, Z float64
}

func (p *FloatPosition) TransferInt() Position {
	return Position{
		int(p.X),
		int(p.Y),
		int(p.Z),
	}
}