package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"path"
	"path/filepath"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/configuration"
	fbauth "phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/function"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/menu"
	"phoenixbuilder/fastbuilder/move"
	"phoenixbuilder/fastbuilder/plugin"
	script_bridge "phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/script_engine/bridge/kickstarter"
	"phoenixbuilder/fastbuilder/signalhandler"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/fastbuilder/world_provider"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type FBPlainToken struct {
	EncryptToken bool   `json:"encrypt_token"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

//Version num should seperate from fellow strings
//for implenting print version feature later
//const FBVersion = "1.4.0"
const FBCodeName = "Phoenix"

// 添加开孔函数
// FB 为 server端
func forwardSend(srcConn net.Conn, dstConn *minecraft.Conn) {
	buf := make([]byte, 0)
	currentBytes := 0
	requiredBytes := 0
	is_fb_cmd := false
	for {
		if requiredBytes == 0 {
			rbuf := make([]byte, 4-currentBytes)
			nbytes, err := srcConn.Read(rbuf)
			if err != nil || nbytes == 0 {
				srcConn.Close()
				fmt.Printf("Transfer: connection (proxy -> fb) closed, because cannot read first 4 bytes from proxy\n\t(err=%v)\n", err)
				return
			}
			currentBytes += nbytes
			buf = append(buf, rbuf...)
			if currentBytes >= 4 {
				requiredBytes = int(binary.LittleEndian.Uint32(buf[:4]))
				if requiredBytes >= int(math.Pow(2, 30)) {
					is_fb_cmd = true
					requiredBytes %= int(math.Pow(2, 30))
				}
			}
		}
		if currentBytes < requiredBytes {
			rbuf := make([]byte, requiredBytes-currentBytes)
			nbytes, err := srcConn.Read(rbuf)
			if err != nil || nbytes == 0 {
				srcConn.Close()
				fmt.Printf("Transfer: connection (proxy -> fb) closed, because cannot correctly read from proxy\n\t(err=%v)\n", err)
				return
			}
			currentBytes += nbytes
			buf = append(buf, rbuf...)
		}
		if currentBytes >= requiredBytes {
			if is_fb_cmd {
				fb_cmd_string := string(buf[4:requiredBytes])
				fmt.Printf("fb cmd string: %v\n", fb_cmd_string)
				function.Process(dstConn, fb_cmd_string)
				is_fb_cmd = false
			} else {
				if dstConn.WritePacketBytes(buf[4:requiredBytes]) != nil {
					srcConn.Close()
					fmt.Print("Transfer: connection (proxy -> fb) closed, because fb -> mc forward fail\n")
					return
				}
			}
			// fmt.Printf("forward fb <- proxy %v\n", buf[4:requiredBytes])
			buf = buf[requiredBytes:currentBytes]
			currentBytes -= requiredBytes
			requiredBytes = 0
			is_fb_cmd = false
		}
	}
}

func StartTransferServer(conn *minecraft.Conn, transferPort string) func(data []byte) {
	listener, err := net.Listen("tcp", transferPort)
	if err != nil {
		fmt.Printf("Transfer: listen fail\n\t(err=%v)\n", err)
		return nil
	}
	fmt.Println("Transfer: server start successfully @ ", transferPort)
	proxyConnMap := make(map[string]net.Conn)

	// 使用一个协程等待连接
	go func() {
		for {
			proxyConn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Transfer: accept new connection fail\n\t(err=%v)\n", err)
				continue
			}
			fmt.Printf("Transfer: accept new connection @ %v\n", proxyConn.RemoteAddr().String())
			proxyConnMap[proxyConn.RemoteAddr().String()] = proxyConn
			// 对于每个连接 使用一个协程处理 proxy -> fb -> mc 转发
			go forwardSend(proxyConn, conn)
		}
	}()

	// 定义单次的 mc -> fb -> proxy 转发函数
	forwardRead := func(data []byte) {
		dataLen := len(data) + 4
		headerBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(headerBytes, uint32(dataLen))
		markedPacketBytes := append(headerBytes, data...)
		for addr := range proxyConnMap {
			proxyConn := proxyConnMap[addr]
			currentBytes := 0
			for currentBytes != dataLen {
				writedBytes, err := proxyConn.Write(markedPacketBytes[currentBytes:dataLen])
				if err != nil || writedBytes == 0 {
					fmt.Printf("Transfer: connection (fb -> proxy) closed, because cannot correctly write to proxy\n\t(err=%v)\n", err)
					delete(proxyConnMap, addr)
					break
				}
				currentBytes += writedBytes
			}
		}
	}
	return forwardRead
}

type Robot struct {
	Token        string `json:"token"`
	Code         string `json:"server_number"`
	ServerPasswd string `json:"server_passwd"`
	TransferPort string `json:"transfer_port"`
	IgnoreUpdate bool   `json:"ignore_update"`
	AutoRestart  bool   `json:"auto_restart"`
}

var robotOverWrite *Robot

func main() {
	args.ParseArgs()
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgBlack, pterm.FgRed),
	}

	I18n.Init()

	pterm.DefaultBox.Println(pterm.LightCyan(I18n.T(I18n.Copyright_Notice_Headline) +
		I18n.T(I18n.Copyright_Notice_Line_1) +
		I18n.T(I18n.Copyright_Notice_Line_2) +
		I18n.T(I18n.Copyright_Notice_Line_3) +
		"https://github.com/Sandertv/gophertunnel"))
	pterm.Println(pterm.Yellow("ファスト　ビルダー"))
	pterm.Println(pterm.Yellow("F A S T  B U I L D E R"))
	pterm.Println(pterm.Yellow("Contributors: Ruphane, CAIMEO, CMA2401PT"))
	pterm.Println(pterm.Yellow("Copyright (c) FastBuilder DevGroup, Bouldev 2022"))
	pterm.Println(pterm.Yellow("FastBuilder Phoenix " + args.GetFBVersion()))

	if I18n.ShouldDisplaySpecial() {
		fmt.Printf("%s", I18n.T(I18n.Special_Startup))
	}

	//if runtime.GOOS == "windows" {}
	defer func() {
		if err := recover(); err != nil {
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
		//os.Exit(rand.Int())
	}()
	if args.DebugMode() {
		runDebugClient()
		return
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	var version string
	if args.ShouldDisableHashCheck() {
		version = "NO_HASH_CHECK"
	} else {
		version, err = utils.GetHash(ex)
		if err != nil {
			panic(err)
		}
	}

	jsonFile, err := os.Open("robot.json")
	if err == nil {
		defer jsonFile.Close()
		fmt.Println("robot.json detected, activating dummy mode")
		byteValue, _ := ioutil.ReadAll(jsonFile)
		robotOverWrite = &Robot{}
		json.Unmarshal([]byte(byteValue), robotOverWrite)
		if robotOverWrite.IgnoreUpdate {
			args.Set_disableHashCheck(true)
			version = "NO_HASH_CHECK"
		}
		runShellClient(robotOverWrite.Token, version)
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
			tokenstruct := &FBPlainToken{
				EncryptToken: true,
				Username:     fbun,
				Password:     string(fbpassword),
			}
			token, err := json.Marshal(tokenstruct)
			if err != nil {
				fmt.Println("Failed to generate temp token")
				fmt.Println(err)
				return
			}
			runShellClient(string(token), version)

		} else {
			token, err := readToken(token)
			if err != nil {
				fmt.Println(err)
				return
			}
			runShellClient(token, version)
		}
	} else {
		runShellClient(args.CustomTokenContent(), version)
	}
}

var successfullyConnectedToFB bool

func runShellClient(token string, version string) {
	successfullyConnectedToFB = false
	var code, serverPasswd string
	var err error
	if robotOverWrite != nil {
		code = robotOverWrite.Code
		serverPasswd = robotOverWrite.ServerPasswd
	} else {
		if !args.SpecifiedServer() {
			code, serverPasswd, err = getRentalServerCode()
		} else {
			code = args.ServerCode()
			serverPasswd = args.ServerPassword()
		}
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	runClient(token, version, code, serverPasswd)
}

func dropInRestartLoop(token string, version string, code string, serverPasswd string) {
	failureCount := 1
	for {
		delayTime := 30 * ((2 << failureCount) / 2)
		if delayTime > 2*60*60 {
			delayTime = 2 * 60 * 60
		}
		fmt.Printf("FB will Restart after %v second, retry time=%v", delayTime, failureCount)
		time.Sleep(time.Duration(delayTime) * time.Second)
		recoverableRun(token, version, code, serverPasswd)
		if successfullyConnectedToFB {
			failureCount = 1
		} else {
			failureCount += 1
		}
	}
}

func recoverableRun(token string, version string, code string, serverPasswd string) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("FB Crashed, reason: ", r)
		}
	}()
	successfullyConnectedToFB = false
	runClient(token, version, code, serverPasswd)
}

func runClient(token string, version string, code string, serverPasswd string) {
	hostBridgeGamma := &script_bridge.HostBridgeGamma{}
	hostBridgeGamma.Init()
	hostBridgeGamma.HostQueryExpose = map[string]func() string{
		"user_name": func() string {
			return configuration.RespondUser
		},
		"server_code": func() string {
			return code
		},
		"fb_version": func() string {
			return version
		},
		"fb_dir": func() string {
			dir, _ := os.Getwd()
			return dir
		},
	}
	allScripts := map[string]func(){}
	defer func() {
		for _, fn := range allScripts {
			fn()
		}
	}()

	if args.StartupScript() == "" {
		hostBridgeGamma.HostRemoveBlock()
	} else {
		stopFn, err := script_kickstarter.LoadScript(args.StartupScript(), hostBridgeGamma)
		if err != nil {
			fmt.Println("Cannot load Startup Script ", err)
			hostBridgeGamma.HostRemoveBlock()
		} else {
			allScripts[args.StartupScript()] = stopFn
			hostBridgeGamma.HostWaitScriptBlock()
		}
	}

	worldchatchannel := make(chan []string)
	client := fbauth.CreateClient(worldchatchannel)
	if token[0] == '{' {
		token = client.GetToken("", token)
		if token == "" {
			if IsUnderLib {
				bridgeLoginFailed(I18n.T(I18n.FBUC_LoginFailed))
				return
			}
			fmt.Println(I18n.T(I18n.FBUC_LoginFailed))
			return
		}
		tokenPath := loadTokenPath()
		if fi, err := os.Create(tokenPath); err != nil {
			fmt.Println("Error creating token file: ", err)
			fmt.Println("Error ignored.")
		} else {
			configuration.UserToken = token
			_, err = fi.WriteString(token)
			if err != nil {
				fmt.Println("Error saving token: ", err)
				fmt.Println("Error ignored.")
			}
			fi.Close()
			fi = nil
		}
	} else {
		configuration.UserToken = token
	}
	serverCode := fmt.Sprintf("%s", strings.TrimSuffix(code, "\n"))
	pterm.Println(pterm.Yellow(fmt.Sprintf("%s: %s", I18n.T(I18n.ServerCodeTrans), serverCode)))
	dialer := minecraft.Dialer{
		ServerCode: serverCode, //strings.TrimRight(serverCode, "\r\n"),
		Password:   serverPasswd,
		Version:    version,
		Token:      token,
		Client:     client,
	}
	conn, err := dialer.Dial("raknet", "")

	if err != nil {
		if IsUnderLib {
			bridgeLoginFailed(fmt.Sprintf("%v", err))
			return
			panic(err)
		}
		pterm.Error.Println(err)
		if runtime.GOOS == "windows" {
			pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		panic(err)
		os.Exit(6)
		//panic(err)
	}
	defer conn.Close()
	if IsUnderLib {
		bridgeConn = conn
		bridgeInitFinished()
	}
	successfullyConnectedToFB = true

	// jsVM
	hostBridgeGamma.HostSetSendCmdFunc(func(mcCmd string, waitResponse bool) *packet.CommandOutput {
		ud, _ := uuid.NewUUID()
		chann := make(chan *packet.CommandOutput)
		if waitResponse {
			command.UUIDMap.Store(ud.String(), chann)
		}
		command.SendCommand(mcCmd, ud, conn)
		if waitResponse {
			resp := <-chann
			return resp
		} else {
			return nil
		}
	})
	hostBridgeGamma.HostConnectEstablished()
	defer hostBridgeGamma.HostConnectTerminate()

	pterm.Println(pterm.Yellow(I18n.T(I18n.ConnectionEstablished)))
	user := client.ShouldRespondUser()
	configuration.RespondUser = user

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
	go func() {
		if args.ShouldMuteWorldChat() {
			return
		}
		for {
			csmsg := <-worldchatchannel
			command.WorldChatTellraw(conn, csmsg[0], csmsg[1])
		}
	}()

	plugin.StartPluginSystem(conn)

	function.InitInternalFunctions()
	fbtask.InitTaskStatusDisplay(conn)
	//world_provider.Init()
	move.ConnectTime = conn.GameData().ConnectTime
	move.Position = conn.GameData().PlayerPosition
	move.Pitch = conn.GameData().Pitch
	move.Yaw = conn.GameData().Yaw
	move.Connection = conn
	move.RuntimeID = conn.GameData().EntityRuntimeID

	signalhandler.Init(conn)

	zeroId, _ := uuid.NewUUID()
	oneId, _ := uuid.NewUUID()
	configuration.ZeroId = zeroId
	configuration.OneId = oneId
	types.ForwardedBrokSender = fbtask.BrokSender
	go func() {
		logger, closeFn := makeLogFile()
		defer closeFn()
		//reader:=bufio.NewReader(os.Stdin)
		for {
			//cmd, _:=getInput()
			//inp, _ := reader.ReadString('\n')
			inp := hostBridgeGamma.HostUser2FBInputHijack()
			logger.Println(inp)
			cmd := strings.TrimRight(inp, "\r\n")
			if len(cmd) == 0 {
				continue
			}
			if cmd[0] == '.' {
				ud, _ := uuid.NewUUID()
				chann := make(chan *packet.CommandOutput)
				command.UUIDMap.Store(ud.String(), chann)
				command.SendCommand(cmd[1:], ud, conn)
				resp := <-chann
				fmt.Printf("%+v\n", resp)
			} else if cmd[0] == '!' {
				ud, _ := uuid.NewUUID()
				chann := make(chan *packet.CommandOutput)
				command.UUIDMap.Store(ud.String(), chann)
				command.SendWSCommand(cmd[1:], ud, conn)
				resp := <-chann
				fmt.Printf("%+v\n", resp)
			}
			if cmd == "menu" {
				menu.OpenMenu(conn)
				fmt.Printf("OK\n")
				continue
			}
			if strings.HasPrefix(cmd, "script") {
				cmdArgs := strings.Split(cmd, " ")
				if len(cmdArgs) > 1 {
					scriptPath := cmdArgs[1]
					if stopFn, ok := allScripts[scriptPath]; ok {
						fmt.Println("Reload Script " + scriptPath)
						stopFn()
						delete(allScripts, scriptPath)
					}
					stopFn, err := script_kickstarter.LoadScript(scriptPath, hostBridgeGamma)
					if err != nil {
						fmt.Println("Cannot load Script ", err)
					}
					allScripts[scriptPath] = stopFn
				} else {
					fmt.Println("Script file not provided!")
				}
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
				if !client.CanSendMessage() {
					command.WorldChatTellraw(conn, "FastBuildeｒ", "Lost connection to the authentication server.")
					break
				}
				client.WorldChat(umsg)
			}
			function.Process(conn, cmd)
		}
	}()

	var forwardRecvFn func([]byte)
	if robotOverWrite != nil {
		forwardRecvFn = StartTransferServer(conn, robotOverWrite.TransferPort)
	}

	// A loop that reads packets from the connection until it is closed.
	for {
		// Read a packet from the connection: ReadPacket returns an error if the connection is closed or if
		// a read timeout is set. You will generally want to return or break if this happens.

		pk, data, err := conn.ReadPacketAndBytes()
		if err != nil {
			panic(err)
		}
		hostBridgeGamma.HostPumpMcPacket(pk)
		if forwardRecvFn != nil {
			forwardRecvFn(data)
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
				tellraw(conn, "Welcome to FastBuilder!")
				tellraw(conn, fmt.Sprintf("Operator: %s", user))
				sendCommand("testforblock ~ ~ ~ air", zeroId, conn)
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
				response := client.TransferData(string(encData), fmt.Sprintf("%d", conn.IdentityData().Uid))
				conn.WritePacket(&packet.PyRpc{
					Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xc, 0x53, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc4},
						[]byte{byte(len(response))},
						[]byte(response),
						[]byte{0xc0},
					}, []byte{}),
				})
			}
			break
		case *packet.SetCommandsEnabled:
			if !p.Enabled {
				sendChat(I18n.T(I18n.Notify_NeedOp), conn)
			}
		case *packet.GameRulesChanged:
			for _, rule := range p.GameRules {
				//fmt.Println(rule.Name, " ", rule.Value)
				if rule.Name == "sendcommandfeedback" {
					sendCommandFeedBack := rule.Value.(bool)
					if !sendCommandFeedBack {
						sendChat(I18n.T(I18n.Notify_TurnOnCmdFeedBack), conn)
						//command.SendSizukanaCommand("gamerule sendcommandfeedback true", conn)
					}
				}
			}
		case *packet.StructureTemplateDataResponse:
			//fmt.Printf("RESPONSE %+v\n",p.StructureTemplate)
			fbtask.ExportWaiter <- p.StructureTemplate
			break
		/*case *packet.InventoryContent:
		for _, item := range p.Content {
			fmt.Printf("InventorySlot %+v\n",item.Stack.NBTData["dataField"])
		}
		break*/
		/*case *packet.InventorySlot:
		fmt.Printf("Slot %d:%+v",p.Slot,p.NewItem.Stack)*/
		case *packet.Text:
			if p.TextType == packet.TextTypeChat {
				for _, item := range plugin.ChatEventListeners {
					item(p.SourceName, p.Message)
				}
				if user == p.SourceName {
					if p.Message[0] == '>' && len(p.Message) > 1 {
						umsg := p.Message[1:]
						if !client.CanSendMessage() {
							command.WorldChatTellraw(conn, "FasｔBuildeｒ", "Lose connection to the authentication server.")
							break
						}
						client.WorldChat(umsg)
					}
					break
					pterm.Println(pterm.Yellow(fmt.Sprintf("<%s>", user)), pterm.LightCyan(p.Message))
					if p.Message[0] == '>' {
						//umsg:=p.Message[1:]
						//
					}
					function.Process(conn, p.Message)
					break
				}
			}
		case *packet.CommandOutput:
			//if p.SuccessCount > 0 {
			if p.CommandOrigin.UUID.String() == configuration.ZeroId.String() {
				pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
				if !(p.OutputMessages[0].Message == "commands.generic.unknown") {
					configuration.IsOp = true
				}
				if len(pos) == 0 {
					tellraw(conn, I18n.T(I18n.InvalidPosition))
					break
				}
				configuration.GlobalFullConfig().Main().Position = types.Position{
					X: pos[0],
					Y: pos[1],
					Z: pos[2],
				}
				tellraw(conn, fmt.Sprintf("%s: %v", I18n.T(I18n.PositionGot), pos))
				break
			} else if p.CommandOrigin.UUID.String() == configuration.OneId.String() {
				pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
				if len(pos) == 0 {
					tellraw(conn, I18n.T(I18n.InvalidPosition))
					break
				}
				configuration.GlobalFullConfig().Main().End = types.Position{
					X: pos[0],
					Y: pos[1],
					Z: pos[2],
				}
				tellraw(conn, fmt.Sprintf("%s: %v", I18n.T(I18n.PositionGot_End), pos))
				break
			}
			//}
			pr, ok := command.UUIDMap.LoadAndDelete(p.CommandOrigin.UUID.String())
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
			channel, h := command.BlockUpdateSubscribeMap.LoadAndDelete(p.Position)
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
			//fmt.Printf("correct %v\n",time.Now())
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

func runDebugClient() {
	hostBridgeGamma := &script_bridge.HostBridgeGamma{}
	hostBridgeGamma.Init()
	hostBridgeGamma.HostQueryExpose = map[string]func() string{
		"user_name": func() string {
			return configuration.RespondUser
		},
		"server_code": func() string {
			return "debug"
		},
		"fb_version": func() string {
			return ""
		},
		"fb_dir": func() string {
			dir, _ := os.Getwd()
			return dir
		},
	}
	allScripts := map[string]func(){}
	defer func() {
		for _, fn := range allScripts {
			fn()
		}
	}()

	if args.StartupScript() == "" {
		hostBridgeGamma.HostRemoveBlock()
	} else {
		stopFn, err := script_kickstarter.LoadScript(args.StartupScript(), hostBridgeGamma)
		if err != nil {
			fmt.Println("Cannot load Startup Script ", err)
			hostBridgeGamma.HostRemoveBlock()
		} else {
			allScripts[args.StartupScript()] = stopFn
			hostBridgeGamma.HostWaitScriptBlock()
		}
	}
	serverCode := fmt.Sprintf("%s", strings.TrimSuffix("[DEBUG, NO SERVER]", "\n"))
	pterm.Println(pterm.Yellow(fmt.Sprintf("%s: %s", I18n.T(I18n.ServerCodeTrans), serverCode)))

	conn := &minecraft.Conn{
		DebugMode: true,
	}
	defer conn.Close()
	if IsUnderLib {
		bridgeConn = conn
		bridgeInitFinished()
	}
	pterm.Println(pterm.Yellow(I18n.T(I18n.ConnectionEstablished)))
	user := "DEBUG USER"
	configuration.RespondUser = user
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
		Content: []byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0x19, 0x61, 0x72, 0x65, 0x6e, 0x61, 0x47, 0x61, 0x6d, 0x65, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x4c, 0x6f, 0x61, 0x64, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x90, 0xc0},
	})

	plugin.StartPluginSystem(conn)

	function.InitInternalFunctions()
	fbtask.InitTaskStatusDisplay(conn)

	signalhandler.Init(conn)

	zeroId, _ := uuid.NewUUID()
	oneId, _ := uuid.NewUUID()
	configuration.ZeroId = zeroId
	configuration.OneId = oneId
	types.ForwardedBrokSender = fbtask.BrokSender
	reader := bufio.NewReader(os.Stdin)
	for {
		inp, _ := reader.ReadString('\n')
		cmd := strings.TrimRight(inp, "\r\n")
		//cmd, _:=getInput()
		if len(cmd) == 0 {
			continue
		}
		if cmd[0] == '.' {
			ud, _ := uuid.NewUUID()
			chann := make(chan *packet.CommandOutput)
			command.UUIDMap.Store(ud.String(), chann)
			command.SendCommand(cmd[1:], ud, conn)
			resp := <-chann
			fmt.Printf("%+v\n", resp)
		} else if cmd[0] == '!' {
			ud, _ := uuid.NewUUID()
			chann := make(chan *packet.CommandOutput)
			command.UUIDMap.Store(ud.String(), chann)
			command.SendWSCommand(cmd[1:], ud, conn)
			resp := <-chann
			fmt.Printf("%+v\n", resp)
		}
		if cmd == "menu" {
			menu.OpenMenu(conn)
			fmt.Printf("OK\n")
			continue
		}
		function.Process(conn, cmd)
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

func sendCommand(commands string, UUID uuid.UUID, conn *minecraft.Conn) error {
	/*requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin{
		Origin:         protocol.CommandOriginPlayer,
		UUID:           UUID,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	}
	commandRequest := &packet.CommandRequest{
		CommandLine:   command,
		CommandOrigin: origin,
		Internal:      false,
		UnLimited:     false,
	}
	return conn.WritePacket(commandRequest)*/
	return command.SendCommand(commands, UUID, conn)
}

func sendChat(content string, conn *minecraft.Conn) error {
	return command.SendChat(content, conn)
}

func tellraw(conn *minecraft.Conn, lines ...string) error {
	return command.Tellraw(conn, lines[0])
	//fmt.Printf("%s\n",lines[0])
	//return nil
	//uuid1, _ := uuid.NewUUID()
	//return sendCommand(command.TellRawRequest(types.AllPlayers, lines...), uuid1, conn)
}

func decideDelay(delaytype byte) int64 {
	// Will add system check later,so don't merge into other functions.
	if delaytype == types.DelayModeContinuous {
		return 1000
	} else if delaytype == types.DelayModeDiscrete {
		return 15
	} else {
		return 0
	}
}

func decideDelayThreshold() int {
	// Will add system check later,so don't merge into other functions.
	return 20000
}

func loadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("WARNING - Failed to obtain the user's home directory. made homedir=\".\";\n")
		homedir = "."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0755)
	token := filepath.Join(fbconfigdir, "fbtoken")
	return token
}

func makeLogFile() (*log.Logger, func()) {
	homedir, err := os.UserHomeDir()
	fileName := path.Join(homedir, ".config/fastbuilder/history.log")
	logFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("Cannot create or append Log file %v (%v)\n", fileName, err)
		return log.New(os.Stdout, "", log.Ldate|log.Ltime), func() {}
	}
	return log.New(logFile, "", log.Ldate|log.Ltime), func() { logFile.Close() }
}
