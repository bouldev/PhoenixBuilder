package mctype

type Position struct {
	X, Y, Z int
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