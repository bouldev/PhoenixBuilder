package omega_side

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"phoenixbuilder/omega/utils"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

type pushController struct {
	side              *OmegaSide
	subClientCount    int
	anyPacketWaitor   map[int]*websocket.Conn
	typedPacketWaitor map[int]map[int]*websocket.Conn
	regInfoRemapper   map[int]map[int]bool
	mu                sync.RWMutex
}

func newPushController(s *OmegaSide) *pushController {
	p := &pushController{}
	p.anyPacketWaitor = map[int]*websocket.Conn{}
	p.typedPacketWaitor = map[int]map[int]*websocket.Conn{}
	p.regInfoRemapper = map[int]map[int]bool{}
	p.side = s
	return p
}

func (p *pushController) nextSubClientID() int {
	p.subClientCount++
	return p.subClientCount
}

func (p *pushController) deRegTransportor(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if pkts, hasK := p.regInfoRemapper[id]; hasK {
		for pktID, _ := range pkts {
			if pktID == 0 {
				delete(p.anyPacketWaitor, id)
			} else {
				delete(p.typedPacketWaitor, id)
			}
		}
	}
}

func (p *pushController) regPushType(clientID, pktID int, conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if pktID == 0 {
		if pkts, hasK := p.regInfoRemapper[clientID]; hasK {
			// fmt.Println(pkts)
			for pktID, _ := range pkts {
				if pktID != 0 {
					delete(p.typedPacketWaitor[pktID], clientID)
				}
			}
		} else {
			p.regInfoRemapper[clientID] = map[int]bool{}
		}
		p.anyPacketWaitor[clientID] = conn
		p.regInfoRemapper[clientID] = map[int]bool{0: true}
	} else {
		if clients, hasK := p.typedPacketWaitor[pktID]; hasK {
			clients[clientID] = conn
		} else {
			p.typedPacketWaitor[pktID] = map[int]*websocket.Conn{clientID: conn}
		}
		if mapper, hasK := p.regInfoRemapper[clientID]; hasK {
			mapper[pktID] = true
		} else {
			p.regInfoRemapper[clientID] = map[int]bool{pktID: true}
		}
	}
}

func (p *pushController) pushMCPkt(pktID int, data interface{}) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	name := utils.PktIDInvMapping[pktID]
	if waitors, hasK := p.typedPacketWaitor[pktID]; hasK {
		for _, w := range waitors {
			w.WriteJSON(ServerPush{ID0: 0, Type: "mcPkt", SubType: name, Data: data})
		}
	}
	for _, w := range p.anyPacketWaitor {
		w.WriteJSON(ServerPush{ID0: 0, Type: "mcPkt", SubType: name, Data: data})
	}
}

type omegaSideTransporter struct {
	side        *OmegaSide
	controller  *pushController
	conn        *websocket.Conn
	subClinetId int
	funcMapping map[string]func(args map[string]interface{}, writer func(interface{}))
}

func newTransporter(p *pushController, conn *websocket.Conn) *omegaSideTransporter {
	clientID := p.nextSubClientID()
	transporter := omegaSideTransporter{
		side:        p.side,
		controller:  p,
		subClinetId: clientID,
		conn:        conn,
	}
	transporter.initMapping()
	return &transporter
}

func (t *omegaSideTransporter) regPkt(pktId int) {
	t.controller.regPushType(t.subClinetId, pktId, t.conn)
}

func (t *omegaSideTransporter) response(data []byte, writeFn func(interface{}) error) {
	msg := &clientMsg{}
	if err := json.Unmarshal(data, &msg); err != nil {
		writeFn(serverResp{ID: 0, Violate: true, Data: RespViolatePkt{Err: fmt.Sprintf("cannot decode msg %v", err)}})
		return
	}
	if doFunc, hasK := t.funcMapping[msg.Action]; hasK {
		defer func() {
			r := recover()
			if r != nil {
				writeFn(serverResp{ID: msg.ID, Violate: true, Data: RespViolatePkt{Err: fmt.Sprintf("%v", r)}})
			}
		}()
		doFunc(msg.Args, wrapWriteFn(msg.ID, writeFn))
	} else {
		writeFn(serverResp{ID: msg.ID, Violate: true, Data: RespViolatePkt{Err: fmt.Sprintf("action %v not found", msg.Action)}})
	}
}

func (o *OmegaSide) handle(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		pterm.Error.Println("Omega side WS error:", err)
		return
	}
	defer conn.Close()
	transportor := newTransporter(o.pushController, conn)
	defer func() {
		o.pushController.deRegTransportor(transportor.subClinetId)
	}()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				pterm.Error.Println("An omega side client terminated")
			}
			return
		}
		transportor.response(data, conn.WriteJSON)
	}
}

func (o *OmegaSide) SideUp() {
	handler := http.NewServeMux()
	handler.HandleFunc("/omega_side", o.handle)
	server := http.Server{Handler: handler}
	ln, err := net.Listen("tcp", o.PreferPort)
	if err != nil {
		pterm.Warning.Println("无法使用偏好端口 " + o.PreferPort)
		ln, err = net.Listen("tcp", "localhost:0")
		if err != nil {
			panic("无法打开一个有效端口供 Omega Side 使用")
		}
	}
	addr := ln.Addr().String()
	pterm.Success.Printfln("成功打开了位于 %v 的端口供 Omega Side 使用 (ws://%v/omega_side)", addr, addr)
	o.closeCtx = make(chan struct{})
	o.pushController = newPushController(o)
	go func() {
		err := server.Serve(ln)
		if err != nil {
			pterm.Error.Println("Omega Side 服务端关闭 " + err.Error())
		}
		close(o.closeCtx)
	}()
	if o.DebugServerOnly {
		pterm.Success.Printfln("根据你的设置，Omega Side 将不会启动任何插件，而仅会打开一个用于开发和调试插件的WebSocket端口 ws://%v/omega_side", addr)
		return
	}
	for _, cmd := range o.StartUpCmds {
		remapper := cmd.Remapper
		remapper["[addr]"] = addr
		remapper["[name]"] = cmd.Name
		remapper["[python]"] = o.pythonPath
		o.runCmd(cmd.Name, cmd.Cmd, remapper, o.getWorkingDir())
	}
}
