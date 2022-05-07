package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/configuration"
	fbauth "phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/function"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/menu"
	"phoenixbuilder/fastbuilder/move"
	script_bridge "phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/script_engine/bridge/script_holder"
	"phoenixbuilder/fastbuilder/signalhandler"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/fastbuilder/world_provider"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/cli/embed"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"golang.org/x/term"

	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/external"
	"phoenixbuilder/fastbuilder/readline"
)

func main() {
	args.ParseArgs()
	if len(args.PackScripts()) != 0 {
		os.Exit(script_bridge.MakePackage(args.PackScripts(), args.PackScriptsOut()))
	}
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgBlack, pterm.FgRed),
	}

	I18n.Init()

	pterm.DefaultBox.Println(pterm.LightCyan("https://github.com/LNSSPsd/PhoenixBuilder"))
	pterm.Println(pterm.Yellow("Contributors: Ruphane, CAIMEO, CMA2401PT"))
	pterm.Println(pterm.Yellow("Copyright (c) FastBuilder DevGroup, Bouldev 2022"))
	pterm.Println(pterm.Yellow("PhoenixBuilder " + args.GetFBVersion()))

	if !args.NoReadline() {
		readline.InitReadline()
	}

	if I18n.ShouldDisplaySpecial() {
		fmt.Printf("%s", I18n.T(I18n.Special_Startup))
	}

	defer func() {
		if err := recover(); err != nil {
			if !args.NoReadline() {
				readline.HardInterrupt()
			}
			debug.PrintStack()
			pterm.Error.Println(I18n.T(I18n.Crashed_Tip))
			pterm.Error.Println(I18n.T(I18n.Crashed_StackDump_And_Error))
			pterm.Error.Println(err)
			if runtime.GOOS == "windows" {
				pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
				_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			}
			os.Exit(1)
		}
		os.Exit(0)
	}()
	if args.DebugMode() {
		init_and_run_debug_client()
		return
	}
	if !args.ShouldDisableHashCheck() {
		fmt.Printf("Checking update, please wait...")
		hasUpdate, latestVersion := utils.CheckUpdate(args.GetFBVersion())
		fmt.Printf("OK\n")
		if hasUpdate {
			fmt.Printf("A newer version (%s) of PhoenixBuilder is available.\n", latestVersion)
			fmt.Printf("Please update.\n")
			// To ensure user won't ignore it directly, can be suppressed by command line argument.
			os.Exit(0)
		}
	}

	if !args.SpecifiedToken() {
		token := loadTokenPath()
		if _, err := os.Stat(token); os.IsNotExist(err) {
			fbusername, err := getInputUserName()
			if err != nil {
				panic(err)
			}
			fbuntrim := fmt.Sprintf("%s", strings.TrimSuffix(fbusername, "\n"))
			fbun := strings.TrimRight(fbuntrim, "\r\n")
			fmt.Printf(I18n.T(I18n.EnterPasswordForFBUC))
			fbpassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Printf("\n")
			tokenstruct := &map[string]interface{}{
				"encrypt_token": true,
				"username":      fbun,
				"password":      string(fbpassword),
			}
			token, err := json.Marshal(tokenstruct)
			if err != nil {
				fmt.Println("Failed to generate temp token")
				fmt.Println(err)
				return
			}
			runInteractiveClient(string(token))

		} else {
			token, err := readToken(token)
			if err != nil {
				fmt.Println(err)
				return
			}
			runInteractiveClient(token)
		}
	} else {
		runInteractiveClient(args.CustomTokenContent())
	}
}

