package main

import (
	"phoenixbuilder/minecraft/builder"
	"phoenixbuilder/minecraft/mctype"
	"testing"
)

func TestBuilder(t *testing.T) {
	mcfg := mctype.MainConfig{
		Execute: "acme",
		Block: mctype.Block{
			Name: "",
			Data: 0,
		},
		OldBlock: mctype.Block{
			Name: "",
			Data: 0,
		},
		Begin: mctype.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		End:       mctype.Position{},
		Position:  mctype.Position{},
		Radius:    0,
		Length:    0,
		Width:     0,
		Height:    0,
		Method:    "",
		OldMethod: "",
		Facing:    "y",
		Path:      "../210606_115440_79x87x115.mcacblock",
		Shape:     "",
	}
	b, err := builder.Generate(mcfg)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(b)
}
