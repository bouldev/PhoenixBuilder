package core

import (
	"fmt"
	"bufio"
	"net/http"
	//"encoding/json"
	"phoenixbuilder/fastbuilder/args"

	"golang.org/x/net/websocket"
)

type Message struct {
	RequestID string `json:"request_id,omitempty"`
	// ^ Make empty to dishonor response, set to "sync" to wait until the response available
	Action string `json:"action"`
	// Payloads are read dynamically
}

func defaultWebsocketHandler(ws *websocket.Conn) {
	go func() {
		for {
			reader:=bufio.NewReader(ws)
			message, err:=reader.ReadString(0)
			if err!=nil {
				fmt.Printf("Error: WebSocket %v\n", err)
				ws.Close()
				return
			}
			fmt.Printf("%s\n", message)
		}
	} ()
}

func CreateWebsocketServer() {
	server:=&http.Server {
		Addr: args.ListenAddress,
		Handler: websocket.Handler(defaultWebsocketHandler),
	}
	err:=server.ListenAndServe()
	if err!=nil {
		panic(err)
	}
}
