package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/function"
	I18n "phoenixbuilder/fastbuilder/i18n"
	fbauth "phoenixbuilder/fastbuilder/pv4"
	"phoenixbuilder/fastbuilder/py_rpc"
	"phoenixbuilder/fastbuilder/readline"
	"phoenixbuilder/fastbuilder/signalhandler"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/types"
	GameInterface "phoenixbuilder/game_control/game_interface"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/io/assembler"
	"phoenixbuilder/mirror/io/global"
	"phoenixbuilder/mirror/io/lru"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

func EnterReadlineThread(env *environment.PBEnvironment, breaker chan struct{}) {
	if args.NoReadline {
		return
	}
	defer Fatal()
	gameInterface := env.GameInterface
	functionHolder := env.FunctionHolder.(*function.FunctionHolder)
	for {
		if breaker != nil {
			select {
			case <-breaker:
				return
			default:
			}
		}
		cmd := readline.Readline(env)
		if len(cmd) == 0 {
			continue
		}
		if cmd[0] == '.' {
			resp := gameInterface.SendCommandWithResponse(
				cmd[1:],
				ResourcesControl.CommandRequestOptions{
					TimeOut: ResourcesControl.CommandRequestDefaultDeadLine,
				},
			)
			if resp.Error != nil {
				env.GameInterface.Output(
					pterm.Error.Sprintf(
						"Failed to get respond of \"%v\", and the following is the error log.",
						cmd[1:],
					),
				)
				env.GameInterface.Output(pterm.Error.Sprintf("%v", resp.Error.Error()))
			} else {
				fmt.Printf("%+v\n", resp.Respond)
			}
		} else if cmd[0] == '!' {
			resp := gameInterface.SendWSCommandWithResponse(
				cmd[1:],
				ResourcesControl.CommandRequestOptions{
					TimeOut: ResourcesControl.CommandRequestDefaultDeadLine,
				},
			)
			if resp.Error != nil {
				env.GameInterface.Output(
					pterm.Error.Sprintf(
						"Failed to get respond of \"%v\", and the following is the error log.",
						cmd[1:],
					),
				)
				env.GameInterface.Output(pterm.Error.Sprintf("%v", resp.Error.Error()))
			} else {
				fmt.Printf("%+v\n", resp.Respond)
			}
		} else if cmd[0] == '*' {
			gameInterface.SendSettingsCommand(cmd[1:], false)
		}
		functionHolder.Process(cmd)
	}
}

func onPyRpc(p *packet.PyRpc, env *environment.PBEnvironment) {
	conn := env.Connection.(*minecraft.Conn)
	if p.Value == nil {
		return
	}
	go_p_val := p.Value.MakeGo()
	/*
		json_val, _:=json.MarshalIndent(go_p_val, "", "\t")
		fmt.Printf("Received PyRpc: %s\n", json_val)
	*/
	if go_p_val == nil {
		return
	}
	pyrpc_val, ok := go_p_val.([]interface{})
	if !ok || len(pyrpc_val) < 2 {
		return
	}
	command, ok := pyrpc_val[0].(string)
	if !ok {
		return
	}
	data, ok := pyrpc_val[1].([]interface{})
	if !ok {
		return
	}
	switch command {
	case "S2CHeartBeat":
		conn.WritePacket(&packet.PyRpc{
			Value: py_rpc.FromGo([]interface{}{
				"C2SHeartBeat",
				data,
				nil,
			}),
		})
	case "GetStartType":
		client := env.FBAuthClient.(*fbauth.Client)
		response := client.TransferData(data[0].(string))
		conn.WritePacket(&packet.PyRpc{
			Value: py_rpc.FromGo([]interface{}{
				"SetStartType",
				[]interface{}{response},
				nil,
			}),
		})
	case "GetMCPCheckNum":
		if env.GetCheckNumEverPassed {
			break
		}
		firstArg := data[0].(string)
		secondArg := (data[1].([]interface{}))[0].(string)
		client := env.FBAuthClient.(*fbauth.Client)
		arg, _ := json.Marshal([]interface{}{firstArg, secondArg, env.Connection.(*minecraft.Conn).GameData().EntityUniqueID})
		ret := client.TransferCheckNum(string(arg))
		ret_p := []interface{}{}
		json.Unmarshal([]byte(ret), &ret_p)
		conn.WritePacket(&packet.PyRpc{
			Value: py_rpc.FromGo([]interface{}{
				"SetMCPCheckNum",
				[]interface{}{
					ret_p,
				},
				nil,
			}),
		})
		env.GetCheckNumEverPassed = true
	}
}