func runInteractiveClient(token string) {
	var code, serverPasswd string
	var err error
	if !args.SpecifiedServer() {
		code, serverPasswd, err = getRentalServerCode()
	} else {
		code = args.ServerCode()
		serverPasswd = args.ServerPassword()
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	init_and_run_client(token, code, serverPasswd)
}

func create_environment() *environment.PBEnvironment {
	env := &environment.PBEnvironment{}
	env.UQHolder = nil
	env.ActivateTaskStatus = make(chan bool)
	env.TaskHolder = fbtask.NewTaskHolder()
	functionHolder := function.NewFunctionHolder(env)
	env.FunctionHolder = functionHolder
	hostBridgeGamma := &script_bridge.HostBridgeGamma{}
	hostBridgeGamma.Init()
	hostBridgeGamma.HostQueryExpose = map[string]func() string{
		"responduserDEPRECATED": func() string {
			return env.RespondUser
		},
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
	/*if args.StartupScript() == "" {
		hostBridgeGamma.HostRemoveBlock()
	} else {
		if scriptHolder.LoadScript(args.StartupScript(), env) {
			hostBridgeGamma.HostWaitScriptBlock()
		} else {
			hostBridgeGamma.HostRemoveBlock()
		}
	}*/
	if args.StartupScript() != "" {
		scriptHolder.LoadScript(args.StartupScript(), env)
	}
	hostBridgeGamma.HostRemoveBlock()
	return env
}

func init_and_run_debug_client() {
	env := create_environment()
	env.IsDebug = true

	scriptHolder := env.ScriptHolder.(*script_holder.ScriptHolder)
	defer scriptHolder.Destroy()

	runClient(env)
}

func init_and_run_client(token string, code string, server_password string) {
	env := create_environment()
	env.LoginInfo = environment.LoginInfo{
		Token:          token,
		ServerCode:     code,
		ServerPasscode: server_password,
	}

	scriptHolder := env.ScriptHolder.(*script_holder.ScriptHolder)
	defer scriptHolder.Destroy()

	client := fbauth.CreateClient(env)
	env.FBAuthClient = client
	if token[0] == '{' {
		token = client.GetToken("", token)
		if token == "" {
			fmt.Println(I18n.T(I18n.FBUC_LoginFailed))
			return
		}
		tokenPath := loadTokenPath()
		if fi, err := os.Create(tokenPath); err != nil {
			fmt.Println("Error creating token file: ", err)
			fmt.Println("Error ignored.")
		} else {
			env.LoginInfo.Token = token
			_, err = fi.WriteString(token)
			if err != nil {
				fmt.Println("Error saving token: ", err)
				fmt.Println("Error ignored.")
			}
			fi.Close()
			fi = nil
		}
	}
	runClient(env)
}

func runClient(env *environment.PBEnvironment) {
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
				panic("connection not established after very long time")
			}
		}()
		fbauthclient := env.FBAuthClient.(*fbauth.Client)
		dialer := minecraft.Dialer{
			ServerCode: env.LoginInfo.ServerCode, //strings.TrimRight(serverCode, "\r\n"),
			Password:   env.LoginInfo.ServerPasscode,
			Token:      env.LoginInfo.Token,
			Client:     fbauthclient,
		}
		cconn, err := dialer.Dial("raknet", "")

		if err != nil {
			pterm.Error.Println(err)
			if runtime.GOOS == "windows" {
				pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
				_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			}
			panic(err)
		}
		conn = cconn
		go func() {
			user := fbauthclient.ShouldRespondUser()
			env.RespondUser = user
		}()
		env.WorldChatChannel = make(chan []string)
	}
	defer conn.Close()
	defer func() {
		env.Stop()
		env.WaitStopped()
	}()

	pterm.Println(pterm.Yellow(I18n.T(I18n.ConnectionEstablished)))

	runtimeid := fmt.Sprintf("%d", conn.GameData().EntityUniqueID)
	if !args.NoPyRpc() {
		conn.WritePacket(&packet.PyRpc{
			Content: []byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xc, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x73, 0x69, 0x6e, 0x67, 0x4d, 0x6f, 0x64, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0x90, 0xc0},
		})
		conn.WritePacket(&packet.PyRpc{
			Content: []byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xf, 0x53, 0x79, 0x6e, 0x63, 0x56, 0x69, 0x70, 0x53, 0x6b, 0x69, 0x6e, 0x55, 0x75, 0x69, 0x64, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc0, 0xc0},
		})
		conn.WritePacket(&packet.PyRpc{
			Content: []byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0x1f, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4c, 0x6f, 0x61, 0x64, 0x41, 0x64, 0x64, 0x6f, 0x6e, 0x73, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x65, 0x64, 0x46, 0x72, 0x6f, 0x6d, 0x47, 0x61, 0x63, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x90, 0xc0},
		})
		conn.WritePacket(&packet.PyRpc{
			Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xb, 0x4d, 0x6f, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x43, 0x32, 0x53, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x94, 0xc4, 0x9, 0x4d, 0x69, 0x6e, 0x65, 0x63, 0x72, 0x61, 0x66, 0x74, 0xc4, 0x6, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0xc4, 0x12, 0x47, 0x65, 0x74, 0x4c, 0x6f, 0x61, 0x64, 0x65, 0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x81, 0xc4, 0x8, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x49, 0x64, 0xc4},
				[]byte{byte(len(runtimeid))},
				[]byte(runtimeid),
				[]byte{0xc0},
			}, []byte{}),
		})
		conn.WritePacket(&packet.PyRpc{
			Content: []byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0x19, 0x61, 0x72, 0x65, 0x6e, 0x61, 0x47, 0x61, 0x6d, 0x65, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x4c, 0x6f, 0x61, 0x64, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x90, 0xc0},
		})
		conn.WritePacket(&packet.PyRpc{
			Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xb, 0x4d, 0x6f, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x43, 0x32, 0x53, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x94, 0xc4, 0x9, 0x4d, 0x69, 0x6e, 0x65, 0x63, 0x72, 0x61, 0x66, 0x74, 0xc4, 0xe, 0x76, 0x69, 0x70, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0xc4, 0xc, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x55, 0x69, 0x49, 0x6e, 0x69, 0x74, 0xc4},
				[]byte{byte(len(runtimeid))},
				[]byte(runtimeid),
				[]byte{0xc0},
			}, []byte{}),
		})
	}
	conn.WritePacket(&packet.ClientCacheStatus{
		Enabled: false,
	})
	env.Connection = conn
	env.UQHolder = uqHolder.NewUQHolder(conn.GameData().EntityRuntimeID)
	env.UQHolder.(*uqHolder.UQHolder).CurrentTick = uint64(time.Now().Sub(conn.GameData().ConnectTime).Milliseconds()) / 50

	if args.ShouldEnableOmegaSystem() {
		fmt.Println("Omega System Enabled!")
		embed.EnableOmegaSystem(env)
	}

	commandSender := command.InitCommandSender(env)
	functionHolder := env.FunctionHolder.(*function.FunctionHolder)
	function.InitInternalFunctions(functionHolder)
	fbtask.InitTaskStatusDisplay(env)
	move.ConnectTime = conn.GameData().ConnectTime
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
			commandSender.WorldChatTellraw(csmsg[0], csmsg[1])
		}
	}()

	zeroId, _ := uuid.NewUUID()
	oneId, _ := uuid.NewUUID()
	configuration.ZeroId = zeroId
	configuration.OneId = oneId
	taskholder := env.TaskHolder.(*fbtask.TaskHolder)
	types.ForwardedBrokSender = taskholder.BrokSender
	var captureFp *os.File
	if captureOutputFileName := args.CaptureOutputFile(); captureOutputFileName != "" {
		if fp, err := os.OpenFile(captureOutputFileName, os.O_CREATE|os.O_WRONLY, 0755); err != nil {
			panic(err)
		} else {
			captureFp = fp
			fmt.Println("Capture On: FastBuilder > ", captureOutputFileName)
		}
	}
	go func() {
		if args.NoReadline() {
			return
		}
		for {
			cmd := readline.Readline(env)
			if len(cmd) == 0 {
				continue
			}
			if env.OmegaAdaptorHolder != nil && !strings.Contains(cmd, "exit") {
				env.OmegaAdaptorHolder.(*embed.EmbeddedAdaptor).FeedBackendCommand(cmd)
				continue
			}
			if strings.TrimSpace(cmd) == "capture close" {
				if captureFp != nil {
					captureFp.Close()
					captureFp = nil
					fmt.Println("Capture Closed")
				}
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
			if cmd == "menu" {
				menu.OpenMenu(env)
				fmt.Printf("OK\n")
				continue
			}
			if cmd == "move" {
				go func() {
					/*var counter int=0
					var direction bool=false
					for{
						if counter%20==0 {
							//move.Jump()
						}
						if counter>280 {
							counter=0
							direction= !direction
						}
						if direction {
							move.Move(-2+2*moveP/100,0,2*moveP/100)
							time.Sleep(time.Second/20)
							counter++
							continue
						}else{
							move.Move(2*moveP/100,0,-2+2*moveP/100)
							time.Sleep(time.Second/20)
							counter++
							continue
						}
					}*/
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
						commandSender.WorldChatTellraw("FastBuildeｒ", "Lost connection to the authentication server.")
						break
					}
					fbcl.WorldChat(umsg)
				}
			}
			functionHolder.Process(cmd)
		}
	}()

	if args.ExternalListenAddress() != "" {
		external.ListenExt(env, args.ExternalListenAddress())
	}
	env.UQHolder.(*uqHolder.UQHolder).UpdateFromConn(conn)
	for {
		pk, data, err := conn.ReadPacketAndBytes()
		if err != nil {
			panic(err)
		}
		if captureFp != nil {
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(len(data)))
			_, err := captureFp.Write(buf)
			if err != nil {
				panic("dump to capture file (len hdr) fail " + err.Error())
			}
			_, err = captureFp.Write(data)
			if err != nil {
				panic("dump to capture file fail " + err.Error())
			}
		}
		if env.OmegaAdaptorHolder != nil {
			env.OmegaAdaptorHolder.(*embed.EmbeddedAdaptor).FeedPacket(pk)
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

		switch p := pk.(type) {
		case *packet.PyRpc:
			if args.NoPyRpc() {
				break
			}
			//fmt.Printf("PyRpc!\n")
			if strings.Contains(string(p.Content), "GetLoadingTime") {
				//fmt.Printf("GetLoadingTime!!\n")
				uid := conn.IdentityData().Uid
				num := uid&255 ^ (uid&65280)>>8
				curTime := time.Now().Unix()
				num = curTime&3 ^ (num&7)<<2 ^ (curTime&252)<<3 ^ (num&248)<<8
				numcont := make([]byte, 2)
				binary.BigEndian.PutUint16(numcont, uint16(num))
				conn.WritePacket(&packet.PyRpc{
					Content: []byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0x12, 0x53, 0x65, 0x74, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x6f, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x54, 0x69, 0x6d, 0x65, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xcd, numcont[0], numcont[1], 0xc0},
				})
				// Good job, netease, you wasted 3 days of my idle time
				// (-Ruphane)

				// See analyze/nemcfix/final.py for its python version
				// and see analyze/ for how I did it.
				//tellraw(conn, "Welcome to FastBuilder!")
				//tellraw(conn, fmt.Sprintf("Operator: %s", env.RespondUser))
				//sendCommand("testforblock ~ ~ ~ air", zeroId, conn)
			} else if strings.Contains(string(p.Content), "check_server_contain_pet") {
				//fmt.Printf("Checkpet!!\n")

				// Pet req
				/*conn.WritePacket(&packet.PyRpc {
					Content: bytes.Join([][]byte{[]byte{0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x93,0xc4,0xb,0x4d,0x6f,0x64,0x45,0x76,0x65,0x6e,0x74,0x43,0x32,0x53,0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x94,0xc4,0x9,0x4d,0x69,0x6e,0x65,0x63,0x72,0x61,0x66,0x74,0xc4,0x3,0x70,0x65,0x74,0xc4,0x12,0x73,0x75,0x6d,0x6d,0x6f,0x6e,0x5f,0x70,0x65,0x74,0x5f,0x72,0x65,0x71,0x75,0x65,0x73,0x74,0x87,0xc4,0x13,0x61,0x6c,0x6c,0x6f,0x77,0x5f,0x73,0x74,0x65,0x70,0x5f,0x6f,0x6e,0x5f,0x62,0x6c,0x6f,0x63,0x6b,0xc2,0xc4,0xb,0x61,0x76,0x6f,0x69,0x64,0x5f,0x6f,0x77,0x6e,0x65,0x72,0xc3,0xc4,0x7,0x73,0x6b,0x69,0x6e,0x5f,0x69,0x64,0xcd,0x27,0x11,0xc4,0x9,0x70,0x6c,0x61,0x79,0x65,0x72,0x5f,0x69,0x64,0xc4},
							[]byte{byte(len(runtimeid))},
								[]byte(runtimeid),
								[]byte{0xc4,0x6,0x70,0x65,0x74,0x5f,0x69,0x64,0x1,0xc4,0xa,0x6d,0x6f,0x64,0x65,0x6c,0x5f,0x6e,0x61,0x6d,0x65,0xc4,0x14,0x74,0x79,0x5f,0x79,0x75,0x61,0x6e,0x73,0x68,0x65,0x6e,0x67,0x68,0x75,0x6c,0x69,0x5f,0x30,0x5f,0x30,0xc4,0x4,0x6e,0x61,0x6d,0x65,0xc4,0xc,0xe6,0x88,0x91,0xe7,0x9a,0x84,0xe4,0xbc,0x99,0xe4,0xbc,0xb4,0xc0},
						},[]byte{}),
				})*/

			} else if strings.Contains(string(p.Content), "summon_pet_response") {
				//fmt.Printf("summonpetres\n")
				/*conn.WritePacket(&packet.PyRpc {
					Content: []byte{0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x93,0xc4,0x19,0x61,0x72,0x65,0x6e,0x61,0x47,0x61,0x6d,0x65,0x50,0x6c,0x61,0x79,0x65,0x72,0x46,0x69,0x6e,0x69,0x73,0x68,0x4c,0x6f,0x61,0x64,0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x90,0xc0},
				})
				conn.WritePacket(&packet.PyRpc {
					Content: bytes.Join([][]byte{[]byte{0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x93,0xc4,0xb,0x4d,0x6f,0x64,0x45,0x76,0x65,0x6e,0x74,0x43,0x32,0x53,0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x94,0xc4,0x9,0x4d,0x69,0x6e,0x65,0x63,0x72,0x61,0x66,0x74,0xc4,0xe,0x76,0x69,0x70,0x45,0x76,0x65,0x6e,0x74,0x53,0x79,0x73,0x74,0x65,0x6d,0xc4,0xc,0x50,0x6c,0x61,0x79,0x65,0x72,0x55,0x69,0x49,0x6e,0x69,0x74,0xc4},
							[]byte{byte(len(runtimeid))},
								[]byte(runtimeid),
								[]byte{0xc0},
							},[]byte{}),
				})*/
				/*conn.WritePacket(&packet.PyRpc {
					Content: []byte{0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x93,0xc4,0x1f,0x43,0x6c,0x69,0x65,0x6e,0x74,0x4c,0x6f,0x61,0x64,0x41,0x64,0x64,0x6f,0x6e,0x73,0x46,0x69,0x6e,0x69,0x73,0x68,0x65,0x64,0x46,0x72,0x6f,0x6d,0x47,0x61,0x63,0x82,0xc4,0x8,0x5f,0x5f,0x74,0x79,0x70,0x65,0x5f,0x5f,0xc4,0x5,0x74,0x75,0x70,0x6c,0x65,0xc4,0x5,0x76,0x61,0x6c,0x75,0x65,0x90,0xc0},
				})*/
			} else if strings.Contains(string(p.Content), "GetStartType") {
				// 2021-12-22 10:51~11:55
				// Thank netease for wasting my time again ;)
				encData := p.Content[68 : len(p.Content)-1]
				client := env.FBAuthClient.(*fbauth.Client)
				response := client.TransferData(string(encData), fmt.Sprintf("%s", env.Uid))
				conn.WritePacket(&packet.PyRpc{
					Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xc, 0x53, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc4},
						[]byte{byte(len(response))},
						[]byte(response),
						[]byte{0xc0},
					}, []byte{}),
				})
			} else if strings.Contains(string(p.Content), "GetMCPCheckNum") {
				firstArgLenB := p.Content[69:71]
				firstArgLen := binary.BigEndian.Uint16(firstArgLenB)
				secondArgLen := uint16(p.Content[72+firstArgLen])
				secondArg := string(p.Content[73+firstArgLen : 73+firstArgLen+secondArgLen])
				valM := utils.GetMD5(fmt.Sprintf("296<6puv?ol%sk", secondArg))
				conn.WritePacket(&packet.PyRpc{
					Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xe, 0x53, 0x65, 0x74, 0x4d, 0x43, 0x50, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc4, 0x20},
						[]byte(valM),
						[]byte{0xc0},
					}, []byte{}),
				})
			}
			break
		case *packet.StructureTemplateDataResponse:
			fbtask.ExportWaiter <- p.StructureTemplate
			break
		case *packet.Text:
			if p.TextType == packet.TextTypeChat {
				/*for _, item := range plugin.ChatEventListeners {
					item(p.SourceName, p.Message)
				}*/
				break
			}
		case *packet.CommandOutput:
			if p.CommandOrigin.UUID.String() == configuration.ZeroId.String() {
				pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
				if !(p.OutputMessages[0].Message == "commands.generic.unknown") {
					configuration.IsOp = true
				}
				if len(pos) == 0 {
					commandSender.Tellraw(I18n.T(I18n.InvalidPosition))
					break
				}
				configuration.GlobalFullConfig(env).Main().Position = types.Position{
					X: pos[0],
					Y: pos[1],
					Z: pos[2],
				}
				commandSender.Tellraw(fmt.Sprintf("%s: %v", I18n.T(I18n.PositionGot), pos))
				break
			} else if p.CommandOrigin.UUID.String() == configuration.OneId.String() {
				pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
				if len(pos) == 0 {
					commandSender.Tellraw(I18n.T(I18n.InvalidPosition))
					break
				}
				configuration.GlobalFullConfig(env).Main().End = types.Position{
					X: pos[0],
					Y: pos[1],
					Z: pos[2],
				}
				commandSender.Tellraw(fmt.Sprintf("%s: %v", I18n.T(I18n.PositionGot_End), pos))
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
		case *packet.LevelChunk:
			if world_provider.ChunkInput != nil {
				world_provider.ChunkInput <- p
			} else {
				world_provider.DoCache(p)
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

func getInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	inp, err := reader.ReadString('\n')
	inpl := strings.TrimRight(inp, "\r\n")
	return inpl, err
}

func getInputUserName() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	pterm.Printf(I18n.T(I18n.Enter_FBUC_Username))
	fbusername, err := reader.ReadString('\n')
	return fbusername, err
}

func getRentalServerCode() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(I18n.T(I18n.Enter_Rental_Server_Code))
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Printf(I18n.T(I18n.Enter_Rental_Server_Password))
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return strings.TrimRight(code, "\r\n"), string(bytePassword), err
}

func readToken(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func loadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("WARNING - Failed to obtain the user's home directory. made homedir=\".\";\n")
		homedir = "."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0700)
	token := filepath.Join(fbconfigdir, "fbtoken")
	return token
}
