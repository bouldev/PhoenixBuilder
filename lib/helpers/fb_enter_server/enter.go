package fb_enter_server

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/core"
	"phoenixbuilder/lib/fbauth"
	"phoenixbuilder/lib/helpers/bot_privilege"
	"phoenixbuilder/lib/helpers/fbuser"
	"phoenixbuilder/lib/minecraft/neomega/bundle"
	"phoenixbuilder/lib/minecraft/neomega/decouple/cmdsender"
	neomega_core "phoenixbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/lib/minecraft/neomega/uqholder"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
)

func AccessServer(ctx context.Context, options *Options) (omegaCore *bundle.MicroOmega, deadReason chan error, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if options.MaximumWaitTime > 0 {
		ctx, _ = context.WithTimeout(ctx, options.MaximumWaitTime)
	}

	fmt.Println("正在与FB服务器建立连接...")
	fbClient := fbauth.NewClient(ctx)
	{
		connectCtx := ctx
		if options.FBServerConnectionTimeOut != 0 {
			connectCtx, _ = context.WithTimeout(ctx, options.FBServerConnectionTimeOut)
		}
		err = fbClient.EstablishConnectionToAuthServer(connectCtx, options.AuthServer)
		if err != nil {
			if connectCtx.Err() != nil {
				return nil, nil, ErrFBServerConnectionTimeOut
			}
			return nil, nil, fmt.Errorf("%v :%v", ErrFailToConnectFBServer, err)
		}
	}
	if options.FBUserToken == "" {
		fmt.Println("正在登陆FB服务器并取得Token...")
		connectCtx := ctx
		if options.FBGetTokenTimeOut != 0 {
			connectCtx, _ = context.WithTimeout(ctx, options.FBGetTokenTimeOut)
		}
		options.FBUserToken, err = fbauth.GetTokenByPassword(connectCtx, fbClient, options.FBUserName, options.FBUserPassword)
		if err != nil {
			if connectCtx.Err() != nil {
				return nil, nil, ErrGetTokenTimeOut
			}
			return nil, nil, fmt.Errorf("%v: %v", ErrFBUserCenterLoginFail, err)
		}
		if options.WriteBackToken {
			fbuser.WriteToken(options.FBUserToken, fbuser.LoadTokenPath())
		}
	}
	authenticator := fbauth.NewAccessWrapper(fbClient, options.FBUserToken)
	authenticator.SetServerInfo(options.ServerCode, options.ServerPassword)
	fmt.Printf("正在登陆网易租赁服(服号:%v)...\n", authenticator.ServerCode)
	var conn *minecraft.Conn
	{
		connectMCServer := func() (conn *minecraft.Conn, err error) {
			connectCtx := ctx
			if options.MCServerConnectionTimeOut != 0 {
				connectCtx, _ = context.WithTimeout(ctx, options.MCServerConnectionTimeOut)
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
			if options.MCServerConnectRetryTimes <= 0 {
				break
			}
			options.MCServerConnectRetryTimes--
			fmt.Println("连接失败，重试中...")
		}
		if err != nil {
			return nil, nil, err
		}
	}
	fmt.Println("检查和配置租赁服状态中...")
	omegaCore = bundle.NewMicroOmega(neomega_core.NewInteractCore(conn), func() omega.MicroUQHolder {
		return uqholder.NewMicroUQHolder(conn)
	}, bundle.MicroOmegaOption{
		CmdSenderOptions: cmdsender.Options{
			ExpectedCmdFeedBack: options.ExpectedCmdFeedBack,
		},
	})
	deadReason = make(chan error, 0)
	challengeSolver := bot_privilege.NewPyRPCResponser(omegaCore, authenticator.GetFBUid(), fbClient.Closed(),
		func(content, uid string) string {
			connectCtx := ctx
			if options.TransferTimeOut != 0 {
				connectCtx, _ = context.WithTimeout(ctx, options.TransferTimeOut)
			}
			data, err := authenticator.TransferData(connectCtx, content, uid)
			if err != nil {
				if connectCtx.Err() != nil {
					deadReason <- ErrFBTransferDataTimeOut
				} else {
					deadReason <- fmt.Errorf("%v: %v", ErrFBTransferDataFail, err)
				}
			}
			return data
		},
		func(firstArg, secondArg string, botEntityUniqueID int64) (valM, valS, valT string) {
			connectCtx := ctx
			if options.TransferTimeOut != 0 {
				connectCtx, _ = context.WithTimeout(ctx, options.TransferCheckNumTimeOut)
			}
			valM, valS, valT, err = authenticator.TransferCheckNum(connectCtx, firstArg, secondArg, botEntityUniqueID)
			if err != nil {
				if connectCtx.Err() != nil {
					deadReason <- ErrFBTransferCheckNumTimeOut
				} else {
					deadReason <- fmt.Errorf("%v: %v", ErrFBTransferCheckNumFail, err)
				}
			}
			return
		},
	)
	helper := bot_privilege.NewSetupHelper(omegaCore, func() {
		if options.OpPrivilegeRemovedCallBack != nil {
			options.OpPrivilegeRemovedCallBack()
		}
		if options.DeadOnOpPrivilegeRemoved {
			deadReason <- ErrBotOpPrivilegeRemoved
		}
	})
	go func() {
		options.ReadLoopFunction(conn, deadReason, omegaCore)
	}()
	err = helper.WaitOK(ctx, challengeSolver.ChallengeCompete)
	if err != nil {
		return nil, nil, err
	}
	if options.MakeBotCreative {
		omegaCore.GetGameControl().SendPlayerCmdAndInvokeOnResponseWithFeedback("gamemode c @s", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				fmt.Println("机器人已变更为创造模式")
			}
		})
	}
	if options.DisableCommandBlock {
		omegaCore.GetGameControl().SendPlayerCmdAndInvokeOnResponseWithFeedback("gamerule commandblocksenabled false", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				fmt.Println("为了提高性能及租赁服稳定性，命令块已关闭")
			}
		})
	}
	return omegaCore, deadReason, nil
}
