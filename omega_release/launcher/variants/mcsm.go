//go:build mcsm
// +build mcsm

package variants

import (
	"os"
	"time"

	"github.com/pterm/pterm"
)

func IsMCSM() bool {
	return true
}

func PrintAuxReturn() {
	// fmt.Println("\b")
}

func PrintVariant() {
	time.Sleep(1)
	pterm.Info.Println("MCSM Special Version")
}

func GetCurrentDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return currentDir
}
