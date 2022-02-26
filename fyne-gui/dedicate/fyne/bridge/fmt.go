// +build fyne_gui

package bridge

import (
	"fmt"
)

var HookFunc func(string)

func init() {
	HookFunc = func(s string) {
		fmt.Print(s)
	}
}

func Printf(format string, args ...interface{}) (int, error) {
	s := fmt.Sprintf(format, args...)
	HookFunc(s)
	return len([]byte(s)), nil
}

func Println(args ...interface{}) (int, error) {
	s := fmt.Sprintf("%s\n", args...)
	HookFunc(s)
	return len([]byte(s)), nil
}

func Print(s string) (int, error) {
	HookFunc(s)
	return len([]byte(s)), nil
}
