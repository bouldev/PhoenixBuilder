package main

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/lib/rental_server_impactor"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/remote_omega/transfer"
)

func main() {
	authServer := "wss://api.fastbuilder.pro:2053/"
	fmt.Println("Reading Info...")
	username, userPassword, userToken, serverCode, serverPassword, err := utils.ReadUserInfo("", "", "", "", "")
	if err != nil {
		panic(err)
	}

	accessOption := rental_server_impactor.DefaultOptions()
	accessOption.AuthServer = authServer
	accessOption.FBUsername = username
	accessOption.FBUserPassword = userPassword
	accessOption.FBUserToken = userToken
	accessOption.ServerCode = serverCode
	accessOption.ServerPassword = serverPassword
	accessOption.MakeBotCreative = true
	accessOption.DisableCommandBlock = false

	transferHandler, err := transfer.NewTransferAccessPoint(
		context.Background(),
		transfer.DefaultPubSubAccessPoint,
		transfer.DefaultCtrlAccessPoint,
		nil,
	)
	if err != nil {
		panic(err)
	}
	directPubMode := transfer.DefaultDirectPubSubModeEnable
	if directPubMode {
		accessOption.ReadLoopFunction = func(conn *minecraft.Conn, deadReason chan<- error, omega omega.ReactCore) {
			for {
				pkt, pktDataShared, err := conn.ReadPacketAndBytes()
				pktData := make([]byte, len(pktDataShared))
				copy(pktData, pktDataShared)
				if err != nil {
					deadReason <- fmt.Errorf("%v: %v", rental_server_impactor.ErrRentalServerDisconnected, err)
				}
				err = transferHandler.PubGamePacketData(pktData)
				if err != nil {
					deadReason <- fmt.Errorf("fail to remote dispatch packets")
				}
				omega.HandlePacket(pkt)
			}
		}
	}

	conn, omegaCore, deadReason, err := rental_server_impactor.ImpactServer(nil, accessOption)
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
