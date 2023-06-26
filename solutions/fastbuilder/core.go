package fastbuilder

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/GameControl/GlobalAPI"
	"phoenixbuilder/GameControl/ResourcesControlCenter"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/core"
	fbauth "phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/external"
	"phoenixbuilder/fastbuilder/function"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/move"
	"phoenixbuilder/fastbuilder/readline"
	script_bridge "phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/signalhandler"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/io/commands"
	utils_core "phoenixbuilder/lib/utils/core"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/io/assembler"
	"phoenixbuilder/mirror/io/global"
	"phoenixbuilder/mirror/io/lru"
	"phoenixbuilder/omega/cli/embed"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

func EnterReadlineThread(env *environment.PBEnvironment, breaker chan struct{}) {
	if args.NoReadline() {
		return
	}
	defer Fatal()
	commandSender := env.CommandSender.(*commands.CommandSender)
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
		if env.OmegaAdaptorHolder != nil && !strings.Contains(cmd, "exit") {
			env.OmegaAdaptorHolder.(*embed.EmbeddedAdaptor).FeedBackendCommand(cmd)
			continue
		}
		if cmd[0] == '.' {
			resp, _ := env.GlobalAPI.(*GlobalAPI.GlobalAPI).SendCommandWithResponce(cmd[1:])
			fmt.Printf("%+v\n", resp)
		} else if cmd[0] == '!' {
			resp, _ := env.GlobalAPI.(*GlobalAPI.GlobalAPI).SendWSCommandWithResponce(cmd[1:])
			fmt.Printf("%+v\n", resp)
		}
		if cmd == "move" {
			go func() {
				for {
					move.Auto()
					time.Sleep(time.Second / 20)
				}
			}()
			continue
		}
		if cmd[0] == '>' && len(cmd) > 1 {
			umsg := cmd[1:]
			if env.FBAuthClient != nil {
				fbcl := env.FBAuthClient.(*fbauth.Client)
				if !fbcl.CanSendMessage() {
					commandSender.WorldChatOutput("FastBuildeｒ", "Lost connection to the authentication server.")
					break
				}
				fbcl.WorldChat(umsg)
			}
		}
		functionHolder.Process(cmd)
	}
}

