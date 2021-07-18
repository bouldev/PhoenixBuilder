package main

import (
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft"
	//"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/google/uuid"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
				//Also: OP Check
				fmt.Printf("<%s> %s\n",p.SourceName,p.Message);
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

func sendCommand(command string,reqid string,conn *minecraft.Conn) {
	zerouuid , _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin {
		Origin:protocol.CommandOriginPlayer,
		UUID:zerouuid,
		RequestID:reqid,
		PlayerUniqueID:0,
	};
	cmdpkt:=&packet.CommandRequest {
		CommandLine:command,
		CommandOrigin:origin,
		Internal:false,
		UnLimited:false,
	}
	conn.WritePacket(cmdpkt)
}

func sendChat(text string, conn *minecraft.Conn) {
	textpacket:=&packet.Text{
		TextType:packet.TextTypeChat,
		SourceName:"FastBuilder",
		NeedsTranslation:false,
		Message:text,
	}
	conn.WritePacket(textpacket)
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
