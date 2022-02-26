package utils

import (
	"fyne.io/fyne/v2"
	"golang.org/x/mobile/asset"
	"io/ioutil"
)

func LoadFromAssets(asName string, path string) (fyne.Resource, error) {
	f, err := asset.Open(path)
	if err != nil {
		return nil, err
	}
	buf, errRead := ioutil.ReadAll(f)
	f.Close()
	if errRead != nil {
		return nil, err
	}
	return fyne.NewStaticResource(asName, buf), nil
}
