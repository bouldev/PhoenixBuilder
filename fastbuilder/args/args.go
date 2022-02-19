package args

/*
extern char args_isDebugMode;
extern char replaced_auth_server;
extern char *newAuthServer;
extern char args_disableHashCheck;
extern char args_muteWorldChat;
extern char args_noPyRpc;
extern char args_noNBT;
*/
import "C"

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

