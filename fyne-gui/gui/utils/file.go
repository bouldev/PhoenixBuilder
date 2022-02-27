package utils

import (
	"fyne.io/fyne/v2"
)

func WriteOrCreateFile(storage fyne.Storage,filename string,data []byte) error {
	// since append is not provided, we have to do this
	hasFileFlag := false
	for _, fn := range storage.List() {
		if fn == filename {
			hasFileFlag = true
			break
		}
	}
	var fp fyne.URIWriteCloser
	var err error
	if hasFileFlag {
		fp, err = storage.Save(filename)
	} else {
		fp, err = storage.Create(filename)
	}
	if err != nil {
		return err
	}
	_, err = fp.Write(data)
	defer fp.Close()
	return err
}