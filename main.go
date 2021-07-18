package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
	"gophertunnel/minecraft"
	"gophertunnel/minecraft/builder"
	"gophertunnel/minecraft/command"
	"gophertunnel/minecraft/mctype"
	"gophertunnel/minecraft/parse"
	"gophertunnel/minecraft/protocol"
	"gophertunnel/minecraft/protocol/packet"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	config := readConfig()
	dialer := minecraft.Dialer{
		ServerCode:"77329692",
	}
	conn,err:=dialer.Dial("raknet",config.Connection.RemoteAddress);
	if err != nil {
		panic(err)
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
			X: 0,
			Y: 0,
			Z: 0,
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
	sendChat("欢迎使用 FastBuilder!",conn)
	// You will then want to start a for loop that reads packets from the connection until it is closed.
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
		switch p := pk.(type) {
		case *packet.Text:
			if p.TextType==packet.TextTypeChat {
				//TODO: SourceName == FBAuthorizedUserName check
				//TODO: OP Check
				if p.SourceName == "CAIMEO" {
					chat := strings.Split(p.Message, " ")
					if chat[0] == "set" {
						X, _ := strconv.Atoi(chat[1])
						Y, _ := strconv.Atoi(chat[2])
						Z, _ := strconv.Atoi(chat[3])
						mConfig.Position = mctype.Position{
							X: X,
							Y: Y,
							Z: Z,
						}
					}
					fmt.Printf("<%s> %s\n", p.SourceName, p.Message)
					cfg := parse.Parse(p.Message, mConfig)
					blocks, err := builder.Generate(cfg)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println(cfg, blocks)
					for _, b := range blocks {
						request := command.SetBlockRequest(b, cfg)
						uuid1, _ := uuid.NewUUID()
						err := sendCommand(request, uuid1.String(), conn)
						if err != nil {
							fmt.Println(err)
						}
					}

				}
			}


		case *packet.CommandOutput:
			//For example,check op:
			//1: p.SuccessCount>0
			//2: p.CommandOrigin.RequestID==[reqid used on sendCommand().]
		}

		// Write a packet to the connection: Similarly to ReadPacket, WritePacket will (only) return an error
		// if the connection is closed.
		/*p := &packet.RequestChunkRadius{ChunkRadius: 32}
		if err := conn.WritePacket(p); err != nil {
			break
		}*/
	}
}

func sendCommand(command string, requestId string,conn *minecraft.Conn) error {
	unitUuid, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin {
		Origin:         protocol.CommandOriginPlayer,
		UUID:           unitUuid,
		RequestID:      requestId,
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

func sendChat(text string, conn *minecraft.Conn) {
	textpacket:=&packet.Text{
		TextType:packet.TextTypeChat,
		SourceName:"FastBuilder",
		NeedsTranslation:false,
		Message:text,
	}
	err := conn.WritePacket(textpacket)
	if err != nil {
		return
	}
}

type config struct {
	Connection struct {
		LocalAddress  string
		RemoteAddress string
	}
}

func readConfig() config {
	c := config{}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		f, err := os.Create("config.toml")
		if err != nil {
			log.Fatalf("error creating config: %v", err)
		}
		data, err := toml.Marshal(c)
		if err != nil {
			log.Fatalf("error encoding default config: %v", err)
		}
		if _, err := f.Write(data); err != nil {
			log.Fatalf("error writing encoded default config: %v", err)
		}
		_ = f.Close()
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("error decoding config: %v", err)
	}
	if c.Connection.LocalAddress == "" {
		c.Connection.LocalAddress = "0.0.0.0:19132"
	}
	data, _ = toml.Marshal(c)
	if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
		log.Fatalf("error writing config file: %v", err)
	}
	return c
}
