package core

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"bufio"
	"fmt"
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
			reader := bufio.NewReader(ws)
			message, err := reader.ReadString(0)
			if err != nil {
				fmt.Printf("Error: WebSocket %v\n", err)
				ws.Close()
				return
			}
			fmt.Printf("%s\n", message)
		}
	}()
}

func CreateWebsocketServer() {
	server := &http.Server{
		Addr:    args.ListenAddress,
		Handler: websocket.Handler(defaultWebsocketHandler),
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
