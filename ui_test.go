package main

import (
	"github.com/pterm/pterm"
	"testing"
)

func TestGUI(t *testing.T) {
	pterm.Println(pterm.Yellow("FastBuilder Phoenix Alpha 0.0.2"))
	pterm.DefaultBox.Println(pterm.LightCyan("Copyright notice: \n" +
		"FastBuilder Phoenix used codes\n" +
		"from Sandertv's Gophertunnel that\n" +
		"licensed under MIT license,at:\n" +
		"https://github.com/Sandertv/gophertunnel"))
	pterm.Println(pterm.Yellow("ファスト　ビルダー！"))
	pterm.Println(pterm.Yellow("F A S T  B U I L D E R"))
	pterm.Println(pterm.Yellow("Contributors: Ruphane, CAIMEO"))
	pterm.Println(pterm.Yellow("Copyright (c) FastBuilder DevGroup, Bouldev 2021"))
}
