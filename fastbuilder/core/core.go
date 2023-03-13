package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"phoenixbuilder/fastbuilder/args"
	blockNBT_API "phoenixbuilder/fastbuilder/bdump/blockNBT/API"
	"phoenixbuilder/fastbuilder/configuration"
	fbauth "phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/external"
	"phoenixbuilder/fastbuilder/function"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/move"
	"phoenixbuilder/fastbuilder/readline"
	script_bridge "phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/script_engine/bridge/script_holder"
	"phoenixbuilder/fastbuilder/signalhandler"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/io/commands"
	"phoenixbuilder/io/special_tasks"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/io/assembler"
	"phoenixbuilder/mirror/io/global"
	"phoenixbuilder/mirror/io/lru"
	"phoenixbuilder/omega/cli/embed"
	"phoenixbuilder/omega/suggest"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

var PassFatal bool = false

func create_environment() *environment.PBEnvironment {
	env := &environment.PBEnvironment{}
	env.UQHolder = nil
	env.NewUQHolder = nil
	env.ActivateTaskStatus = make(chan bool)
	env.TaskHolder = fbtask.NewTaskHolder()
	functionHolder := function.NewFunctionHolder(env)
	env.FunctionHolder = functionHolder
	env.Destructors = []func(){}
	hostBridgeGamma := &script_bridge.HostBridgeGamma{}
	hostBridgeGamma.Init()
	hostBridgeGamma.HostQueryExpose = map[string]func() string{
		"server_code": func() string {
			return env.LoginInfo.ServerCode
		},
		"fb_version": func() string {
			return args.GetFBVersion()
		},
		"uc_username": func() string {
			return env.FBUCUsername
		},
	}
	for _, key := range args.CustomSEUndefineConsts {
		_, found := hostBridgeGamma.HostQueryExpose[key]
		if found {
			delete(hostBridgeGamma.HostQueryExpose, key)
		}
	}
	for key, val := range args.CustomSEConsts {
		hostBridgeGamma.HostQueryExpose[key] = func() string { return val }
	}
	env.ScriptBridge = hostBridgeGamma
	scriptHolder := script_holder.InitScriptHolder(env)
	env.ScriptHolder = scriptHolder
	if args.StartupScript() != "" {
		scriptHolder.LoadScript(args.StartupScript(), env)
	}
	env.Destructors = append(env.Destructors, func() {
		scriptHolder.Destroy()
	})
	hostBridgeGamma.HostRemoveBlock()
	env.LRUMemoryChunkCacher = lru.NewLRUMemoryChunkCacher(12, false)
	env.ChunkFeeder = global.NewChunkFeeder()
	return env
}

// Shouldn't be called when running a debug client
func InitRealEnvironment(token string, server_code string, server_password string) *environment.PBEnvironment {
	env := create_environment()
	env.LoginInfo = environment.LoginInfo{
		Token:          token,
		ServerCode:     server_code,
		ServerPasscode: server_password,
	}
	env.FBAuthClient = fbauth.CreateClient(env)
	return env
}

func InitDebugEnvironment() *environment.PBEnvironment {
	env := create_environment()
	env.IsDebug = true
	env.LoginInfo = environment.LoginInfo{
		ServerCode: "[DEBUG]",
	}
	return env
}

func ProcessTokenDefault(env *environment.PBEnvironment) bool {
	token := env.LoginInfo.Token
	client := fbauth.CreateClient(env)
	env.FBAuthClient = client
	if token[0] == '{' {
		token = client.GetToken("", token)
		if token == "" {
			fmt.Println(I18n.T(I18n.FBUC_LoginFailed))
			return false
		}
		tokenPath := loadTokenPath()
		if fi, err := os.Create(tokenPath); err != nil {
			fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnCreate), err)
			fmt.Println(I18n.T(I18n.ErrorIgnored))
		} else {
			env.LoginInfo.Token = token
			_, err = fi.WriteString(token)
			if err != nil {
				fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnSave), err)
				fmt.Println(I18n.T(I18n.ErrorIgnored))
			}
			fi.Close()
			fi = nil
		}
	}
	return true
}

