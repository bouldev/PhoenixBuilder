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
extern char args_disableVersionCheck;
extern char enable_omega_system;
extern char startup_script;
extern char server_code;
extern char server_password;
extern char token_content;
extern char externalListenAddr;
extern char capture_output_file;
extern char args_no_readline;
extern char pack_scripts;
extern char pack_scripts_out;
extern char custom_gamename;
extern char ingame_response;
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
	print(C.args_disableVersionCheck)
	print(C.enable_omega_system)
	print(C.startup_script)
	print(C.server_code)
	print(C.server_password)
	print(C.token_content)
	print(C.externalListenAddr)
	print(C.capture_output_file)
	print(C.args_no_readline)
	print(C.pack_scripts)
	print(C.pack_scripts_out)
	print(C.custom_gamename)
	print(C.ingame_response)
}

var FBVersion string=*(*string)(unsafe.Pointer(&__cgo_args_var_fbversion_struct))
var FBPlainVersion string=*(*string)(unsafe.Pointer(&__cgo_args_var_fbplainversion_struct))
var FBCommitHash string=*(*string)(unsafe.Pointer(&__cgo_args_fb_commit_struct))
var DebugMode bool=*(*bool)(unsafe.Pointer(&__cgo_args_isDebugMode))
var AuthServer string=*(*string)(unsafe.Pointer(&__cgo_newAuthServer))
var ShouldDisableVersionCheck=*(*bool)(unsafe.Pointer(&__cgo_args_disableVersionCheck))
var ShouldEnableOmegaSystem=*(*bool)(unsafe.Pointer(&__cgo_enable_omega_system))
var StartupScript=*(*string)(unsafe.Pointer(&__cgo_startup_script))

//go:linkname SpecifiedServer args_has_specified_server
func SpecifiedServer() bool

var ServerCode=*(*string)(unsafe.Pointer(&__cgo_server_code))
var ServerPassword=*(*string)(unsafe.Pointer(&__cgo_server_password))

//go:linkname SpecifiedToken args_specified_token
func SpecifiedToken() bool

var CustomTokenContent=*(*string)(unsafe.Pointer(&__cgo_token_content))

var CustomSEConsts map[string]string = map[string]string{}
var CustomSEUndefineConsts []string = []string{}

//export custom_script_engine_const
func custom_script_engine_const(key, val *C.char) {
	CustomSEConsts[C.GoString(key)] = C.GoString(val)
}

//export do_suppress_se_const
func do_suppress_se_const(key *C.char) {
	CustomSEUndefineConsts = append(CustomSEUndefineConsts, C.GoString(key))
}

var ExternalListenAddress=*(*string)(unsafe.Pointer(&__cgo_externalListenAddr))
var CaptureOutputFile=*(*string)(unsafe.Pointer(&__cgo_capture_output_file))
var NoReadline=*(*bool)(unsafe.Pointer(&__cgo_args_no_readline))
var PackScripts=*(*string)(unsafe.Pointer(&__cgo_pack_scripts))
var PackScriptsOut=*(*string)(unsafe.Pointer(&__cgo_pack_scripts_out))
var CustomGameName=*(*string)(unsafe.Pointer(&__cgo_custom_gamename))
var InGameResponse=*(*bool)(unsafe.Pointer(&__cgo_ingame_response))

//export go_rmdir_recursive
func go_rmdir_recursive(path *C.char) {
	err:=os.RemoveAll(C.GoString(path))
	if err!=nil {
		panic(err)
	}
}