func EnterWorkerThread(env *environment.PBEnvironment, breaker chan struct{}) {
	conn := env.Connection.(*minecraft.Conn)
	hostBridgeGamma := env.ScriptBridge.(*script_bridge.HostBridgeGamma)
	commandSender := env.CommandSender.(*commands.CommandSender)
	functionHolder := env.FunctionHolder.(*function.FunctionHolder)

	chunkAssembler := assembler.NewAssembler(assembler.REQUEST_AGGRESSIVE, time.Second*5)
	// max 100 chunk request per second
	chunkAssembler.CreateRequestScheduler(func(pk *packet.SubChunkRequest) {
		conn.WritePacket(pk)
	})
	getchecknum_everPassed := false
	// currentChunkConstructor := &world_provider.ChunkConstructor{}
	for {
		if breaker != nil {
			select {
			case <-breaker:
				return
			default:
			}
		}
		pk, data, err := conn.ReadPacketAndBytes()
		if err != nil {
			panic(err)
		}

		{
			p, ok := pk.(*packet.PyRpc)
			if ok {
				if strings.Contains(string(p.Content), "GetStartType") {
					// 2021-12-22 10:51~11:55
					// 2023-05-30
					// Thank netease for wasting my time again ;)
					//fmt.Printf("%X\n", p.Content)
					encData := p.Content[len(p.Content)-163 : len(p.Content)-1]
					//fmt.Printf("%s\n", p.Content)
					//fmt.Printf("%s\n", encData)
					//fmt.Printf("%s\n", env.Uid)
					client := env.FBAuthClient.(*fbauth.Client)
					response := client.TransferData(string(encData), fmt.Sprintf("%s", env.Uid))
					//fmt.Printf("%s\n", response)
					conn.WritePacket(&packet.PyRpc{
						Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xc, 0x53, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc4},
							[]byte{byte(len(response))},
							[]byte(response),
							[]byte{0xc0},
						}, []byte{}),
					})
					//fmt.Printf("%s\n", response)
				} else if strings.Contains(string(p.Content), "GetMCPCheckNum") {
					// This shit sucks, so as netease.
					if getchecknum_everPassed {
						continue
					}
					//fmt.Printf("%X", p.Content)
					firstArgLenB := p.Content[19:21]
					firstArgLen := binary.BigEndian.Uint16(firstArgLenB)
					firstArg := string(p.Content[21 : 21+firstArgLen])
					secondArgLen := uint16(p.Content[23+firstArgLen])
					secondArg := string(p.Content[24+firstArgLen : 24+firstArgLen+secondArgLen])
					//fmt.Printf("%s\n%s\n",firstArg, secondArg)
					//fmt.Printf("%v\n", env.Connection.(*minecraft.Conn).GameData().EntityUniqueID)
					//fmt.Printf("%X\n", p.Content)
					//valM,_:=getUserInputMD5()
					//valS,_:=getUserInputMD5()
					//valT,_:=getUserInputMD5()
					
					client := env.FBAuthClient.(*fbauth.Client)
					valM, valS, valT := client.TransferCheckNum(firstArg, secondArg, env.Connection.(*minecraft.Conn).GameData().EntityUniqueID)
					
					/*conn.WritePacket(&packet.PyRpc{
						Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xe, 0x53, 0x65, 0x74, 0x4d, 0x43, 0x50, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc4, 0x20},
							[]byte(valM),
							[]byte{0xc0},
						}, []byte{}),
					})*/
					conn.WritePacket(&packet.PyRpc{
						Content: bytes.Join([][]byte{[]byte{0x93, 0xc4, 0x0e}, []byte("SetMCPCheckNum"), []byte{0x91, 0x98, 0xc4, 0x20},
							[]byte(valM),
							[]byte{0xc4, 0x20},
							[]byte(valS),
							[]byte{0xc2},
							[]byte{0x90},
							[]byte{0xc4, 0x00},
							[]byte{0xc4, 0x00},
							[]byte{3},
							[]byte{0xc4,0x20},
							[]byte(valT),
							[]byte{0xC0},
						}, []byte{}),
					})
					getchecknum_everPassed = true
					/*go func() {
						time.Sleep(3*time.Second)
						resp, _ := env.GlobalAPI.(*GlobalAPI.GlobalAPI).SendCommandWithResponce("list")
						fmt.Printf("%+v\n", resp)
					} ()*/
				} else {
					//fmt.Printf("PyRpc! %s\n", p.Content)
				}
			}
		}

		if env.OmegaAdaptorHolder != nil {
			env.OmegaAdaptorHolder.(*embed.EmbeddedAdaptor).FeedPacketAndByte(pk, data)
			continue
		}

		go env.ResourcesUpdater.(func(pk *packet.Packet))(&pk)

		env.UQHolder.(*uqHolder.UQHolder).Update(pk)
		hostBridgeGamma.HostPumpMcPacket(pk)
		hostBridgeGamma.HostQueryExpose["uqHolder"] = func() string {
			marshal, err := json.Marshal(env.UQHolder.(*uqHolder.UQHolder))
			if err != nil {
				marshalErr, _ := json.Marshal(map[string]string{"err": err.Error()})
				return string(marshalErr)
			}
			return string(marshal)
		}
		if env.ExternalConnectionHandler != nil {
			env.ExternalConnectionHandler.(*external.ExternalConnectionHandler).PacketChannel <- data
		}
		// fmt.Println(omega_utils.PktIDInvMapping[int(pk.ID())])
		switch p := pk.(type) {
		// case *packet.AdventureSettings:
		// 	if conn.GameData().EntityUniqueID == p.PlayerUniqueID {
		// 		if p.PermissionLevel >= packet.PermissionLevelOperator {
		// 			opPrivilegeGranted = true
		// 		} else {
		// 			opPrivilegeGranted = false
		// 		}
		// 	}
		// case *packet.ClientCacheMissResponse:
		// 	pterm.Info.Println("ClientCacheMissResponse", p)
		// case *packet.ClientCacheStatus:
		// 	pterm.Info.Println("ClientCacheStatus", p)
		// case *packet.ClientCacheBlobStatus:
		// 	pterm.Info.Println("ClientCacheBlobStatus", p)
		case *packet.Text:
			if p.TextType == packet.TextTypeChat {
				if args.InGameResponse() {
					if p.SourceName == env.RespondUser {
						functionHolder.Process(p.Message)
					}
				}
				break
			}
		case *packet.CommandOutput:
			if p.CommandOrigin.UUID.String() == configuration.ZeroId.String() {
				pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
				if !(p.OutputMessages[0].Message == "commands.generic.unknown") {
					configuration.IsOp = true
				}
				if len(pos) == 0 {
					commandSender.Output(I18n.T(I18n.InvalidPosition))
					break
				}
				configuration.GlobalFullConfig(env).Main().Position = types.Position{
					X: pos[0],
					Y: pos[1],
					Z: pos[2],
				}
				commandSender.Output(fmt.Sprintf("%s: %v", I18n.T(I18n.PositionGot), pos))
				break
			} else if p.CommandOrigin.UUID.String() == configuration.OneId.String() {
				pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
				if len(pos) == 0 {
					commandSender.Output(I18n.T(I18n.InvalidPosition))
					break
				}
				configuration.GlobalFullConfig(env).Main().End = types.Position{
					X: pos[0],
					Y: pos[1],
					Z: pos[2],
				}
				commandSender.Output(fmt.Sprintf("%s: %v", I18n.T(I18n.PositionGot_End), pos))
				break
			}
			utils_core.ProcessCommandOutput(commandSender, p)
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
		case *packet.UpdateBlock:
			channel, h := commandSender.BlockUpdateSubscribeMap.LoadAndDelete(p.Position)
			if h {
				ch := channel.(chan bool)
				ch <- true
			}
		case *packet.Respawn:
			if p.EntityRuntimeID == conn.GameData().EntityRuntimeID {
				move.Position = p.Position
			}
		case *packet.MovePlayer:
			if p.EntityRuntimeID == conn.GameData().EntityRuntimeID {
				move.Position = p.Position
			} else if p.EntityRuntimeID == move.TargetRuntimeID {
				move.Target = p.Position
			}
		case *packet.CorrectPlayerMovePrediction:
			move.MoveP += 10
			if move.MoveP > 100 {
				move.MoveP = 0
			}
			move.Position = p.Position
			move.Jump()
		case *packet.AddPlayer:
			if move.TargetRuntimeID == 0 && p.EntityRuntimeID != conn.GameData().EntityRuntimeID {
				move.Target = p.Position
				move.TargetRuntimeID = p.EntityRuntimeID
				//fmt.Printf("Got target: %s\n",p.Username)
			}
		}
	}
}