func EnterWorkerThread(env *environment.PBEnvironment, breaker chan struct{}) {
	conn := env.Connection.(*minecraft.Conn)
	functionHolder := env.FunctionHolder.(*function.FunctionHolder)

	chunkAssembler := assembler.NewAssembler(assembler.REQUEST_AGGRESSIVE, time.Second*5)
	// max 100 chunk requests per second
	chunkAssembler.CreateRequestScheduler(func(pk *packet.SubChunkRequest) {
		conn.WritePacket(pk)
	})

	for {
		if breaker != nil {
			select {
			case <-breaker:
				return
			default:
			}
		}

		pk, err := conn.ReadPacket()
		if err != nil {
			panic(err)
		}

		env.ResourcesUpdater.(func(*packet.Packet))(&pk)

		switch p := pk.(type) {
		case *packet.PyRpc:
			onPyRpc(p, env)
		case *packet.Text:
			if p.TextType == packet.TextTypeChat {
				if args.InGameResponse {
					if p.SourceName == env.RespondTo {
						functionHolder.Process(p.Message)
					}
				}
				break
			}
		case *packet.ActorEvent:
			if p.EventType == packet.ActorEventDeath && p.EntityRuntimeID == conn.GameData().EntityRuntimeID {
				conn.WritePacket(&packet.PlayerAction{
					EntityRuntimeID: conn.GameData().EntityRuntimeID,
					ActionType:      protocol.PlayerActionRespawn,
				})
			}
		case *packet.SubChunk:
			chunkData := chunkAssembler.OnNewSubChunk(p)
			if chunkData != nil {
				env.ChunkFeeder.(*global.ChunkFeeder).OnNewChunk(chunkData)
				env.LRUMemoryChunkCacher.(*lru.LRUMemoryChunkCacher).Write(chunkData)
			}
		case *packet.NetworkChunkPublisherUpdate:
			// pterm.Info.Println("packet.NetworkChunkPublisherUpdate", p)
			// missHash := []uint64{}
			// hitHash := []uint64{}
			// for i := uint64(0); i < 64; i++ {
			// 	missHash = append(missHash, uint64(10184224921554030005+i))
			// 	hitHash = append(hitHash, uint64(6346766690299427078-i))
			// }
			// conn.WritePacket(&packet.ClientCacheBlobStatus{
			// 	MissHashes: missHash,
			// 	HitHashes:  hitHash,
			// })
		case *packet.LevelChunk:
			// pterm.Info.Println("LevelChunk", p.BlobHashes, len(p.BlobHashes), p.CacheEnabled)
			// go func() {
			// 	for {

			// conn.WritePacket(&packet.ClientCacheBlobStatus{
			// 	MissHashes: []uint64{p.BlobHashes[0] + 1},
			// 	HitHashes:  []uint64{},
			// })
			// 		time.Sleep(100 * time.Millisecond)
			// 	}
			// }()
			if fbtask.CheckHasWorkingTask(env) {
				break
			}
			if exist := chunkAssembler.AddPendingTask(p); !exist {
				requests := chunkAssembler.GenRequestFromLevelChunk(p)
				chunkAssembler.ScheduleRequest(requests)
			}
		}
	}
}

