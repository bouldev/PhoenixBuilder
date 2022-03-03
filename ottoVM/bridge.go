package ottoVM

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"os"
	"path/filepath"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"sync"
)

type HostBridge struct {
	isConnect bool
	isCli bool
	connetWaiter chan struct{}

	// user input
	vmUserInputChan    chan string
	vmUserInputMu      sync.Mutex
	isWaitingUserInput bool
	userInputReader *bufio.Reader

	// mux
	cliIsReadingUserInput bool
	cliUserInputChan      chan string
	cliVmOutputChan chan  string
	cliVmInputChan chan  string

	// MC function
	vmMcCmd func(fbCmd string, waitResponse bool) *packet.CommandOutput

	// cb funcs
	vmCbsCount map[uint32]uint64
	vmCbs map[uint32]map[uint64]func(packet.Packet)

	// query
	HostQueryExpose map[string]func()string
}

func (hb *HostBridge) Init(isCli bool)  {
	hb.isCli=isCli

	hb.isConnect=false
	hb.connetWaiter=make(chan struct{})

	hb.vmUserInputChan =make(chan string)
	hb.vmUserInputMu =sync.Mutex{}
	hb.isWaitingUserInput=false

	hb.userInputReader=bufio.NewReader(os.Stdin)
	hb.cliIsReadingUserInput=false
	hb.cliUserInputChan=make(chan string)
	hb.cliVmOutputChan=make(chan string)
	hb.cliVmInputChan=make(chan string)

	hb.vmMcCmd=func(fmcCmd string, waitResponse bool) *packet.CommandOutput {
		panic(fmt.Errorf("vmMcCmd not Set!"))
		return nil
	}
	hb.vmCbsCount= map[uint32]uint64{}
	hb.vmCbs= map[uint32]map[uint64]func(packet.Packet){}
	hb.HostQueryExpose= map[string]func() string{}
}