func EstablishConnectionAndInitEnv(env *environment.PBEnvironment) {
	if env.FBAuthClient == nil {
		env.FBAuthClient = fbauth.CreateClient(env)
	}
	pterm.Println(pterm.Yellow(fmt.Sprintf("%s: %s", I18n.T(I18n.ServerCodeTrans), env.LoginInfo.ServerCode)))

	options := []core.Option{}
	if env.IsDebug {
		options = append(options, core.OptionDebug)
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	authenticator := fbauth.NewAccessWrapper(
		env.FBAuthClient.(*fbauth.Client),
		env.LoginInfo.ServerCode,
		env.LoginInfo.ServerPasscode,
		env.LoginInfo.Token,
	)
	conn, err := core.InitMCConnection(ctx, authenticator, options...)

	if err != nil {
		pterm.Error.Println(err)
		if runtime.GOOS == "windows" {
			pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		panic(err)
	}
	if len(env.RespondUser) == 0 {
		if args.GetCustomGameName() == "" {
			go func() {
				user := env.FBAuthClient.(*fbauth.Client).ShouldRespondUser()
				env.RespondUser = user
			}()
		} else {
			env.RespondUser = args.GetCustomGameName()
		}
	}
	env.Connection = conn
	pterm.Println(pterm.Yellow(I18n.T(I18n.ConnectionEstablished)))
	env.UQHolder = uqHolder.NewUQHolder(conn.GameData().EntityRuntimeID)
	env.UQHolder.(*uqHolder.UQHolder).UpdateFromConn(conn)
	env.UQHolder.(*uqHolder.UQHolder).CurrentTick = 0

	env.Resources = &ResourcesControlCenter.Resources{}
	env.ResourcesUpdater = env.Resources.(*ResourcesControlCenter.Resources).Init()
	env.GlobalAPI = &GlobalAPI.GlobalAPI{
		WritePacket: env.Connection.(*minecraft.Conn).WritePacket,
		BotInfo: GlobalAPI.BotInfo{
			BotName:      env.Connection.(*minecraft.Conn).IdentityData().DisplayName,
			BotIdentity:  env.Connection.(*minecraft.Conn).IdentityData().Identity,
			BotRunTimeID: env.Connection.(*minecraft.Conn).GameData().EntityRuntimeID,
			BotUniqueID:  env.Connection.(*minecraft.Conn).GameData().EntityUniqueID,
		},
		Resources: env.Resources.(*ResourcesControlCenter.Resources),
	}

	if args.ShouldEnableOmegaSystem() {
		_, cb := embed.EnableOmegaSystem(env)
		go cb()
		//cb()
	}

	commandSender := commands.InitCommandSender(env)
	functionHolder := env.FunctionHolder.(*function.FunctionHolder)
	function.InitInternalFunctions(functionHolder)
	fbtask.InitTaskStatusDisplay(env)
	move.ConnectTime = time.Time{}
	move.Position = conn.GameData().PlayerPosition
	move.Pitch = conn.GameData().Pitch
	move.Yaw = conn.GameData().Yaw
	move.Connection = conn
	move.RuntimeID = conn.GameData().EntityRuntimeID

	signalhandler.Install(conn, env)

	hostBridgeGamma := env.ScriptBridge.(*script_bridge.HostBridgeGamma)
	hostBridgeGamma.HostSetSendCmdFunc(func(mcCmd string, waitResponse bool) *packet.CommandOutput {
		ud, _ := uuid.NewUUID()
		if !waitResponse {
			env.GlobalAPI.(*GlobalAPI.GlobalAPI).SendCommand(mcCmd, ud)
			return nil
		}
		resp, _ := env.GlobalAPI.(*GlobalAPI.GlobalAPI).SendCommandWithResponce(mcCmd)
		return &resp
	})
	hostBridgeGamma.HostConnectEstablished()
	defer hostBridgeGamma.HostConnectTerminate()

	go func() {
		if args.ShouldMuteWorldChat() {
			return
		}
		for {
			csmsg := <-env.WorldChatChannel
			commandSender.WorldChatOutput(csmsg[0], csmsg[1])
		}
	}()

	taskholder := env.TaskHolder.(*fbtask.TaskHolder)
	types.ForwardedBrokSender = taskholder.BrokSender

	zeroId, _ := uuid.NewUUID()
	oneId, _ := uuid.NewUUID()
	configuration.ZeroId = zeroId
	configuration.OneId = oneId

	if args.ExternalListenAddress() != "" {
		external.ListenExt(env, args.ExternalListenAddress())
	}
	env.UQHolder.(*uqHolder.UQHolder).UpdateFromConn(conn)
	return
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
