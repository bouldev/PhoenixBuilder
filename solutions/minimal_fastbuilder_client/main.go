package main

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/core"
	"phoenixbuilder/lib/fbauth"
	"phoenixbuilder/lib/minecraft/neomega/bundle"
	"phoenixbuilder/lib/minecraft/neomega/decouple/cmdsender"
	"phoenixbuilder/minecraft/protocol/packet"
)

var errStrFailToConnectAuthServer = "无法连接到登陆服务器"
var errFBUserCenterLoginFail = "无效的 Fastbuilder 用户名或密码"
var errCannotConnectToRentalServer = "无法连接到租赁服, 可能是用户 FBToken 无效、无租赁服登陆权限、租赁服未开放"
var errRentalServerDisconnected = "与租赁服的连接已断开"

func WrapAuthenticator(ctx context.Context, authServer, userName, userPassword, userToken, serverCode, serverPassword string) (authenticator *fbauth.AccessWrapper, err error) {
	client := fbauth.NewClient(ctx)
	err = client.EstablishConnectionToAuthServer(authServer)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", errStrFailToConnectAuthServer, err)
	}
	if userToken == "" {
		authenticator, err = fbauth.NewAccessWrapperByPassword(client, userName, userPassword)
		if err != nil {
			return nil, fmt.Errorf("%v: %v", errFBUserCenterLoginFail, err)
		}
	} else {
		authenticator = fbauth.NewAccessWrapper(client, userToken)
	}
	authenticator.SetServerInfo(serverCode, serverPassword)
	return authenticator, nil
}

func main() {
	authServer := "wss://api.fastbuilder.pro:2053/"
	userName := "userName"
	userPassword := "userPassword"
	userToken := ""
	serverCode := "serverCode"
	serverPassword := ""

	ctx := context.Background()

	fmt.Println("Authenticating...")
	authenticator, err := WrapAuthenticator(ctx, authServer, userName, userPassword, userToken, serverCode, serverPassword)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Server: (Code:%v, Password:%v)\n", authenticator.ServerCode, authenticator.ServerPassword)

	client, err := core.InitMCConnection(ctx, authenticator)
	if err != nil {
		panic(err)
	}
	var pkt packet.Packet
	omega := bundle.NewMicroOmega(client, bundle.MicroOmegaOption{
		CmdSenderOptions: cmdsender.Options{
			ExpectedCmdFeedBack: false,
		},
		PrintUQHolderDebugInfo: false,
	})
	go func() {
		for {
			pkt, err = client.ReadPacket()
			if err != nil {
				panic(fmt.Errorf("%v: %v", errRentalServerDisconnected, err))
			}
			omega.HandlePacket(pkt)
		}
	}()

	helper := NewSetupHelper(omega)
	helper.WaitOK()
	fmt.Println("演示程序执行完毕")
}
