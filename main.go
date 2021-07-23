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
	"phoenixbuilder/minecraft/parse"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/minecraft/utils"
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
	pterm.Println(pterm.Yellow("FastBuilder Phoenix Alpha 0.0.2"))
	pterm.DefaultBox.Println(pterm.LightCyan("Copyright notice: \n" +
		"FastBuilder Phoenix used codes\n" +
		"from Sandertv's Gophertunnel that\n" +
		"licensed under MIT license,at:\n" +
		"https://github.com/Sandertv/gophertunnel"))
	pterm.Println(pterm.Yellow("ファスト　ビルダー！"))
	pterm.Println(pterm.Yellow("F A S T  B U I L D E R"))
	pterm.Println(pterm.Yellow("Contributors: Ruphane, CAIMEO"))
	pterm.Println(pterm.Yellow("Copyright (c) FastBuilder DevGroup, Bouldev 2021"))
	defer func() {
		pterm.Error.Println("Oh no! FastBuilder Phoenix crashed! ")
		if runtime.GOOS == "windows" {
			pterm.Error.Println("Press ENTER to exit.")
			_, _=bufio.NewReader(os.Stdin).ReadString('\n')
		}
		os.Exit(1)
		//os.Exit(rand.Int())
	}()
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
	delay := 1000 //BP MMS
	// Make the client spawn in the world: This is a blocking operation that will return an error if the
	// client times out while spawning.
	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}

	mConfig := mctype.MainConfig{
		Execute: "",
		Block: mctype.Block{
			Name: "iron_block",
			Data: 0,
		},
		OldBlock: mctype.Block{
			Name: "air",
			Data: 0,
		},
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
					if chat[0] == "fbexit" {
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
						ms, err := strconv.Atoi(chat[1])
						if err != nil {
							tellraw(conn, fmt.Sprintf("Setting delay error: ", err))
						} else {
							delay = ms
							tellraw(conn, fmt.Sprintf("Delay set: %d", delay))
						}
					} else if chat[0] == "get" {
						_ = sendCommand("gamerule sendcommandfeedback true", uuid.New(), conn)
						cmd := fmt.Sprintf("execute @a[name=\"%s\"] ~ ~ ~ testforblock ~ ~ ~ air", user)
						sendCommand(cmd, zeroId, conn)
					} else {
						cfg := parse.Parse(p.Message, mConfig)
						blocks, err := builder.Generate(cfg)
						if cfg.Execute == "" {
							break
						}
						if err != nil /*&& cfg.Execute != ""*/ {
							tellraw(conn, fmt.Sprintf("Error: %v", err))
						} else {
							t1 := time.Now()
							go func() {
								for _, b := range blocks {
									request := command.SetBlockRequest(b, cfg)
									uuid1, _ := uuid.NewUUID()
									err := sendCommand(request, uuid1, conn)
									if err != nil {
										panic(err)
									}
									time.Sleep(time.Duration(delay) * time.Microsecond)
								}
								timeUsed := time.Now().Sub(t1)
								tellraw(conn, fmt.Sprintf("%v block(s) have been changed.", len(blocks)))
								tellraw(conn, fmt.Sprintf("Time used: %v second(s)", timeUsed.Seconds()))
								tellraw(conn, fmt.Sprintf("Average speed: %v blocks/second", float64(len(blocks))/timeUsed.Seconds()))
							}()
						}

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
