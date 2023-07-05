package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"phoenixbuilder/lib/minecraft/neomega/bundle"
	"phoenixbuilder/lib/minecraft/neomega/decouple/cmdsender"
	"phoenixbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/remote_omega/transfer"
)

func main() {
	endPoint, err := transfer.NewEndPoint(
		context.Background(),
		"tcp://*:24016",
		"tcp://*:24015",
	)
	if err != nil {
		panic(err)
	}
	endPoint.WaitReady()
	fmt.Println("access point ready")
	var interactCore omega.InteractCore
	directSendMode := true
	directSubMode := false
	if directSendMode {
		interactCore = core.NewRemoteInteractCoreDirect(endPoint.SendPacketData)
		interactCore.(*core.RemoteInteractCoreDirect).UpdateShieldID(endPoint.GetShieldID())
	} else {
		interactCore = core.NewRemoteInteractCore(endPoint.SendPacket)
	}
	option := bundle.MicroOmegaOption{CmdSenderOptions: cmdsender.Options{ExpectedCmdFeedBack: false}}
	omegaCore := bundle.NewMicroOmega(interactCore, endPoint.GetUQHolder, option)

	go func() {
		var err error
		var pkt packet.Packet
		var shieldID int32
		var recvFn func() (pk packet.Packet, shieldID int32, err error)
		if directSubMode {
			recvFn = endPoint.RecvDirectPacket
		} else {
			recvFn = endPoint.RecvPacket
		}
		for {
			pkt, shieldID, err = recvFn()
			if err != nil {
				fmt.Println(err)
				continue
			}
			if directSendMode {
				interactCore.(*core.RemoteInteractCoreDirect).UpdateShieldID(shieldID)
			}
			omegaCore.HandlePacket(pkt)
		}

	}()
	omegaCore.SendWSCmdAndInvokeOnResponse("testforblock ~~~ air", func(output *packet.CommandOutput) {
		fmt.Println(output)
	})
	bufio.NewReader(os.Stdin).ReadByte()
}
