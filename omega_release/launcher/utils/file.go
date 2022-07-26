package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func CopyFile(src, dst string) (nBytes int64, err error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	defer destination.Close()
	nBytes, err = io.Copy(destination, source)
	return nBytes, err

}

func MakeDirP(path string) error {
	stat, err := os.Stat(path)
	if !(err == nil && stat.IsDir()) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if !(err == nil && stat.IsDir()) {
		return false
	}
	return true
}

func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if !(err == nil && !stat.IsDir()) {
		return false
	}
	return true
}

func GetFileData(fname string) ([]byte, error) {
	fp, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	return buf, err
}

func WriteFileData(fname string, data []byte) error {
	fp, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer fp.Close()
	if _, err := fp.Write(data); err != nil {
		return err
	}
	return nil
}

func WriteJsonData(fname string, data interface{}) error {
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	enc.SetEscapeHTML(false)
	err = enc.Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func GetJsonData(fname string, ptr interface{}) error {
	data, err := GetFileData(fname)
	if err != nil {
		return err
	}
	if data == nil || len(data) == 0 {
		return nil
	}
	err = json.Unmarshal(data, ptr)
	if err != nil {
		return err
	}
	return nil
}
