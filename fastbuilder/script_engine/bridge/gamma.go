package bridge

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"phoenixbuilder/fastbuilder/move"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"sync"
	"time"
)

type HostBridgeGamma struct {
	isConnect    bool
	isCli        bool
	connetWaiter chan struct{}
	hostBlock    chan struct{}
	hosBlocked   bool

	// user input
	vmUserInputChan    chan string
	vmUserInputMu      sync.Mutex
	isWaitingUserInput bool
	userInputReader    *bufio.Reader

	// mux
	cliIsReadingUserInput bool
	cliUserInputChan      chan string
	cliVmOutputChan       chan string
	cliVmInputChan        chan string

	// MC function
	vmMcCmd func(fbCmd string, waitResponse bool) *packet.CommandOutput

	// cb funcs
	vmCbsCount map[uint32]uint64
	vmCbs      map[uint32]map[uint64]func(packet.Packet)

	// query
	HostQueryExpose map[string]func() string

	// path
	Root string
}

func (hb *HostBridgeGamma) Init() {
	hb.isConnect = false
	hb.connetWaiter = make(chan struct{})
	hb.hosBlocked = true
	hb.hostBlock = make(chan struct{})

	hb.vmUserInputChan = make(chan string)
	hb.vmUserInputMu = sync.Mutex{}
	hb.isWaitingUserInput = false

	hb.userInputReader = bufio.NewReader(os.Stdin)
	hb.cliIsReadingUserInput = false
	hb.cliUserInputChan = make(chan string)
	hb.cliVmOutputChan = make(chan string)
	hb.cliVmInputChan = make(chan string)

	hb.vmMcCmd = func(fmcCmd string, waitResponse bool) *packet.CommandOutput {
		panic(fmt.Errorf("vmMcCmd not Set!"))
		return nil
	}
	hb.vmCbsCount = map[uint32]uint64{}
	hb.vmCbs = map[uint32]map[uint64]func(packet.Packet){}
	hb.HostQueryExpose = map[string]func() string{}
	homedir, _ := os.UserHomeDir()
	hb.Root = filepath.Join(homedir, ".config/fastbuilder/")
}

func (hb *HostBridgeGamma) HostConnectTerminate() {
	hb.isConnect = false
	hb.connetWaiter = make(chan struct{})
}

func (hb *HostBridgeGamma) HostConnectEstablished() {
	hb.isConnect = true
	close(hb.connetWaiter)
}

func (hb *HostBridgeGamma) HostCliInputHijack() {
	// handle user input, either pump to vm or to hb.cliUserInputChan
	// but will never return until get an input to hb.cliUserInputChan
	if hb.cliIsReadingUserInput {
		return
	}
	hb.cliIsReadingUserInput = true
	for {
		cliInput, _ := hb.userInputReader.ReadString('\n')
		//fmt.Println("User Input ",cliInput)
		cliInput = strings.TrimSpace(cliInput)
		if !hb.isWaitingUserInput {
			// send to FB
			hb.cliIsReadingUserInput = false
			hb.cliUserInputChan <- cliInput
			return
		} else {
			// redirect to VM
			hb.isWaitingUserInput = false
			hb.vmUserInputChan <- cliInput
			hb.vmUserInputMu.Unlock()
		}
	}
}

func (hb *HostBridgeGamma) HostUser2FBInputHijack() string {
	if !hb.cliIsReadingUserInput {
		go hb.HostCliInputHijack()
	}
	strToFB := ""
	select {
	case strToFB = <-hb.cliUserInputChan:
		break
	case strToFB = <-hb.cliVmInputChan:
		break
	}
	return strToFB
}

func (hb *HostBridgeGamma) HostSetSendCmdFunc(fn func(mcCmd string, waitResponse bool) *packet.CommandOutput) {
	hb.vmMcCmd = fn
}

func (hb *HostBridgeGamma) HostPumpMcPacket(pk packet.Packet) {
	go func() {
		pkID := pk.ID()
		cbs, ok := hb.vmCbs[pkID]
		if !ok {
			return
		}
		for _, cb := range cbs {
			cb(pk)
		}
	} ()
}

