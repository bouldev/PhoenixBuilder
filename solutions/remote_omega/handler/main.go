package main

import (
	"bufio"
	"fmt"
	"os"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/remote_omega/transfer"
	"time"
)

func main() {
	omegaCore, err := transfer.ConnectToAccessPoint(nil, nil, nil)
	if err != nil {
		panic(err)
	}
	resultWaitor := make(chan *packet.CommandOutput, 1)
	firstTime := true
	startTime := time.Now()
	go func() {
		for {
			omegaCore.GetGameControl().SendWSCmdAndInvokeOnResponse("testforblock ~~~ air", func(output *packet.CommandOutput) {
				resultWaitor <- output
			})
			select {
			case r := <-resultWaitor:
				if firstTime {
					fmt.Println(r)
					firstTime = false
				}
				fmt.Printf("\ralive %v", time.Since(startTime))
				time.Sleep(time.Second / 10)
			case <-time.NewTimer(time.Second * 3).C:
				panic(fmt.Errorf("no response after 3 second, bot is down (alive %v)", time.Since(startTime)))
			}
		}
	}()
	bufio.NewReader(os.Stdin).ReadByte()
}
