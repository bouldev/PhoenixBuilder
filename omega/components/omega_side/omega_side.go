package omega_side

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

type OmegeSideProcessStartCmd struct {
	Name     string            `json:"旁加载功能名"`
	Cmd      string            `json:"启动指令"`
	Remapper map[string]string `json:"变更选项"`
}

type OmegaSide struct {
	*defines.BasicComponent
	PreferPort     string                     `json:"如果可以则使用这个http端口"`
	StartUpCmds    []OmegeSideProcessStartCmd `json:"启动旁加载进程的指令"`
	closeCtx       chan struct{}
	pushController *pushController
	fileChange     bool
	FileName       string `json:"玩家数据文件"`
	PlayerData     map[string]map[string]interface{}
}

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
		pterm.Error.Println("Omega Side WS Error:", err)
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
				pterm.Error.Println("A Omega Side Client Treminated")
			}
			return
		}
		transportor.response(data, conn.WriteJSON)
	}
}

func (o *OmegaSide) WaitClose() {
	<-o.closeCtx
}

func (o *OmegaSide) getExecDir() string {
	return o.Frame.GetOmegaSideDir()
}

func (o *OmegaSide) OnMCPkt(pktID int, data interface{}) {
	o.pushController.pushMCPkt(pktID, data)
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
	pterm.Success.Printfln("成功打开了位于 %v 的端口供 Omega Side 使用", addr)
	o.closeCtx = make(chan struct{})
	o.pushController = newPushController(o)
	go func() {
		err := server.Serve(ln)
		if err != nil {
			pterm.Error.Println("Omega Side 服务端关闭 " + err.Error())
		}
		close(o.closeCtx)
	}()
	for _, cmd := range o.StartUpCmds {
		remapper := cmd.Remapper
		remapper["[addr]"] = addr
		remapper["[name]"] = cmd.Name
		o.runCmd(cmd.Name, cmd.Cmd, remapper, o.getExecDir())
	}
}

func (o *OmegaSide) runCmd(subProcessName string, cmdStr string, remapping map[string]string, execDir string) {
	cmds := strings.Split(cmdStr, " ")
	execName := ""
	args := []string{}
	i := 0
	for _, frag := range cmds {
		if frag == "" {
			continue
		}
		i++
		if i == 1 {
			execName = frag
		} else {
			for k, v := range remapping {
				frag = strings.ReplaceAll(frag, k, v)
			}
			args = append(args, frag)
		}
	}
	if execName == "" {
		pterm.Info.Println("启动子进程[" + subProcessName + "]: " + cmdStr + " 失败: 未指定 程序名")
		return
	} else {
		pterm.Info.Println("启动子进程["+subProcessName+"]: "+cmdStr+" => 标准化为", strings.Join([]string{pterm.Yellow(execName), pterm.Blue(strings.Join(args, " "))}, " "))
	}
	cmd := exec.Command(execName, args...)
	cmd.Dir = execDir
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	Info := pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.InfoMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoPrefixStyle,
			Text:  fmt.Sprintf("%v", subProcessName),
		},
	}
	Error := pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorPrefixStyle,
			Text:  fmt.Sprintf("%v错误", subProcessName),
		},
	}
	go func() {
		reader := bufio.NewReader(cmdOut)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				// Info.Println("已退出")
				return
			}
			readString = strings.Trim(readString, "\n")
			if readString == "" {
				continue
			}
			o.Frame.GetBackendDisplay().Write(Info.Sprintln(readString))
		}
	}()
	cmdErr, err := cmd.StderrPipe()
	go func() {
		reader := bufio.NewReader(cmdErr)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				Error.Println("已退出")
				return
			}
			readString = strings.Trim(readString, "\n")
			if readString == "" {
				continue
			}
			o.Frame.GetBackendDisplay().Write(Error.Sprintln(readString))
		}
	}()
	go func() {
		err = cmd.Start()
		if err != nil {
			Error.Println(err)
		}
		err = cmd.Wait()
		if err != nil {
			Error.Println(err)
		}
	}()
}

func (o *OmegaSide) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *OmegaSide) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.PlayerData = map[string]map[string]interface{}{}
	err := frame.GetJsonData(o.FileName, &o.PlayerData)
	if err != nil {
		panic(err)
	}
}

func (o *OmegaSide) Activate() {
	o.SideUp()
	o.Frame.GetGameListener().SetOnAnyPacketCallBack(func(p packet.Packet) {
		o.pushController.pushMCPkt(int(p.ID()), p)
	})
}
func (o *OmegaSide) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.PlayerData)
		}
	}
	return nil
}

func (o *OmegaSide) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.PlayerData)
}