func (hb *HostBridge)GetVMInitFn() func(r Runnable) {
	initFn:=func(r Runnable) {
		vm:=r.GetVM()
		name:=r.GetName()
		scriptName:=fmt.Sprintf("Script[%v]",name)
		FBcallError:= func(funcName string,describe string) otto.Value {
			return vm.MakeCustomError("FB_Func_Call",scriptName+" in "+funcName+" "+describe)
		}
		FBReturnError:= func(funcName string,describe string) otto.Value {
			return vm.MakeCustomError("FB_Func_Return",scriptName+" in "+funcName+" "+describe)
		}
		FBDisconnetedError:= func(funcName string) otto.Value {
			return vm.MakeCustomError("FB_Disconnect",scriptName+" call "+funcName+", but FB-MC connection is disconnected!")
		}
		// common functions
		addTimeOut(vm)
		// add websocket function!
		addWebsocket(vm)

		// function FB_WaitConnect() None
		if err := vm.Set("FB_WaitConnect",
			func(otto.FunctionCall) otto.Value {
				hb.vmWaitConnect(name)
				return otto.Value{}
			},
		); err!=nil{fmt.Println(err)}


		// function FB_GeneralCmd(fbCmd string) None
		if err := vm.Set("_FB_GeneralCmd",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1 || call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_GeneralCmd","no argument fbCmd provided!")
				}
				if !hb.isConnect{
					return FBDisconnetedError("FB_GeneralCmd")
				}
				fbCmd, _ :=call.Argument(0).ToString()
				hb.cliVmInputChan<-fbCmd
				return otto.Value{}
			},
		); err!=nil{fmt.Println(err)}

		// function FB_SendMCCmd(mcCmd string) None
		if err := vm.Set(
			"_FB_SendMCCmd",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1|| call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_SendMCCmd","no argument mcCmd provided!")
				}
				if !hb.isConnect{
					return FBDisconnetedError("FB_SendMCCmd")
				}
				mcCmd,_:=call.Argument(0).ToString()
				fmt.Println(scriptName+" MC Cmd (response=false) "+mcCmd)
				hb.vmMcCmd(mcCmd,false)
				return otto.Value{}
			},
		); err!=nil{fmt.Println(err)}

		// function FB_SendMCCmdAndGetResult(mcCmd string) map[string]interface{}
		if err := vm.Set(
			"_FB_SendMCCmdAndGetResult",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1|| call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_SendMCCmdAndGetResult","no argument mcCmd provided!")
				}
				if !hb.isConnect{
					return FBDisconnetedError("FB_SendMCCmdAndGetResult")
				}
				mcCmd,_:=call.Argument(0).ToString()
				cmd_output:=hb.vmMcCmd(mcCmd,true)
				strObj, _ :=json.Marshal(cmd_output)
				jsObj,err:= otto.ToValue(string(strObj))
				if err==nil{
					return jsObj
				}
				return FBReturnError("FB_SendMCCmdAndGetResult",err.Error())
			},
		); err!=nil{fmt.Println(err)}

		// function FB_RequireUserInput(hint string) string
		if err := vm.Set(
			"_FB_RequireUserInput",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1|| call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_RequireUserInput","no argument hint provided!")
				}
				hint, _ :=call.Argument(0).ToString()
				val,err:=vm.ToValue(hb.vmRequireUserInput(name,hint))
				if err==nil{
					return val
				}
				return FBReturnError("FB_RequireUserInput",err.Error())
			},
		); err!=nil{fmt.Println(err)}

		// function FB_Println(msg string) None
		if err := vm.Set(
			"_FB_Println",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1|| call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_Println","no argument msg provided!")
				}
				msg,_:=call.Argument(0).ToString()
				hb.vmPrintln(name,msg)
				return otto.Value{}
			},
		); err!=nil{fmt.Println(err)}

		// function FB_RegPackCallBack(packetType string,callbackFn func(object)) None
		if err := vm.Set(
			"_FB_RegPackCallBack",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<2{
					return FBcallError("FB_RegPackCallBack","argument number insufficient!")
				}
				packetType,err:=call.Argument(0).ToString()
				if err!=nil{
					return FBcallError("FB_RegPackCallBack","argument packetType is not string!")
				}
				cbFn:=call.Argument(1)
				if !cbFn.IsFunction(){
					return FBcallError("FB_RegPackCallBack","argument callbackFn is not func!")
				}
				deRegFn, err := hb.vmRegPackCallBack(packetType,cbFn)
				if err != nil {
					return FBcallError("FB_RegPackCallBack",err.Error())
				}
				// must use vm.ToValue, not otto.ToValue
				jsDeRegFn,err:=vm.ToValue(func(otto.FunctionCall) otto.Value {
					fmt.Println("deRegCalled!")
					deRegFn()
					return  otto.Value{}
				})
				if err != nil {
					return FBcallError("FB_RegPackCallBack",fmt.Sprintf("can not convert deRegFn from go -> js %v",err))
				}
				return jsDeRegFn
			},
		); err!=nil{fmt.Println(err)}

		// function FB_Query(info string) string
		if err := vm.Set(
			"_FB_Query",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1|| call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_Query","no argument info provided!")
				}
				info,_:=call.Argument(0).ToString()
				queryFn,ok:=hb.HostQueryExpose[info]
				if !ok{
					return FBcallError("FB_Query",info+" not provided, cannot query")
				}
				jsStr,err:=vm.ToValue(queryFn())
				if err!=nil{
					return FBReturnError("FB_Query",fmt.Sprintf("caonnot convert result %v to js string"))
				}
				return jsStr
			},
		); err!=nil{fmt.Println(err)}

		// function FB_SaveFile(fileName string, data string)
		if err := vm.Set(
			"_FB_SaveFile",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1|| call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_SaveFile","no argument fileName provided!")
				}
				fileName, _ :=call.Argument(0).ToString()
				data,err:=call.Argument(1).ToString()
				if err != nil {
					return FBcallError("FB_SaveFile",fmt.Sprintf("data %v is not string!",call.Argument(1)))
				}
				err = hb.vmSaveFile(fileName,data)
				if err != nil {
					return FBReturnError("FB_SaveFile",err.Error())
				}
				return otto.Value{}
			},
		); err!=nil{fmt.Println(err)}

		// function FB_ReadFile(fileName string) string
		if err := vm.Set(
			"_FB_ReadFile",
			func(call otto.FunctionCall) otto.Value {
				if len(call.ArgumentList)<1|| call.ArgumentList[0].IsUndefined(){
					return FBcallError("FB_ReadFile","no argument fileName provided!")
				}
				fileName, _ :=call.Argument(0).ToString()
				data, err := hb.vmReadFile(fileName)
				if err != nil {
					return FBReturnError("FB_RequireUserInput",err.Error())
				}
				val,err:=vm.ToValue(data)
				if err==nil{
					return val
				}
				return FBReturnError("FB_RequireUserInput",err.Error())
			},
		); err!=nil{fmt.Println(err)}
	}
	return initFn
}

