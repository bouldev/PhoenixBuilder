package main

import (
	"bufio"
	"fmt"
	"os"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/remote_omega/transfer"
)

func main() {
	omegaCore, err := transfer.ConnectToAccessPoint(nil, nil, nil)
	if err != nil {
		panic(err)
	}
	omegaCore.GetGameControl().SendWSCmdAndInvokeOnResponse("testforblock ~~~ air", func(output *packet.CommandOutput) {
		fmt.Println(output)
	})
	bufio.NewReader(os.Stdin).ReadByte()
}
