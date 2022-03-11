package bridge

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"time"

	"github.com/google/uuid"
)

var PacketNameMap map[string]uint32
// Will be initialized by wayland_v8/host.

type Terminator struct {
	c             chan struct{}
	isTerminated  bool
	TerminateHook []func()
	RootFolder    string
}

func NewTerminator() *Terminator {
	return &Terminator {
		c: make(chan struct{}),
		isTerminated: false,
		TerminateHook: make([]func(), 0),
	}
}

func (t *Terminator) Terminated() bool {
	return t.isTerminated
}

func (t *Terminator) Terminate() {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("recovery in terminate ", r)
		}
	}()
	t.isTerminated = true
	close(t.c)
	for _, fn := range t.TerminateHook {
		fn()
	}
}

type HostBridge interface {
	// wait fb-mc connection
	WaitConnect(t *Terminator)
	IsConnected() bool
	Println(str string,t *Terminator,scriptName string,end...bool)
	RegPacketCallBack(packetType string,onPacket func(packet.Packet),t *Terminator) (func(),error)
	//Query(info string) string
	GetQueries() map[string]func()string

	// if Get input is called before mc started, mc start will be blocked
	GetInput(hint string, t *Terminator, scriptName string) string

	// the following three cannot call until is connected to mc
	FBCmd(fbCmd string, t *Terminator)
	MCCmd(mcCmd string, t *Terminator, waitResult bool) *packet.CommandOutput

	// FileFunction
	LoadFile(path string) (string, error)
	SaveFile(path string, data string) error
	GetAbsPath(path string) string

	// AutoRestart
	RequireAutoRestart()

	// bot pos
	GetBotPos() (float32,float32,float32)
}

type HostBridgeBeta struct {
	isConnected  bool
	connetWaiter chan struct{}
	// cb funcs
	vmCbsCount map[uint32]uint64
	vmCbs      map[uint32]map[uint64]func(packet.Packet)
	// query
	HostQueryExpose map[string]func() string
	Root            string
}

func NewHostBridge() *HostBridgeBeta {
	return &HostBridgeBeta{
		connetWaiter: make(chan struct{}),
		vmCbsCount:   map[uint32]uint64{},
		vmCbs:        map[uint32]map[uint64]func(packet.Packet){},
		HostQueryExpose: map[string]func() string{
			"user_name": func() string {
				return "2401PT"
			},
			"sha_token": func() string {
				return "sha_token12asjkdao23201"
			},
			"server_code": func() string {
				return "96996635"
			},
			//"script_sha"
			// return by FB_Query
		},
		Root: "wayland_v8/testHome",
	}
}

func (hb *HostBridgeBeta) WaitConnect(t *Terminator) {
	if !hb.isConnected {
		timer := time.NewTimer(time.Second * 1)
		go func() {
			<-timer.C
			hb.isConnected = true
			close(hb.connetWaiter)
		}()
	}
	select {
	case <-hb.connetWaiter:
	case <-t.c:
	}
}

func (hb *HostBridgeBeta) IsConnected() bool {
	return hb.isConnected
}

func (hb *HostBridgeBeta) Println(str string, t *Terminator, scriptName string, end ...bool) {
	if t.isTerminated {
		return
	}
	if len(end) == 1 && !end[0] {
		fmt.Print("[" + scriptName + "]: " + str)
	} else {
		fmt.Println("[" + scriptName + "]: " + str)
	}
}

func (hb *HostBridgeBeta) GetQueries() map[string]func()string {
	return hb.HostQueryExpose
}

func (hb *HostBridgeBeta) FBCmd(fbCmd string,t *Terminator)  {
	if t.isTerminated{
		return
	}
	fmt.Println("[FBCmd]: " + fbCmd)
}

func (hb *HostBridgeBeta) MCCmd(mcCmd string, t *Terminator, waitResult bool) *packet.CommandOutput {
	if t.isTerminated {
		return nil
	}
	fmt.Println("[MCCmd]: " + mcCmd)
	if waitResult {
		return &packet.CommandOutput{
			CommandOrigin: protocol.CommandOrigin{
				Origin:         1,
				UUID:           uuid.UUID{1, 2, 3, 4, 5, 6, 7, 83, 2, 13},
				RequestID:      "RequestID",
				PlayerUniqueID: 5,
			},
			OutputType:   0,
			SuccessCount: 1,
			OutputMessages: []protocol.CommandOutputMessage{{
				Success:    true,
				Message:    "hello!",
				Parameters: nil,
			}},
			DataSet: "",
		}
	} else {
		return nil
	}
}

func (hb *HostBridgeBeta) GetInput(hint string, t *Terminator, scriptName string) string {
	if t.isTerminated {
		return ""
	}

	fmt.Print("[scriptName]: " + hint)
	userInputReader := bufio.NewReader(os.Stdin)
	l, _, _ := userInputReader.ReadLine()
	s := strings.TrimSpace(string(l))
	if t.isTerminated {
		return ""
	}

	return s
}

func (hb *HostBridgeBeta) RegPacketCallBack(packetType string, onPacket func(packet.Packet), t *Terminator) (func(), error) {
	packetID, ok := PacketNameMap[packetType]
	if !ok {
		return nil, fmt.Errorf("no such packet type " + packetType)
	}
	_c, ok := hb.vmCbsCount[packetID]
	c := _c
	if !ok {
		hb.vmCbsCount[packetID] = 0
		hb.vmCbs[packetID] = make(map[uint64]func(packet.Packet))
		c = 0
	}
	c += 1
	hb.vmCbsCount[packetID]++
	hb.vmCbs[packetID][c] = onPacket
	go func() {
		<-t.c
		if _, ok := hb.vmCbs[packetID][c]; ok {
			delete(hb.vmCbs[packetID], c)
		}

	}()
	go func() {
		for {
			if cb, ok := hb.vmCbs[packetID][c]; !ok {
				return
			} else {
				cb(&packet.Text{
					TextType:         0,
					NeedsTranslation: false,
					SourceName:       "fakeUser",
					Message:          "hello from routine",
					Parameters:       nil,
					XUID:             "",
					PlatformChatID:   "",
					PlayerRuntimeID:  "",
				})
				time.Sleep(3 * time.Second)
			}
		}
	}()
	return func() {
		fmt.Println("DeReg called!")
		delete(hb.vmCbs[packetID], c)
	}, nil
}

func (hb *HostBridgeBeta) Query(info string) string {
	if fn, ok := hb.HostQueryExpose[info]; ok {
		return fn()
	} else {
		return ""
	}
}

func (hb *HostBridgeBeta) GetAbsPath(p string) string {
	if !path.IsAbs(hb.Root) {
		pwd, _ := os.Getwd()
		hb.Root = path.Join(pwd, hb.Root)
	}
	if !path.IsAbs(p) {
		p = path.Join(hb.Root, p)
	}
	path.Clean(p)
	return p
}

func (hb *HostBridgeBeta) LoadFile(p string) (string, error) {
	p = hb.GetAbsPath(p)
	fp, err := os.OpenFile(p, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		return "", err
	}
	byteData, err := ioutil.ReadAll(fp)
	if err != nil {
		return "", err
	}
	return string(byteData), nil
}

func (hb *HostBridgeBeta) SaveFile(p string, data string) error {
	p = hb.GetAbsPath(p)
	d, _ := path.Split(p)
	err := os.MkdirAll(d, 0755)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	_, err = fp.Write([]byte(data))
	return err
}

func (hb *HostBridgeBeta) RequireAutoRestart() {

}

func (hb *HostBridgeBeta) GetBotPos()(float32,float32,float32) {
	return 0,0,0
}