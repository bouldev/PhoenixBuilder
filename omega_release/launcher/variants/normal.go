//go:build !mcsm
// +build !mcsm

package variants

import (
	"os"
	"path/filepath"
)

func IsMCSM() bool {
	return false
}

func PrintVariant() {
}

func PrintAuxReturn() {
}

func GetCurrentDir() string {
	pathExecutable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dirPathExecutable := filepath.Dir(pathExecutable)
	return dirPathExecutable
}
