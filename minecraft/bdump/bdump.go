package bdump

import (
	"phoenixbuilder/minecraft/mctype"
	//"bufio"
)

type BDump struct {
	Author string
	Blocks []*mctype.Module
}

func (bdump *BDump) formatBlocks() {
	min:=[]int{2147483647,2147483647,2147483647}
	for _, mdl := range bdump.Blocks {
		if mdl.Point.X<min[0] {
			min[0]=mdl.Point.X
		}
		if mdl.Point.Y<min[1] {
			min[1]=mdl.Point.Y
		}
		if mdl.Point.Z<min[2] {
			min[2]=mdl.Point.Z
		}
	}
	for _, mdl := range bdump.Blocks {
		mdl.Point.X-=min[0]
		mdl.Point.Y-=min[1]
		mdl.Point.Z-=min[2]
	}
}

func (bdump *BDump) packBlocks() {
	
}