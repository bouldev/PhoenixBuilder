package args

import (
	"os"
	"unsafe"
)

/*
extern void free(void *);

extern char args_isDebugMode;
extern char replaced_auth_server;
extern char *newAuthServer;
extern char args_disableHashCheck;
extern char args_muteWorldChat;
extern char args_noPyRpc;
extern char args_noNBT;
extern char *startup_script;

extern void parse_args(int argc, char **argv);

extern char use_startup_script;
extern char *get_fb_version(void);
extern char *get_fb_plain_version(void);
extern char *commit_hash(void);
*/
import "C"

func charify(val bool) C.char {
	if(val) {
		return C.char(1)
	}else{
		return C.char(0)
	}
}

func Set_args_isDebugMode(val bool) {
	C.args_isDebugMode=charify(val)
}

func Do_replace_authserver(val string) {
	if(boolify(C.replaced_auth_server)) {
		C.free(unsafe.Pointer(C.newAuthServer))
	}else{
		C.replaced_auth_server=C.char(1)
	}
	C.newAuthServer=C.CString(val)
}

func Set_disableHashCheck(val bool) {
	C.args_disableHashCheck=charify(val)
}

func Set_muteWorldChat(val bool) {
	C.args_muteWorldChat=charify(val)
}

func Set_noPyRpc(val bool) {
	C.args_noPyRpc=charify(val);
}

func Set_noNBT(val bool) {
	C.args_noNBT=charify(val);
}

func GetFBVersion() string {
	return C.GoString(C.get_fb_version())
}

func GetFBPlainVersion() string {
	return C.GoString(C.get_fb_plain_version())
}

func GetFBCommitHash() string {
	return C.GoString(C.commit_hash())
}

func ParseArgs() {
	argv:=make([]*C.char, len(os.Args))
	for i, v:=range os.Args {
		cstr:=C.CString(v)
		defer C.free(unsafe.Pointer(cstr))
		argv[i]=cstr
	}
	C.parse_args(C.int(len(os.Args)),&argv[0])
}

func boolify(v C.char) bool {
	if int(v)==0 {
		return false
	}
	return true
}

func DebugMode() bool {
	if int(C.args_isDebugMode)==0 {
		return false
	}
	return true
}

func AuthServer() string {
	if int(C.replaced_auth_server)==0 {
		return "wss://api.fastbuilder.pro:2053/"
	}
	return C.GoString(C.newAuthServer)
}

func ShouldDisableHashCheck() bool {
	return boolify(C.args_disableHashCheck)
}

func SetShouldDisableHashCheck() {
	C.args_disableHashCheck=C.char(1)
}

func ShouldMuteWorldChat() bool {
	return boolify(C.args_muteWorldChat)
}

func NoPyRpc() bool {
	return boolify(C.args_noPyRpc)
}

func NoNBT() bool {
	return boolify(C.args_noNBT)
}

func StartupScript() string {
	if  int(C.use_startup_script)==0 {
		return ""
	}
	return C.GoString(C.startup_script)
}
