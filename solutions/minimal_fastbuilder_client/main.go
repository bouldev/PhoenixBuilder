package main

import (
	"fmt"
	"phoenixbuilder/lib/helpers/fb_enter_server"
	"phoenixbuilder/lib/helpers/fbuser"
	"phoenixbuilder/lib/minecraft/neomega/omega"
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

	var deadReason chan error
	var omegaCore omega.MicroOmega
	omegaCore, deadReason, err = fb_enter_server.AccessServer(nil, accessOption)
	if err != nil {
		panic(err)
	}
	go func() {
		err = <-deadReason
		panic(err)
	}()
	fmt.Println(omegaCore)
	fmt.Println("Bot ok and now exit")
}
