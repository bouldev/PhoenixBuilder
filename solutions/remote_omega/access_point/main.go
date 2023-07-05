package main

import (
	"context"
	"fmt"
	"phoenixbuilder/lib/helpers/fb_enter_server"
	"phoenixbuilder/lib/helpers/fbuser"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/remote_omega/transfer"
)

func main() {
	authServer := "wss://api.fastbuilder.pro:2053/"
	fmt.Println("Reading Info...")
	userName, userPassword, userToken, serverCode, serverPassword, err := fbuser.ReadInfo("", "", "", "", "")
	if err != nil {
		panic(err)
	}

	accessOption := fb_enter_server.MakeDefaultOption()
	accessOption.AuthServer = authServer
	accessOption.FBUserName = userName
	accessOption.FBUserPassword = userPassword
	accessOption.FBUserToken = userToken
	accessOption.ServerCode = serverCode
	accessOption.ServerPassword = serverPassword
	accessOption.MakeBotCreative = true
	accessOption.DisableCommandBlock = false
	accessOption.ExpectedCmdFeedBack = false

	transferHandler, err := transfer.NewTransferAccessPoint(
		context.Background(),
		"tcp://*:24016",
		"tcp://*:24015",
		nil,
	)
	directPubMode := false
	if directPubMode {
		accessOption.ReadLoopFunction = func(conn *minecraft.Conn, deadReason chan<- error, omega omega.ReactCore) {
			for {
				pkt, pktDataShared, err := conn.ReadPacketAndBytes()
				pktData := make([]byte, len(pktDataShared))
				copy(pktData, pktDataShared)
				if err != nil {
					deadReason <- fmt.Errorf("%v: %v", fb_enter_server.ErrRentalServerDisconnected, err)
				}
				err = transferHandler.PubGamePacketData(pktData, pkt.ID())
				if err != nil {
					deadReason <- fmt.Errorf("fail to remote dispatch packets")
				}
				omega.HandlePacket(pkt)
			}
		}
	}

	conn, omegaCore, deadReason, err := fb_enter_server.AccessServer(nil, accessOption)
	if err != nil {
		panic(err)
	}
	transferHandler.UpdateGetShieldIDFn(conn.GetShieldID)
	if !directPubMode {
		omegaCore.GetGameListener().SetOnAnyPacketCallBack(func(p packet.Packet) {
			transferHandler.PubGamePacket(p)
		})
	}
	go transferHandler.EnableCtrlTransfer(omegaCore)
	fmt.Println("access point ready")

	err = <-deadReason
	panic(err)
}
