package transfer

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/bundle"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
)

func ConnectToAccessPoint(ctx context.Context, endPointOption *EndPointOption, omegaOption *bundle.MicroOmegaOption) (omegaCore omega.MicroOmega, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if endPointOption == nil {
		endPointOption = MakeDefaultEndPointOption()
	}
	if omegaOption == nil {
		omegaOption = bundle.MakeDefaultMicroOmegaOption()
	}
	endPoint, err := NewEndPoint(
		ctx,
		endPointOption.PubAccessPoint,
		endPointOption.CtrlAccessPoint,
	)
	if err != nil {
		return nil, err
	}
	endPoint.WaitReady()
	fmt.Println("connected to access point")
	var interactCore omega.InteractCore
	if endPointOption.DirectSendMode {
		interactCore = core.NewRemoteInteractCoreDirect(endPoint.SendPacketData)
		interactCore.(*core.RemoteInteractCoreDirect).UpdateShieldID(endPoint.GetShieldID())
	} else {
		interactCore = core.NewRemoteInteractCore(endPoint.SendPacket)
	}
	omegaCore = bundle.NewMicroOmega(interactCore, endPoint.GetUQHolder, omegaOption)

	go func() {
		var err error
		var pkt packet.Packet
		var shieldID int32
		var recvFn func() (pk packet.Packet, shieldID int32, err error)
		if endPointOption.DirectSubMode {
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
			if endPointOption.DirectSendMode {
				interactCore.(*core.RemoteInteractCoreDirect).UpdateShieldID(shieldID)
			}
			omegaCore.HandlePacket(pkt)
		}
	}()
	return omegaCore, nil
}
