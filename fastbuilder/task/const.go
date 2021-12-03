package task

import (
	"time"
	"fmt"
)

var ProgressThemes = []func(*AsyncInfo)string {
	func(asyncInfo *AsyncInfo)string {
		return fmt.Sprintf("%d/%d(%.2f%%) %.2fblocks/s",asyncInfo.Built,asyncInfo.Total,(float64(asyncInfo.Built)/float64(asyncInfo.Total))*100,float64(asyncInfo.Built)/time.Now().Sub(asyncInfo.BeginTime).Seconds())
	},
}