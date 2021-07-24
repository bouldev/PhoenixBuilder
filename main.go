package main

import (
	"C"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"golang.org/x/term"
	"io/ioutil"
	"os"
	"path/filepath"
	fbauth "phoenixbuilder/cv4/auth"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/builder"
	"phoenixbuilder/minecraft/command"
	"phoenixbuilder/minecraft/mctype"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/minecraft/utils"
	"phoenixbuilder/minecraft/fbtask"
	"strconv"
	"strings"
	"syscall"
	"time"
	"runtime"
)

type FBPlainToken struct {
	EncryptToken bool   `json:"encrypt_token"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func main() {
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgBlack, pterm.FgRed),
	}
	pterm.Println(pterm.Yellow("FastBuilder Phoenix Alpha 0.1.0 - Hotfix 1"))
	pterm.DefaultBox.Println(pterm.LightCyan("Copyright notice: \n" +
		"FastBuilder Phoenix used codes\n" +
		"from Sandertv's Gophertunnel that\n" +
		"licensed under MIT license, at:\n" +
		"https://github.com/Sandertv/gophertunnel"))
	pterm.Println(pterm.Yellow("ファスト　ビルダー！"))
	pterm.Println(pterm.Yellow("F A S T  B U I L D E R"))
	pterm.Println(pterm.Yellow("Contributors: Ruphane, CAIMEO"))
	pterm.Println(pterm.Yellow("Copyright (c) FastBuilder DevGroup, Bouldev 2021"))
	if runtime.GOOS == "windows" {}
	/*defer func() {
		pterm.Error.Println("Oh no! FastBuilder Phoenix crashed! ")
		if runtime.GOOS == "windows" {
			pterm.Error.Println("Press ENTER to exit.")
			_, _=bufio.NewReader(os.Stdin).ReadString('\n')
		}
		os.Exit(1)
		//os.Exit(rand.Int())
	}()*/
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	currPath := filepath.Dir(ex)
	token := filepath.Join(currPath, "fbtoken")
	version, err := utils.GetHash(ex)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(token); os.IsNotExist(err) {
		fbusername, err := getInputUserName()
		if err != nil {
			panic(err)
		}
		fbuntrim := fmt.Sprintf("%s", strings.TrimSuffix(fbusername, "\n"))
		fbun := strings.TrimRight(fbuntrim, "\r\n")
		fmt.Printf("Enter your FastBuilder User Center password: ")
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
}

//export iOSAppStart
func iOSAppStart(token string, version string, serverCode string, serverPasswd string, onError func()) {
	defer func() {
		onError()
	}()
	runClient(token, version, serverCode, serverPasswd)
}

func runShellClient(token string, version string) {
	code, serverPasswd, err := getRentalServerCode()
	if err != nil {
		fmt.Println(err)
		return
	}
	runClient(token, version, code, serverPasswd)
}

func runClient(token string, version string, code string, serverPasswd string) {
	client := fbauth.CreateClient()
	if token[0] == '{' {
		token = client.GetToken("", token)
		if token == "" {
			fmt.Println("Incorrect username or password")
			return
		}
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		currPath := filepath.Dir(ex)
		tokenPath := filepath.Join(currPath, "fbtoken")
		if fi, err := os.Create(tokenPath); err != nil {
			fmt.Println("Error creating token file: ", err)
			fmt.Println("Error ignored.")
		} else {
			_, err = fi.WriteString(token)
			if err != nil {
				fmt.Println("Error saving token: ", err)
				fmt.Println("Error ignored.")
			}
			fi.Close()
			fi = nil
		}
	}
	serverCode := fmt.Sprintf("%s", strings.TrimSuffix(code, "\n"))
	pterm.Println(pterm.Yellow(fmt.Sprintf("Server: %s", serverCode)))
	dialer := minecraft.Dialer{
		ServerCode: strings.TrimRight(serverCode, "\r\n"),
		Password:   serverPasswd,
		Version:    version,
		Token:      token,
		Client:     client,
	}
	conn, err := dialer.Dial("raknet", "")

	if err != nil {
		pterm.Error.Println(err)
		panic(err)
	}
	defer conn.Close()
	pterm.Println(pterm.Yellow("Successfully created minecraft dialer."))
	user := client.ShouldRespondUser()
	//delay := 1000 //BP MMS
	// Make the client spawn in the world: This is a blocking operation that will return an error if the
	// client times out while spawning.
	
	conn.WritePacket(&packet.PlayerAction {
		EntityRuntimeID: conn.GameData().EntityRuntimeID,
		ActionType: packet.PlayerActionRespawn,
	})
	
	if err := conn.DoSpawn(); err != nil {
		pterm.Error.Println("Failed to spawn")
		panic(err)
	}
	pterm.Println(pterm.Yellow("Player spawned successfully."))

	mConfig := mctype.MainConfig{
		Execute: "",
		Block: builder.IronBlock,
		OldBlock: builder.AirBlock,
		Begin: mctype.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		End: mctype.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		Position: mctype.Position{
			X: 200,
			Y: 100,
			Z: 200,
		},
		Radius:    5,
		Length:    0,
		Width:     0,
		Height:    1,
		Method:    "",
		OldMethod: "",
		Facing:    "y",
		Path:      "",
		Shape:     "solid",
		Delay:     decideDelay(mctype.DelayModeContinuous),
		DelayMode: mctype.DelayModeContinuous,
		DelayThreshold:decideDelayThreshold(),
	}

	zeroId, _ := uuid.NewUUID()
	tellraw(conn, "Welcome to FastBuilder!")
	tellraw(conn, fmt.Sprintf("Operator: %s", user))
	sendCommand("testforblock ~ ~ ~ air", zeroId, conn)
	// A loop that reads packets from the connection until it is closed.
	for {
		// Read a packet from the connection: ReadPacket returns an error if the connection is closed or if
		// a read timeout is set. You will generally want to return or break if this happens.
		pk, err := conn.ReadPacket()
		if err != nil {
			panic(err)
		}

		switch p := pk.(type) {
		case *packet.Text:
			if p.TextType == packet.TextTypeChat {
				if user == p.SourceName {
					chat := strings.Split(p.Message, " ")
					pterm.Println(pterm.Yellow(fmt.Sprintf("<%s>", user)), pterm.LightCyan(p.Message))
					if chat[0] == "test" {
						go func(){
							turn := true
							for {
								if turn {
									conn.WritePacket(&packet.PlayerAction{
										EntityRuntimeID: conn.GameData().EntityRuntimeID,
										ActionType: packet.PlayerActionStartSneak,
									})
									turn=false
								}else{
									conn.WritePacket(&packet.PlayerAction{
										EntityRuntimeID: conn.GameData().EntityRuntimeID,
										ActionType: packet.PlayerActionStopSneak,
									})
									turn=true
								}
								time.Sleep(time.Second)
							}
						}()
					} else if chat[0] == "fbexit" {
						tellraw(conn, "Quit correctly")
						fmt.Printf("Quit correctly\n")
						conn.Close()
						os.Exit(0)
					} else if chat[0] == "set" {
						X, _ := strconv.Atoi(chat[1])
						Y, _ := strconv.Atoi(chat[2])
						Z, _ := strconv.Atoi(chat[3])
						mConfig.Position = mctype.Position{
							X: X,
							Y: Y,
							Z: Z,
						}
						tellraw(conn, fmt.Sprintf("Positon set: (%v, %v, %v)", X, Y, Z))
					} else if chat[0] == "delay" {
						if len(chat) < 3 {
							tellraw(conn, "Invalid suboperand\ndelay mode discrete/continuous/none\ndelay set <delay:s/us>")
							break
						}
						if chat[1] == "set" {
							if mConfig.DelayMode==mctype.DelayModeNone {
								tellraw(conn, "[delay set] is unavailable with delay mode: none")
								break
							}
							ms, err := strconv.Atoi(chat[2])
							if err != nil {
								tellraw(conn, fmt.Sprintf("Setting delay error: ", err))
							} else {
								mConfig.Delay = int64(ms)
								tellraw(conn, fmt.Sprintf("Delay set: %d", ms))
							}
						}else if chat[1]=="mode" {
							delaymode:=mctype.ParseDelayMode(chat[2])
							if delaymode==mctype.DelayModeInvalid {
								tellraw(conn, "Invalid delay mode, possible values are: continuous, discrete, none.")
								break
							}
							mConfig.DelayMode=delaymode
							tellraw(conn, fmt.Sprintf("Delay mode set: %s",chat[2]))
							if delaymode!=mctype.DelayModeNone {
								mConfig.Delay=decideDelay(delaymode)
								tellraw(conn, fmt.Sprintf("Delay automatically set to: %d",mConfig.Delay))
							}
							if delaymode==mctype.DelayModeDiscrete {
								mConfig.DelayThreshold=decideDelayThreshold()
								tellraw(conn, fmt.Sprintf("Delay threshold automatically set to: %d",mConfig.DelayThreshold))
							}
						}else if chat[1]=="threshold" {
							if mConfig.DelayMode!=mctype.DelayModeDiscrete {
								tellraw(conn, "Delay threshold is only available with delay mode: discrete")
								break
							}
							ts, err := strconv.Atoi(chat[2])
							if err != nil {
								tellraw(conn, fmt.Sprintf("Setting delay threshold error: ", err))
								break
							}
							mConfig.DelayThreshold=ts
							tellraw(conn, fmt.Sprintf("Delay threshold set to: %d",ts))
						}else{
							tellraw(conn, "Invalid suboperand\ndelay mode discrete/continuous/none\ndelay set <delay:us>")
							break
						}
					} else if chat[0] == "get" {
						sendCommand("gamerule sendcommandfeedback true", uuid.New(), conn)
						cmd := fmt.Sprintf("execute @a[name=\"%s\"] ~ ~ ~ testforblock ~ ~ ~ air", user)
						sendCommand(cmd, zeroId, conn)
					} else if chat[0] == "task" {
						taskid := int64(-1)
						if len(chat) >= 3 {
							taskido, _ := strconv.Atoi(chat[2])
							taskid=int64(taskido)
						}
						if len(chat) == 1 {
							tellraw(conn, "Invalid suboperand")
							tellraw(conn, "task list\ntask <pause/resume/break> <taskid>")
							break
						}
						if chat[1] == "list" {
							total := 0
							tellraw(conn, "Current tasks:")
							fbtask.TaskMap.Range(func (_tid interface{}, _v interface{}) bool {
								tid,_:=_tid.(int64)
								v,_:=_v.(*fbtask.Task)
								dt:=-1
								dv:=int64(-1)
								if v.Config.DelayMode==mctype.DelayModeDiscrete {
									dt=v.Config.DelayThreshold
								}
								if v.Config.DelayMode!=mctype.DelayModeNone {
									dv=v.Config.Delay
								}
								tellraw(conn, fmt.Sprintf("ID %d - CommandLine:\"%s\", State: %s, Delay: %d, DelayMode: %s, DelayThreshold: %d",tid,v.CommandLine,fbtask.GetStateDesc(v.State),dv,mctype.StrDelayMode(v.Config.DelayMode),dt))
								total++
								return true
							})
							tellraw(conn, fmt.Sprintf("Total: %d",total))
							break
						}else if chat[1] == "pause" {
							if taskid == -1 {
								tellraw(conn, "Invalid taskid.")
								break
							}
							task := fbtask.FindTask(taskid)
							if task == nil {
								tellraw(conn, "Couldn't find a valid task by provided task id.")
								break
							}
							task.Pause()
							tellraw(conn, fmt.Sprintf("[Task %d] - Paused",task.TaskId))
							break
						}else if chat[1] == "resume" {
							if taskid == -1 {
								tellraw(conn, "Invalid taskid.")
								break
							}
							task := fbtask.FindTask(taskid)
							if task == nil {
								tellraw(conn, "Couldn't find a valid task by provided task id.")
								break
							}
							task.Resume()
							tellraw(conn, fmt.Sprintf("[Task %d] - Resumed",task.TaskId))
							break
						}else if chat[1] == "break" {
							if taskid == -1 {
								tellraw(conn, "Invalid taskid.")
								break
							}
							task := fbtask.FindTask(taskid)
							if task == nil {
								tellraw(conn, "Couldn't find a valid task by provided task id.")
								break
							}
							task.Break()
							tellraw(conn, fmt.Sprintf("[Task %d] - Stopped",task.TaskId))
							break
						}else if chat[1] == "setdelay" {
							if len(chat)<4 {
								tellraw(conn, "Arguments count mismatch")
								break
							}
							if taskid == -1 {
								tellraw(conn, "Invalid taskid.")
								break
							}
							idelay, err := strconv.Atoi(chat[3])
							if err != nil {
								tellraw(conn, "Failed to parse delay")
								break
							}
							task := fbtask.FindTask(taskid)
							if task == nil {
								tellraw(conn, "Couldn't find a valid task by provided task id.")
								break
							}
							if task.Config.DelayMode==mctype.DelayModeNone {
								tellraw(conn, "[setdelay] is unavailable with delay mode: none")
								break
							}
							tellraw(conn, fmt.Sprintf("[Task %d] - Delay set: %d",task.TaskId,idelay))
							task.Config.Delay=int64(idelay)
							break
						}else if chat[1] == "setdelaymode" {
							if len(chat)<4 {
								tellraw(conn, "Arguments count mismatch")
								break
							}
							if taskid == -1 {
								tellraw(conn, "Invalid taskid.")
								break
							}
							delaymode := mctype.ParseDelayMode(chat[3])
							if delaymode == mctype.DelayModeInvalid {
								tellraw(conn, "Invalid delay mode, possible values are: continuous, discrete, none.")
								break
							}
							task := fbtask.FindTask(taskid)
							if task == nil {
								tellraw(conn, "Couldn't find a valid task by provided task id.")
								break
							}
							task.Pause()
							task.Config.DelayMode=delaymode
							tellraw(conn, fmt.Sprintf("[Task %d] - Delay mode set: %s",task.TaskId,chat[3]))
							if delaymode!=mctype.DelayModeNone {
								task.Config.Delay=decideDelay(delaymode)
								tellraw(conn, fmt.Sprintf("[Task %d] Delay automatically set to: %d",task.TaskId,task.Config.Delay))
							}
							if delaymode==mctype.DelayModeDiscrete {
								task.Config.DelayThreshold=decideDelayThreshold()
								tellraw(conn, fmt.Sprintf("[Task %d] Delay threshold automatically set to: %d",task.TaskId,task.Config.DelayThreshold))
							}
							break
						}else if chat[1] == "setdelaythreshold" {
							if len(chat)<4 {
								tellraw(conn, "Arguments count mismatch")
								break
							}
							if taskid == -1 {
								tellraw(conn, "Invalid taskid.")
								break
							}
							idelay, err := strconv.Atoi(chat[3])
							if err != nil {
								tellraw(conn, "Failed to parse delay threshold")
								break
							}
							task := fbtask.FindTask(taskid)
							if task == nil {
								tellraw(conn, "Couldn't find a valid task by provided task id.")
								break
							}
							if task.Config.DelayMode==mctype.DelayModeContinuous {
								tellraw(conn, "Delay threshold is unavailable with delay mode: continuous")
								break
							}
							tellraw(conn, fmt.Sprintf("[Task %d] - Delay threshold set: %d",task.TaskId,idelay))
							task.Config.DelayThreshold=idelay
							break
						}else{
							tellraw(conn, "Invalid suboperand")
							tellraw(conn, "task:\ntask list\ntask pause/resume/break <taskid>\ntask setdelay <taskid> <delay>\ntask setdelaymode <taskid> <delaymode:continuous/discrete/none>")
						}
					} else {
						task := fbtask.CreateTask(p.Message, &mConfig, conn)
						if task==nil {
							break
						}
						tellraw(conn, fmt.Sprintf("Task Created, ID=%d.",task.TaskId))
					}
				}
			}

		case *packet.AddPlayer:
			if p.Username == user {
				pterm.Println(pterm.Yellow(fmt.Sprintf("[%s] Operator joined Game", user)))
			}

		case *packet.CommandOutput:
			if p.SuccessCount > 0 && p.CommandOrigin.UUID.String() == zeroId.String() {
				pos, _ := utils.SliceAtoi(p.OutputMessages[0].Parameters)
				mConfig.Position = mctype.Position{
					X: pos[0],
					Y: pos[1],
					Z: pos[2],
				}
				tellraw(conn, fmt.Sprintf("Position got: %v", pos))
			}

		}

	}
}

func getInputUserName() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	pterm.Printf("Enter your FastBuilder User Center username: ")
	fbusername, err := reader.ReadString('\n')
	return fbusername, err
}

func getRentalServerCode() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Please enter the rental server number: ")
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Printf("Enter Password (Press [Enter] if not set, input won't be echoed): ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return code, string(bytePassword), err
}

func readToken(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func sendCommand(command string, UUID uuid.UUID, conn *minecraft.Conn) error {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
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
	return conn.WritePacket(commandRequest)
}

func tellraw(conn *minecraft.Conn, lines ...string) error {
	uuid1, _ := uuid.NewUUID()
	return sendCommand(command.TellRawRequest(mctype.AllPlayers, lines...), uuid1, conn)
}

func decideDelay(delaytype byte) int64 {
	// Will add system check later,so don't merge into other functions.
	if delaytype==mctype.DelayModeContinuous {
		return 1000
	}else if delaytype==mctype.DelayModeDiscrete {
		return 15
	}else{
		return 0
	}
}

func decideDelayThreshold() int {
	// Will add system check later,so don't merge into other functions.
	return 20000
}
