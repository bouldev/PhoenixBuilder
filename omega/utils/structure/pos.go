package structure

import (
	"phoenixbuilder/mirror/define"
	"time"
)

func AlterImportPosStartAndSpeed(inChan chan *IOBlock, offset define.CubePos, startFrom int, speed int, outChanLen int) (outChan chan *IOBlock, stopFn func()) {
	outChan = make(chan *IOBlock, outChanLen)
	stop := false
	go func() {
		counter := 0
		var ticker *time.Ticker
		for {
			if stop {
				return
			}
			if counter < startFrom {
				counter++
				<-inChan
			} else {
				delay := time.Duration((float64(1000) / float64(speed)) * float64(time.Millisecond))
				ticker = time.NewTicker(delay)
				break
			}
		}
		for b := range inChan {
			if stop {
				return
			}
			b.Pos.Add(offset)
			if b.NBT != nil {
				delete(b.NBT, "x")
				delete(b.NBT, "y")
				delete(b.NBT, "z")
			}
			outChan <- b
			<-ticker.C
		}
	}()
	return outChan, func() {
		stop = true
	}
}