func (hb *HostBridgeGamma) WaitConnect(t *Terminator) {
	if hb.hosBlocked {
		close(hb.hostBlock)
		hb.hosBlocked = false
	}
	select {
	case <-hb.connetWaiter:
	case <-t.c:
	}
}

func (hb *HostBridgeGamma) HostRemoveBlock() {
	if hb.hosBlocked {
		close(hb.hostBlock)
		hb.hosBlocked = false
	}
}

func (hb *HostBridgeGamma) IsConnected() bool {
	return hb.isConnect
}

func (hb *HostBridgeGamma) Println(str string, t *Terminator, scriptName string, end ...bool) {
	if t.isTerminated {
		return
	}
	if len(end) == 1 && !end[0] {
		fmt.Print("[" + scriptName + "]: " + str)
	} else {
		fmt.Println("[" + scriptName + "]: " + str)
	}
}

func (hb *HostBridgeGamma) FBCmd(fbCmd string, t *Terminator) {
	if t.isTerminated {
		return
	}
	hb.cliVmInputChan <- fbCmd
}

func (hb *HostBridgeGamma) MCCmd(mcCmd string, t *Terminator, waitResult bool) *packet.CommandOutput {
	if t.isTerminated {
		return nil
	}
	return hb.vmMcCmd(mcCmd, waitResult)
}

func (hb *HostBridgeGamma) HostWaitScriptBlock() {
	time.Sleep(time.Second)
	<-hb.hostBlock
}

func (hb *HostBridgeGamma) GetQueries() map[string]func()string {
	return hb.HostQueryExpose
}

func (hb *HostBridgeGamma) GetInput(hint string,t *Terminator,scriptName string) string{
	if t.isTerminated {
		return ""
	}
	// if FB is not connected to MC, at this time
	if !hb.IsConnected(){
		fmt.Printf("[%v]: %v", scriptName, hint)
		userInputReader:=bufio.NewReader(os.Stdin)
		l,_, _ :=userInputReader.ReadLine()
		s:=strings.TrimSpace(string(l))
		if t.isTerminated{
			return ""
		}
		return s
	}
	// it is possible that two vm requires user input at the same time
	// so we need a mutex
	hb.vmUserInputMu.Lock()
	hb.isWaitingUserInput = true
	fmt.Printf("[%v]: %v", scriptName, hint)

	return <-hb.vmUserInputChan
}

func (hb *HostBridgeGamma) RegPacketCallBack(packetType string, onPacket func(packet.Packet), t *Terminator) (func(), error) {
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
	hb.vmCbs[packetID][c] = func(p packet.Packet) {
		if t.isTerminated {
			return
		}
		onPacket(p)
	}
	t.TerminateHook = append(t.TerminateHook, func() {
		delete(hb.vmCbs[packetID], c)
	})
	return func() {
		fmt.Println("DeReg called!")
		delete(hb.vmCbs[packetID], c)
	}, nil
}

func (hb *HostBridgeGamma) Query(info string) string {
	if fn, ok := hb.HostQueryExpose[info]; ok {
		return fn()
	} else {
		return ""
	}
}

func (hb *HostBridgeGamma) GetAbsPath(p string) string {
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

func (hb *HostBridgeGamma) LoadFile(p string) (string, error) {
	p = hb.GetAbsPath(p)
	fp, err := os.OpenFile(p, os.O_RDONLY|os.O_CREATE, 0755)
	defer fp.Close()
	if err != nil {
		return "", err
	}
	byteData, err := ioutil.ReadAll(fp)
	if err != nil {
		return "", err
	}
	return string(byteData), nil
}

func (hb *HostBridgeGamma) SaveFile(p string, data string) error {
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
	defer fp.Close()
	_, err = fp.Write([]byte(data))
	return err
}

func (hb *HostBridgeGamma) GetBotPos() (float32,float32,float32){
	return move.Position.X(),move.Position.Y(),move.Position.Z()
}