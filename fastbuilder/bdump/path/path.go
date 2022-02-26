// +build !fyne_gui

package path

import (
	"os"
)

func CreateFile(path string) (FileWriter,error){
	file, err:=os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE,0666)
	return file,err
}