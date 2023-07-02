//go:build windows

package fixtimer

import (
	"fmt"
	"syscall"
)

func init() {
	var (
		winmmDLL            = syscall.NewLazyDLL("winmm.dll")
		procTimeBeginPeriod = winmmDLL.NewProc("timeBeginPeriod")
	)
	//fmt.Println("DEBUG: Try to use high precision timer")
	procTimeBeginPeriod.Call(uintptr(1))
}