func InitializeMinecraftConnection(ctx context.Context, authenticator minecraft.Authenticator) (conn *minecraft.Conn, err error) {
	if args.DebugMode {
		conn = &minecraft.Conn{
			DebugMode: true,
		}
	} else {
		dialer := minecraft.Dialer{
			Authenticator: authenticator,
		}
		conn, err = dialer.DialContext(ctx, "raknet")
	}
	if err != nil {
		return
	}
	conn.WritePacket(&packet.ClientCacheStatus{
		Enabled: false,
	})
	runtimeid := fmt.Sprintf("%d", conn.GameData().EntityUniqueID)
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"SyncUsingMod",
			[]interface{}{},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"SyncVipSkinUuid",
			[]interface{}{nil},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"ClientLoadAddonsFinishedFromGac",
			[]interface{}{},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"ModEventC2S",
			[]interface{}{
				"Minecraft",
				"preset",
				"GetLoadedInstances",
				map[string]interface{}{
					"playerId": runtimeid,
				},
			},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"arenaGamePlayerFinishLoad",
			[]interface{}{},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{}{
			"ModEventC2S",
			[]interface{}{
				"Minecraft",
				"vipEventSystem",
				"PlayerUiInit",
				runtimeid,
			},
			nil,
		}),
	})
	return
}

func EstablishConnectionAndInitEnv(env *environment.PBEnvironment) {
	if env.FBAuthClient == nil {
		env.ClientOptions.AuthServer = args.AuthServer
		env.ClientOptions.RespondUserOverride = args.CustomGameName
		env.FBAuthClient = fbauth.CreateClient(env.ClientOptions)
	}
	pterm.Println(pterm.Yellow(fmt.Sprintf("%s: %s", I18n.T(I18n.ServerCodeTrans), env.LoginInfo.ServerCode)))

	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	authenticator := fbauth.NewAccessWrapper(
		env.FBAuthClient.(*fbauth.Client),
		env.LoginInfo.ServerCode,
		env.LoginInfo.ServerPasscode,
		env.LoginInfo.Token,
		env.LoginInfo.Username,
		env.LoginInfo.Password,
	)
	conn, err := InitializeMinecraftConnection(ctx, authenticator)

	if err != nil {
		pterm.Error.Println(err)
		if runtime.GOOS == "windows" {
			pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		panic(err)
	}
	if len(env.RespondTo) == 0 {
		if args.CustomGameName != "" {
			env.RespondTo = args.CustomGameName
		} else {
			env.RespondTo = env.FBAuthClient.(*fbauth.Client).RespondTo
		}
	}

	env.Connection = conn
	pterm.Println(pterm.Yellow(I18n.T(I18n.ConnectionEstablished)))

	env.Resources = &ResourcesControl.Resources{}
	env.ResourcesUpdater = env.Resources.(*ResourcesControl.Resources).Init()
	env.GameInterface = &GameInterface.GameInterface{
		WritePacket: env.Connection.(*minecraft.Conn).WritePacket,
		ClientInfo: GameInterface.ClientInfo{
			DisplayName:     env.Connection.(*minecraft.Conn).IdentityData().DisplayName,
			ClientIdentity:  env.Connection.(*minecraft.Conn).IdentityData().Identity,
			XUID:            env.Connection.(*minecraft.Conn).IdentityData().XUID,
			EntityRuntimeID: env.Connection.(*minecraft.Conn).GameData().EntityRuntimeID,
			EntityUniqueID:  env.Connection.(*minecraft.Conn).GameData().EntityUniqueID,
		},
		Resources: env.Resources.(*ResourcesControl.Resources),
	}
	functionHolder := env.FunctionHolder.(*function.FunctionHolder)
	function.InitPresetFunctions(functionHolder)
	fbtask.InitTaskStatusDisplay(env)

	signalhandler.Install(conn, env)

	taskholder := env.TaskHolder.(*fbtask.TaskHolder)
	types.ForwardedBrokSender = taskholder.BrokSender
}

func getUserInputMD5() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("MD5: ")
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(code, "\r\n"), err
}
