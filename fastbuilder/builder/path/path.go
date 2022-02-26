// +build !fyne_gui

package path

import (
	"os"
)

func ReadFile(path string) (FileReader,error){
	file, err := os.Open(path)
	return file,err
}