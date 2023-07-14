package main

import (
	"bufio"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/fastbuilder/lib/rental_server_impact/access_helper"
	"phoenixbuilder/fastbuilder/lib/rental_server_impact/info_collect_utils"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

func main() {
	authServer := "https://api.fastbuilder.pro"
	username, userPassword, userToken, serverCode, serverPassword, err := info_collect_utils.ReadUserInfo("", "", "", "", "")
	if err != nil {
		panic(err)
	}

	accessOption := access_helper.DefaultOptions()
	accessOption.AuthServer = authServer
	accessOption.FBUsername = username
	accessOption.FBUserPassword = userPassword
	accessOption.FBUserToken = userToken
	accessOption.ServerCode = serverCode
	accessOption.ServerPassword = serverPassword
	accessOption.MakeBotCreative = true
	accessOption.DisableCommandBlock = false
	accessOption.ReasonWithPrivilegeStuff = true

	var deadReason chan error
	var omegaCore omega.MicroOmega
	_, omegaCore, deadReason, err = access_helper.ImpactServer(nil, accessOption)
	if err != nil {
		panic(err)
	}
	go func() {
		err = <-deadReason
		panic(err)
	}()
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
