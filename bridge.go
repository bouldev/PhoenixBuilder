package main

import (
	"os"
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/fastbuilder/configuration"
	"path/filepath"
	"runtime/debug"
	"unsafe"
)

/*
static void** cbs=NULL;
static inline void SetCallbackSheet(void *cbs1) {
	cbs=cbs1;
}
static inline void ReportPanic(char *error) {
	((void(*)(char*))cbs[0])(error);
}
static inline void LoginFailed(char *msg) {
	((void(*)(char*))cbs[1])(msg);
}
static inline void InitFinished() {
	((void(*)(void))cbs[2])();
}
*/
import "C"

var bridgeConn *minecraft.Conn = nil

func bridgeLoginFailed(msg string) {
	C.LoginFailed(C.CString(msg))
}

func bridgeInitFinished() {
	C.InitFinished()
}

//export loadToken
func loadToken() *C.char {
	token := loadTokenPath()
	if _, err := os.Stat(token); os.IsNotExist(err) {
		panic("cgo -> loadToken() when no token found.");
	} else {
		token, err := readToken(token)
		if err != nil {
			fmt.Println(err)
			panic("cgo -> loadToken() when no token found.");
		}
		return C.CString(token)
	}
}

//export generateTempToken
func generateTempToken(username string, password string) *C.char {
	tokenstruct := &FBPlainToken{
		EncryptToken: true,
		Username:     username,
		Password:     password,
	}
	token, err := json.Marshal(tokenstruct)
	if err != nil {
		panic("Failed to marshal json for generating temp token.")
	}
	return C.CString(string(token))
}

//export hasToken
func hasToken() bool {
	token := loadTokenPath()
	if _, err := os.Stat(token); os.IsNotExist(err) {
		return false
	} else {
		_, err := readToken(token)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	return true
}

var IsUnderLib bool=false

//export setCallbacks
func setCallbacks(cbs unsafe.Pointer) {
	C.SetCallbackSheet(cbs)
}

//export runLibClient
func runLibClient(_guiversion *C.char, _token *C.char, _serverCode *C.char, _serverPassword *C.char) {
	guiversion:=C.GoString(_guiversion)
	token:=C.GoString(_token)
	serverCode:=C.GoString(_serverCode)
	serverPassword:=C.GoString(_serverPassword)
	IsUnderLib=true
	defer func() {
		if err:=recover(); err!=nil {
			debug.PrintStack()
			C.ReportPanic(C.CString(fmt.Sprintf("%v",err)))
		}
		return
	} ()
	/*ex, err := os.Executable()
	if err != nil {
		panic(err)
	}*/
	ex:="phoenixbuilder-windows-shared.dll"
	version, err := utils.GetHash(ex)
	if err != nil {
		panic(err)
	}
	go func() {
		defer func() {
			if err:=recover(); err!=nil {
				debug.PrintStack()
				C.ReportPanic(C.CString(fmt.Sprintf("%v",err)))
			}
			return
		} ()
		if(len(os.Args)>1&&os.Args[1]=="--debug") {
			runDebugClient()
			return
		}
		runClient(token, version+"+"+guiversion, serverCode, serverPassword)
	} ()
}

//export GetFBVersion
func GetFBVersion() string {
	return FBVersion
}

//export isDebugMode
func isDebugMode() bool {
	if(len(os.Args)>1&&os.Args[1]=="--debug") {
		return true
	}
	return false
}

//export removeToken
func removeToken() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("WARNING - Failed to obtain the user's home directory. made homedir=\".\";")
		homedir="."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0755)
	token := filepath.Join(fbconfigdir,"fbtoken")
	os.Remove(token)
}

//export _executeFastBuilderCommand
func _executeFastBuilderCommand(cmd *C.char) {
	function.Process(bridgeConn, C.GoString(cmd))
}

//export _executeMinecraftCommand
func _executeMinecraftCommand(cmd *C.char) {
	command.SendSizukanaCommand(C.GoString(cmd), bridgeConn)
}

//export _sendChat
func _sendChat(content *C.char) {
	command.SendChat(C.GoString(content), bridgeConn)
}

//export _getBuildPos
func _getBuildPos() *C.char {
	pos:=configuration.GlobalFullConfig().Main().Position
	o:=fmt.Sprintf("%d, %d, %d",pos.X,pos.Y,pos.Z)
	return C.CString(o)
}

//export _getEndPos
func _getEndPos() *C.char {
	pos:=configuration.GlobalFullConfig().Main().End
	o:=fmt.Sprintf("%d, %d, %d",pos.X,pos.Y,pos.Z)
	return C.CString(o)
}

//export teardownFastBuilder
func teardownFastBuilder() {
	fmt.Printf("Quit correctly\n")
	if(bridgeConn!=nil) {
		bridgeConn.Close()
	}
	os.Exit(0)
}
