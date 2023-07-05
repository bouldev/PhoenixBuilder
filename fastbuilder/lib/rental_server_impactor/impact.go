package rental_server_impactor

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/core"
	"phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/lib/rental_server_impactor/challenges"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/bundle"
	neomega_core "phoenixbuilder/fastbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/uqholder"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
)

func ImpactServer(ctx context.Context, options *Options) (conn *minecraft.Conn, omegaCore *bundle.MicroOmega, deadReason chan error, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if options.MaximumWaitTime > 0 {
		ctx, _ = context.WithTimeout(ctx, options.MaximumWaitTime)
	}
	env:=&environment.PBEnvironment{
		AuthServer: options.AuthServer,
	}
	fbClient:=fbauth.CreateClient(env)
	if options.FBUserToken == "" {
		var err_val string
		options.FBUserToken, err_val = fbClient.GetToken(options.FBUsername, options.FBUserPassword)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%v: %s", ErrFBUserCenterLoginFail, err_val)
		}
		if options.WriteBackToken {
			utils.WriteFBToken(options.FBUserToken, utils.LoadTokenPath())
		}
	}
	authenticator := fbauth.NewAccessWrapper(fbClient, options.ServerCode, options.ServerPassword, options.FBUserToken)
	{
		connectMCServer := func() (conn *minecraft.Conn, err error) {
			connectCtx := ctx
			if options.ServerConnectionTimeout != 0 {
				connectCtx, _ = context.WithTimeout(ctx, options.ServerConnectionTimeout)
			}
			conn, err = core.InitializeMinecraftConnection(connectCtx, authenticator)
			if err != nil {
				if connectCtx.Err() != nil {
					return nil, ErrRentalServerConnectionTimeOut
				}
				return nil, fmt.Errorf("%v :%v", ErrFailToConnectRentalServer, err)
			}
			return conn, nil
		}

		for {
			conn, err = connectMCServer()
			if err == nil {
				break
			} else {
				fmt.Println(err)
			}
			if options.ServerConnectRetryTimes <= 0 {
				break
			}
			options.ServerConnectRetryTimes--
		}
		if err != nil {
			return nil, nil, nil, err
		}
	}
	omegaCore = bundle.NewMicroOmega(neomega_core.NewInteractCore(conn), func() omega.MicroUQHolder {
		return uqholder.NewMicroUQHolder(conn)
	}, options.MicroOmegaOption)
	deadReason = make(chan error, 0)
	challengeSolver := challenges.NewPyRPCResponder(omegaCore, env.Uid,
		fbClient.TransferData,
		fbClient.TransferCheckNum,
	)
	go func() {
		options.ReadLoopFunction(conn, deadReason, omegaCore)
	}()
	if options.ReasonWithPrivilegeStuff {
		helper := challenges.NewOperatorChallenge(omegaCore, func() {
			if options.OpPrivilegeRemovedCallBack != nil {
				options.OpPrivilegeRemovedCallBack()
			}
			if options.DieOnLosingOpPrivilege {
				deadReason <- ErrBotOpPrivilegeRemoved
			}
		})
		waitErr := make(chan error)
		go func() {
			waitErr <- helper.WaitForPrivilege(ctx, challengeSolver.ChallengeCompete)
		}()
		select {
		case err = <-waitErr:
		case err = <-deadReason:
		}
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if options.MakeBotCreative {
		omegaCore.GetGameControl().SendPlayerCmdAndInvokeOnResponseWithFeedback("gamemode c @s", func(output *packet.CommandOutput) {
		})
	}
	if options.DisableCommandBlock {
		omegaCore.GetGameControl().SendPlayerCmdAndInvokeOnResponseWithFeedback("gamerule commandblocksenabled false", func(output *packet.CommandOutput) {
		})
	}
	return conn, omegaCore, deadReason, nil
}