func (hb *HostBridge) vmRegPackCallBack(packetType string,cbFn otto.Value) (func(),error) {
	packetID,ok:=PacketNameMap[packetType]
	if !ok{
		return nil,fmt.Errorf("no such packet type")
	}
	cb:=func (pk packet.Packet){
		jsonPacket, err := json.Marshal(pk)
		strPacket:=string(jsonPacket)
		if err != nil {
			fmt.Printf("VM: Convert Packet %v to Json fail: %v",pk,err)
		}
		_, err = cbFn.Call(otto.UndefinedValue(), strPacket)
		if err != nil {
			fmt.Printf("VM: Send Packet %v to Js fail: %v",strPacket,err)
		}
	}
	c,ok:=hb.vmCbsCount[packetID]
	if !ok{
		hb.vmCbsCount[packetID]=0
		hb.vmCbs[packetID]=make(map[uint64]func(packet.Packet))
		c=0
	}
	c+=1
	hb.vmCbs[packetID][c]=cb
	return func(){delete(hb.vmCbs[packetID],c)},nil
}

func (hb *HostBridge) vmPrintln(name string,msg string){
	if hb.isCli{
		fmt.Printf("[%v]: %v\n",name,msg)
	}
}

func (hb *HostBridge) vmWaitConnect(name string){
	<-hb.connetWaiter
}

func (hb *HostBridge) vmGeneralCmd(name string,fbCmd string){
	if hb.isCli{
		hb.cliVmInputChan<-fbCmd
	}
}

func (hb *HostBridge) vmRequireUserInput(name,hint string) string{
	// it is possible that two vm requires user input at the same time
	// so we need a mutex
	hb.vmUserInputMu.Lock()
	hb.isWaitingUserInput=true
	if hb.isCli{
		fmt.Printf("[%v]: %v",name,hint)
	}
	return <-hb.vmUserInputChan
}

func (hb *HostBridge) vmReadFile(path string) (string,error) {
	if strings.Contains(path,"/") || strings.Contains(path,"\\"){
		return "",fmt.Errorf("Can only visit current folder")
	}
	path="[Script_Storage]"+path
	homedir, _ := os.UserHomeDir()
	path = filepath.Join(homedir, ".config/fastbuilder/",path)
	if hb.isCli{
		fp, err := os.OpenFile(path,os.O_RDONLY|os.O_CREATE,0755)
		if err != nil {
			return "", err
		}
		byteData,_:=ioutil.ReadAll(fp)
		return string(byteData),nil
	}
	return "",fmt.Errorf("Not Implemented Now!")
}

func (hb *HostBridge) vmSaveFile(path string, data string) (error) {
	if strings.Contains(path,"/") || strings.Contains(path,"\\"){
		return fmt.Errorf("Can only visit current folder")
	}
	path="[Script_Storage]"+path
	homedir, _ := os.UserHomeDir()
	path = filepath.Join(homedir, ".config/fastbuilder/",path)
	if hb.isCli {
		fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		_, err = fp.Write([]byte(data))
		return err
	}
	return fmt.Errorf("Not Implemented Now!")
}

func (hb *HostBridge) HostConnectTerminate(){
	hb.isConnect=false
	hb.connetWaiter=make(chan struct{})
}

func (hb *HostBridge) HostConnectEstablished(){
	hb.isConnect=true
	close(hb.connetWaiter)
}

func (hb *HostBridge) HostCliInputHijack() {
	// handle user input, either pump to vm or to hb.cliUserInputChan
	// but will never return until get an input to hb.cliUserInputChan
	if hb.cliIsReadingUserInput {
		return
	}
	hb.cliIsReadingUserInput =true
	for{
		cliInput, _ := hb.userInputReader.ReadString('\n')
		//fmt.Println("User Input ",cliInput)
		cliInput = strings.TrimSpace(cliInput)
		if !hb.isWaitingUserInput{
			// send to FB
			hb.cliIsReadingUserInput =false
			hb.cliUserInputChan<-cliInput
			return
		}else{
			// redirect to VM
			hb.isWaitingUserInput=false
			hb.vmUserInputChan <-cliInput
			hb.vmUserInputMu.Unlock()
		}
	}
}

func (hb *HostBridge) HostUser2FBInputHijack() string{
	if !hb.cliIsReadingUserInput {
		go hb.HostCliInputHijack()
	}
	strToFB:=""
	select {
	case strToFB=<-hb.cliUserInputChan:
		break
	case strToFB=<-hb.cliVmInputChan:
		break
	}
	return strToFB
}

func (hb *HostBridge) HostSetSendCmdFunc(fn func(mcCmd string, waitResponse bool) *packet.CommandOutput) {
	hb.vmMcCmd=fn
}

func (hb *HostBridge) HostPumpMcPacket(pk packet.Packet){
	pkID:=pk.ID()
	cbs,ok:=hb.vmCbs[pkID]
	if !ok{
		return
	}
	for _,cb :=range cbs{
		cb(pk)
	}
}