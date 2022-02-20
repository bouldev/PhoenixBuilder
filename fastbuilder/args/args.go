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

extern void parse_args(int argc, char **argv);

extern char *get_fb_version(void);
*/
import "C"

func GetFBVersion() string {
	return C.GoString(C.get_fb_version())
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

func ShouldMuteWorldChat() bool {
	return boolify(C.args_muteWorldChat)
}

func NoPyRpc() bool {
	return boolify(C.args_noPyRpc)
}

func NoNBT() bool {
	return boolify(C.args_noNBT)
}

