package main

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/core"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/lib/fbauth"
	"phoenixbuilder/lib/minecraft/neomega/bundle"
	"phoenixbuilder/lib/minecraft/neomega/decouple/cmdsender"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

var errStrFailToConnectAuthServer = "无法连接到登陆服务器"
var errFBUserCenterLoginFail = "无效的 Fastbuilder 用户名或密码"
var errCannotConnectToRentalServer = "无法连接到租赁服, 可能是用户 FBToken 无效、无租赁服登陆权限、租赁服未开放"
var errRentalServerDisconnected = "与租赁服的连接已断开"

func WrapAuthenticator(connectContext context.Context, client *fbauth.Client, userName, userPassword, userToken, serverCode, serverPassword string) (authenticator *fbauth.AccessWrapper, writeBackToken string, err error) {
	if userToken == "" {
		authenticator, writeBackToken, err = fbauth.NewAccessWrapperByPassword(connectContext, client, userName, userPassword)
		if err != nil {
			return nil, "", fmt.Errorf("%v: %v", errFBUserCenterLoginFail, err)
		}
	} else {
		authenticator = fbauth.NewAccessWrapper(client, userToken)
	}
	authenticator.SetServerInfo(serverCode, serverPassword)
	return authenticator, writeBackToken, nil
}

func ReadInfo() (userName, userPassword, userToken, serverCode, serverPassword string, err error) {
	// read token or get user input
	I18n.Init()
	if userName == "" && userPassword == "" && userToken == "" {
		userToken, err = ReadToken(LoadTokenPath())
		if err != nil || userToken == "" {
			for userName == "" {
				userName, err = GetUserInput(I18n.T(I18n.Enter_FBUC_Username))
				if err != nil {
					return "", "", "", "", "", err
				}
			}
			for userPassword == "" {
				userPassword, err = GetUserPasswordInput(I18n.T(I18n.EnterPasswordForFBUC))
				if err != nil {
					return "", "", "", "", "", err
				}
			}
		}
	}

	// read server code and password
	if serverCode == "" {
		serverCode, serverPassword, err = GetRentalServerCode()
		if err != nil {
			return "", "", "", "", "", err
		}
	}
	return userName, userPassword, userToken, serverCode, serverPassword, nil
}

func main() {
	authServer := "wss://api.fastbuilder.pro:2053/"
	fmt.Println("Connecting to FB Server...")
	ctx := context.Background()
	fbClient := fbauth.NewClient(ctx)
	{
		connectCtx, _ := context.WithTimeout(ctx, 30*time.Second)
		err := fbClient.EstablishConnectionToAuthServer(connectCtx, authServer)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Reading Info...")
	userName, userPassword, userToken, serverCode, serverPassword, err := ReadInfo()
	if err != nil {
		panic(err)
	}

	fmt.Println("Authenticating...")
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)
	authenticator, writeBackToken, err := WrapAuthenticator(ctx, fbClient, userName, userPassword, userToken, serverCode, serverPassword)
	if err != nil {
		panic(err)
	}
	if writeBackToken != "" {
		WriteToken(writeBackToken, LoadTokenPath())
	}

	fmt.Printf("Connecting to MC Server: (Code:%v, Password:%v)\n", authenticator.ServerCode, authenticator.ServerPassword)
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)
	client, err := core.InitMCConnection(ctx, authenticator)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully Connected to MC Server!\n")
	var pkt packet.Packet
	omega := bundle.NewMicroOmega(client, bundle.MicroOmegaOption{
		CmdSenderOptions: cmdsender.Options{
			ExpectedCmdFeedBack: false,
		},
		PrintUQHolderDebugInfo: false,
	})
	fmt.Printf("Adding Omega Components...\n")
	NewPyRPCResponser(omega, authenticator.GetFBUid(),
		func(content, uid string) string {
			ctx, _ = context.WithTimeout(ctx, 30*time.Second)
			data, err := authenticator.TransferData(ctx, content, uid)
			if err != nil {
				panic(err)
			}
			return data
		},
		func(firstArg, secondArg string) (valM string, valS string) {
			ctx, _ = context.WithTimeout(ctx, 30*time.Second)
			valM, valS, err = authenticator.TransferCheckNum(ctx, firstArg, secondArg)
			if err != nil {
				panic(err)
			}
			return
		},
	)
	helper := NewSetupHelper(omega)
	fmt.Printf("Running Omega...\n")
	go func() {
		for {
			pkt, err = client.ReadPacket()
			if err != nil {
				panic(fmt.Errorf("%v: %v", errRentalServerDisconnected, err))
			}
			omega.HandlePacket(pkt)
		}
	}()

	helper.WaitOK()
	fmt.Println("Minimal Client exited.")
}
