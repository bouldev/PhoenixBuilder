package args

import (
	"os"
	"unsafe"
)
/*
extern char args_var_fbversion_struct;
extern char args_var_fbplainversion_struct;
extern char args_fb_commit_struct;
extern char args_isDebugMode;
extern char newAuthServer;
extern char args_disableVersionChecking;
extern char server_code;
extern char server_password;
extern char token_content;
extern char args_no_readline;
extern char custom_gamename;
extern char ingame_response;
extern char listen_address;
*/
import "C"

// ^ cgo_import_static is disallowed for normal go files,
// so we have to use fake definitions to take advantage of cmd/cgo

func referenceHolder() {
	// This won't really be called, but is here for honoring those C variables
	// Don't ever try calling this, that'd be horrible
	print(C.args_var_fbversion_struct)
	print(C.args_var_fbplainversion_struct)
	print(C.args_fb_commit_struct)
	print(C.args_isDebugMode)
	print(C.newAuthServer)
	print(C.args_disableVersionChecking)
	print(C.server_code)
	print(C.server_password)
	print(C.token_content)
	print(C.args_no_readline)
	print(C.custom_gamename)
	print(C.ingame_response)
	print(C.listen_address)
}

var FBVersion string=*(*string)(unsafe.Pointer(&__cgo_args_var_fbversion_struct))
var FBPlainVersion string=*(*string)(unsafe.Pointer(&__cgo_args_var_fbplainversion_struct))
var FBCommitHash string=*(*string)(unsafe.Pointer(&__cgo_args_fb_commit_struct))
var DebugMode bool=*(*bool)(unsafe.Pointer(&__cgo_args_isDebugMode))
var AuthServer string=*(*string)(unsafe.Pointer(&__cgo_newAuthServer))
var ShouldDisableVersionChecking=*(*bool)(unsafe.Pointer(&__cgo_args_disableVersionChecking))

//go:linkname SpecifiedServer args_has_specified_server
func SpecifiedServer() bool

var ServerCode=*(*string)(unsafe.Pointer(&__cgo_server_code))
var ServerPassword=*(*string)(unsafe.Pointer(&__cgo_server_password))

//go:linkname SpecifiedToken args_specified_token
func SpecifiedToken() bool

var CustomTokenContent=*(*string)(unsafe.Pointer(&__cgo_token_content))

var NoReadline=*(*bool)(unsafe.Pointer(&__cgo_args_no_readline))
var CustomGameName=*(*string)(unsafe.Pointer(&__cgo_custom_gamename))
var InGameResponse=*(*bool)(unsafe.Pointer(&__cgo_ingame_response))
var ListenAddress=*(*string)(unsafe.Pointer(&__cgo_listen_address))

//export go_rmdir_recursive
func go_rmdir_recursive(path *C.char) {
	err:=os.RemoveAll(C.GoString(path))
	if err!=nil {
		panic(err)
	}
}