func InitClient(env *environment.PBEnvironment) {
	if env.FBAuthClient == nil {
		env.FBAuthClient = fbauth.CreateClient(env)
	}
	pterm.Println(pterm.Yellow(fmt.Sprintf("%s: %s", I18n.T(I18n.ServerCodeTrans), env.LoginInfo.ServerCode)))
	var conn *minecraft.Conn
	if env.IsDebug {
		conn = &minecraft.Conn{
			DebugMode: true,
		}
	} else {
		connDeadline := time.NewTimer(time.Minute * 3)
		go func() {
			<-connDeadline.C
			if env.Connection == nil {
				panic(I18n.T(I18n.Crashed_No_Connection))
			}
		}()
		fbauthclient := env.FBAuthClient.(*fbauth.Client)
		dialer := minecraft.Dialer{
			Authenticator: fbauth.NewAccessWrapper(
				fbauthclient,
				env.LoginInfo.ServerCode,
				env.LoginInfo.ServerPasscode,
				env.LoginInfo.Token,
			),
			// EnableClientCache: true,
		}
		cconn, err := dialer.Dial("raknet")

		if err != nil {
			pterm.Error.Println(err)
			if runtime.GOOS == "windows" {
				pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
				_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			}
			panic(err)
		}
		conn = cconn
		if len(env.RespondUser) == 0 {
			if args.GetCustomGameName() == "" {
				go func() {
					user := fbauthclient.ShouldRespondUser()
					env.RespondUser = user
				}()
			} else {
				env.RespondUser = args.GetCustomGameName()
			}
		}
	}
	env.Connection = conn
	conn.WritePacket(&packet.ClientCacheStatus{
		Enabled: false,
	})
	env.UQHolder = uqHolder.NewUQHolder(conn.GameData().EntityRuntimeID)
	env.UQHolder.(*uqHolder.UQHolder).UpdateFromConn(conn)
	env.UQHolder.(*uqHolder.UQHolder).CurrentTick = 0

	env.NewUQHolder = &blockNBT_API.PacketHandleResult{}
	env.NewUQHolder.(*blockNBT_API.PacketHandleResult).InitValue()
	// for blockNBT

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
		chann := make(chan *packet.CommandOutput)
		if waitResponse {
			commandSender.UUIDMap.Store(ud.String(), chann)
		}
		commandSender.SendCommand(mcCmd, ud)
		if waitResponse {
			resp := <-chann
			return resp
		} else {
			return nil
		}
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
			ud, _ := uuid.NewUUID()
			chann := make(chan *packet.CommandOutput)
			commandSender.UUIDMap.Store(ud.String(), chann)
			commandSender.SendCommand(cmd[1:], ud)
			resp := <-chann
			fmt.Printf("%+v\n", resp)
		} else if cmd[0] == '!' {
			ud, _ := uuid.NewUUID()
			chann := make(chan *packet.CommandOutput)
			commandSender.UUIDMap.Store(ud.String(), chann)
			commandSender.SendWSCommand(cmd[1:], ud)
			resp := <-chann
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

		env.NewUQHolder.(*blockNBT_API.PacketHandleResult).HandlePacket(&pk) // for blockNBT

		if env.OmegaAdaptorHolder != nil {
			env.OmegaAdaptorHolder.(*embed.EmbeddedAdaptor).FeedPacketAndByte(pk, data)
			continue
		}
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
		case *packet.StructureTemplateDataResponse:
			special_tasks.ExportWaiter <- p.StructureTemplate
			break
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
			pr, ok := commandSender.UUIDMap.LoadAndDelete(p.CommandOrigin.UUID.String())
			if ok {
				pu := pr.(chan *packet.CommandOutput)
				pu <- p
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

func DestroyClient(env *environment.PBEnvironment) {
	env.Stop()
	env.WaitStopped()
	env.Connection.(*minecraft.Conn).Close()
}

func loadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(I18n.T(I18n.Warning_UserHomeDir))
		homedir = "."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0700)
	token := filepath.Join(fbconfigdir, "fbtoken")
	return token
}

func Fatal() {
	if PassFatal {
		return
	}
	if err := recover(); err != nil {
		if !args.NoReadline() {
			readline.HardInterrupt()
		}
		debug.PrintStack()
		pterm.Error.Println(I18n.T(I18n.Crashed_Tip))
		pterm.Error.Println(I18n.T(I18n.Crashed_StackDump_And_Error))
		pterm.Error.Println(err)
		if args.ShouldEnableOmegaSystem() {
			omegaSuggest := suggest.GetOmegaErrorSuggest(fmt.Sprintf("%v", err))
			fmt.Print(omegaSuggest)
		}
		if runtime.GOOS == "windows" {
			pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		os.Exit(1)
	}
	os.Exit(0)
}
