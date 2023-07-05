package main

import (
	"fmt"
	"phoenixbuilder/fastbuilder/lib/rental_server_impactor"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
)

func main() {
	authServer := "wss://api.fastbuilder.pro:2053/"
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
	accessOption.ReasonWithPrivilegeStuff = false

	var deadReason chan error
	var omegaCore omega.MicroOmega
	_, omegaCore, deadReason, err = rental_server_impactor.ImpactServer(nil, accessOption)
	if err != nil {
		panic(err)
	}
	go func() {
		err = <-deadReason
		panic(err)
	}()
	fmt.Println(omegaCore)
}
