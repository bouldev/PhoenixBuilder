// +build !fyne_gui

package bridge_fmt

import "fmt"

func Printf(format string, args ...interface{}) (int, error) {
	return fmt.Printf(format, args...)
}

func Println(args ...interface{}) (int, error) {
	return fmt.Println(args)
}

func Print(s string) (int, error) {
	return fmt.Print(s)
}