package main

import (
	"C"
	"bufio"
	"crypto/sha256"
	fbauth "phoenixbuilder/cv4/auth"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/term"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/builder"
	"phoenixbuilder/minecraft/command"
	"phoenixbuilder/minecraft/mctype"
	"phoenixbuilder/minecraft/parse"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"encoding/json"
)

type FBPlainToken struct {
	EncryptToken bool `json:"encrypt_token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	fmt.Println("FastBuilder Phoenix Alpha 0.0.1")
	fmt.Println("===============================")
	fmt.Println("Copyright notice:")
	fmt.Println("FastBuilder Phoenix used codes")
	fmt.Println("from Sandertv's gophertunnel that")
	fmt.Println("licensed under MIT license,at:")
	fmt.Println("https://github.com/Sandertv/gophertunnel")
	fmt.Println("===============================")
	fmt.Println("ファスト　ビルダー！")
	fmt.Println("F A S T  B U I L D E R")
	fmt.Println("= = = =  = = = = = = =")
	fmt.Println("Authors: Ruphane, CAIMEO")
	fmt.Println("Copyright (c) FastBuilder DevGroup,")
	fmt.Println("Bouldev 2021")
	defer func() {
		fmt.Println("Oh no! FastBuilder Phoenix crashed! ")
		os.Exit(rand.Int())
	}()
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	currPath := filepath.Dir(ex)
	token := filepath.Join(currPath, "fbtoken")
	version, err := getHash(ex)
	if err != nil {
		fmt.Println("Error reading version: ", err)
		return
	}
	if _, err := os.Stat(token) ; os.IsNotExist(err) {
		fmt.Printf("Enter your FastBuilder User Center username: ")
		fbusername, err := getInputUserName()
		if err != nil {
			fmt.Println(err)
			return
		}
		fbuntrim := fmt.Sprintf("%s",strings.TrimSuffix(fbusername, "\n"))
		fbun := strings.TrimRight(fbuntrim, "\r\n")
		fmt.Printf("Enter your FastBuilder User Center password: ")
		fbpassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Printf("\n")
		tokenstruct:=&FBPlainToken {
			EncryptToken: true,
			Username:fbun,
			Password:string(fbpassword),
		}
		token,err := json.Marshal(tokenstruct)
		if err != nil {
			fmt.Println("Failed to generate temp token")
			fmt.Println(err)
			return
		}
		runShellClient(string(token),version)
		/*fmt.Println("fbtoken not found, please put fbtoken file in the same directory of PhoenixBuilder.")
		os.Exit(2)
		if fi, err := os.Create(token) ; err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Printf("Please enter the Token (Without echoing) : ")
			byteToken, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Println("Error reading token: ", err)
				return
			}
			_, err = fi.WriteString(string(byteToken))
			if err != nil {
				fmt.Println("Error saving token: ", err)
				return 
			}
			defer fi.Close()
			fmt.Printf("\n")
			runShellClient(string(byteToken), version)
		}*/
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
func iOSAppStart(token string,version string,serverCode string, serverPasswd string, onError func()) {
	defer func() {
		onError()
	}()
	runClient(token, version,serverCode,serverPasswd)
}

func runShellClient(token string, version string) {
	code, serverPasswd, err := getRentalServerCode()
	if err != nil {
		fmt.Println(err)
		return
	}
	runClient(token,version,code,serverPasswd)
}

func runClient(token string, version string, code string, serverPasswd string) {
	client := fbauth.CreateClient()
	if token[0] == '{' {
		token = client.GetToken("",token)
		if token == "" {
			fmt.Println("Incorrect username or password")
			return
		}
	}
	serverCode := fmt.Sprintf("%s",strings.TrimSuffix(code, "\n"))
	fmt.Printf("Server: %s \n", serverCode)
	dialer := minecraft.Dialer{
		ServerCode: strings.TrimRight(serverCode, "\r\n"),
		Password:   serverPasswd,
		Version:    version,
		Token:      token,
		Client: 	client,
	}
	conn, err := dialer.Dial("raknet", "")

	if err != nil {
		fmt.Println(err)
		panic(err)
		return
	}
	defer conn.Close()

	// Make the client spawn in the world: This is a blocking operation that will return an error if the
	// client times out while spawning.
	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}

	mConfig := mctype.MainConfig {
		Execute:   "",
		Block:     mctype.Block{
			Name: "iron_block",
			Data: 0,
		},
		OldBlock:  mctype.Block{
			Name: "air",
			Data: 0,
		},
		Begin:     mctype.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		End:       mctype.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		Position:  mctype.Position{
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
	sendCommand("testforblock ~ ~ ~ air", zeroId, conn)


	// A loop that reads packets from the connection until it is closed.
	for {
		// Read a packet from the connection: ReadPacket returns an error if the connection is closed or if
		// a read timeout is set. You will generally want to return or break if this happens.
		pk, err := conn.ReadPacket()
		if err != nil {
			fmt.Printf("%v\n",err)
			break
		}

		// The pk variable is of type packet.Packet, which may be type asserted to gain access to the data
		// they hold:
		user := ""
		switch p := pk.(type) {
		case *packet.Text:
			if p.TextType==packet.TextTypeChat {
				//TODO: SourceName == FBAuthorizedUserName check
				//TODO: OP Check
				if client.ShouldRespondUser(p.SourceName) {
					user = p.SourceName
					chat := strings.Split(p.Message, " ")
					fmt.Printf("<%s> %s\n", p.SourceName, p.Message)

					if chat[0] == "set" {
						X, _ := strconv.Atoi(chat[1])
						Y, _ := strconv.Atoi(chat[2])
						Z, _ := strconv.Atoi(chat[3])
						mConfig.Position = mctype.Position{
							X: X,
							Y: Y,
							Z: Z,
						}
						tellraw(conn, fmt.Sprintf("Positon set: (%v, %v, %v)", X, Y,Z))
					}
					if chat[0] == "get" {
						sendCommand(fmt.Sprintf("execute %s ~ ~ ~ testforblock ~ ~ ~ air", user), zeroId, conn)
					} else{
						cfg := parse.Parse(p.Message, mConfig)
						blocks, err := builder.Generate(cfg)
						if err != nil {
							fmt.Println(err)
						} else {
							t1 := time.Now()
							for _, b := range blocks {
								request := command.SetBlockRequest(b, cfg)
								uuid1, _ := uuid.NewUUID()
								err := sendCommand(request, uuid1, conn)
								if err != nil {
									fmt.Println(err)
								}
							}
							timeUsed := time.Now().Sub(t1)
							tellraw(conn, fmt.Sprintf("%v block(s) have been changed.", len(blocks)))
							tellraw(conn, fmt.Sprintf("Time used: %v second(s)", timeUsed.Seconds()))
							tellraw(conn, fmt.Sprintf("Average speed: %v blocks/second", float64(len(blocks)) / timeUsed.Seconds()))

						}


					}
					}
			}

		case *packet.StartGame:
			pos := p.PlayerPosition
			X := int(pos[0])
			Y := int(pos[1])
			Z := int(pos[2])
			mConfig.Position = mctype.Position{
				X: X,Y: Y,Z: Z,
			}
			tellraw(conn, fmt.Sprintf("Position got: %v %v %v", X, Y, Z))

		case *packet.CommandOutput:
			if p.SuccessCount > 0 && p.CommandOrigin.UUID.String() == zeroId.String() {
				pos, _ := sliceAtoi(p.OutputMessages[0].Parameters)
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
	fmt.Printf("Enter your FastBuilder User Center username: ")
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

func getHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func sendCommand(command string, UUID uuid.UUID, conn *minecraft.Conn) error {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin {
		Origin:         protocol.CommandOriginPlayer,
		UUID:           UUID,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	};
	commandRequest :=&packet.CommandRequest {
		CommandLine:command,
		CommandOrigin:origin,
		Internal:false,
		UnLimited:false,
	}
	return conn.WritePacket(commandRequest)
}

func tellraw(conn *minecraft.Conn, lines ...string) error {
	uuid1, _ := uuid.NewUUID()
	return sendCommand(TellRawRequest(mctype.AllPlayers, lines...), uuid1, conn)
}

func TellRawRequest(target mctype.Target, lines ...string) string {
	now := time.Now().Format("§6[15:04:05]§b")
	cmd := fmt.Sprintf(`tellraw %v {"rawtext":[`, target)
	for i, text := range lines {
		msg := fmt.Sprintf("%v %v", now, text)
		cmd += `{"text":"` + msg + `"}`
		if i != len(lines)-1 {
			cmd += `,`
		}
	}
	return cmd + `]}`
}

func sliceAtoi(sa []string) ([]int, error) {
	si := make([]int, 0, len(sa))
	for _, a := range sa {
		i, err := strconv.Atoi(a)
		if err != nil {
			return si, err
		}
		si = append(si, i)
	}
	return si, nil
}
