// +build !windows,!android android,!arm
// +build !no_readline

package readline

/*
#cgo !native CFLAGS: -I${SRCDIR}/../../depends/readline-master/include
#cgo !native LDFLAGS: -L${SRCDIR}/../../depends/stub -L${SRCDIR}/../../depends/readline-master/prebuilt -L${SRCDIR}/../../depends/ncurses-6.3-20220319/prebuilt
#cgo native,darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/readline/lib -lreadline
#cgo native,!darwin LDFLAGS: -lreadline
#cgo !native,!darwin,!ios,!windows,!android LDFLAGS: -lreadline
#cgo !native,android_shared LDFLAGS: -lreadline
#cgo !native,android,!android_shared,arm LDFLAGS: -lreadline-armv7-android -lncurses-arm-android
#cgo !native,android,!android_shared,arm64 LDFLAGS: -lreadline-aarch64-android -lncurses-aarch64-android
#cgo !native,android,!android_shared,386 LDFLAGS: -lreadline-i686-android -lncurses-i686-android
#cgo !native,android,!android_shared,amd64 LDFLAGS: -lreadline-x86_64-android -lncurses-x86_64-android
#cgo !native,darwin,!ios,arm64 LDFLAGS: -lreadline-aarch64-macos -lncurses
#cgo !native,darwin,amd64 LDFLAGS: -lreadline-x86_64-macos -lncurses
#cgo !native,ios,arm64 LDFLAGS: -lreadline-aarch64-ios -lncurses
#cgo netbsd LDFLAGS: -lreadline -lterminfo
extern char **strengthenStringArray(const char **source, int entries);
extern char *doReadline();
extern void init_readline();
extern void free(void *pointer);
extern void do_interrupt();
extern void do_sigint_interrupt();
*/
import "C"

import (
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/environment"
	"unsafe"
)

var isBootstrapped bool=false
var SelfTermination chan bool

var currentFunctionMap map[string]*function.Function

//export GetFunctionList
func GetFunctionList() **C.char {
	functionList:=make([]*C.char, len(currentFunctionMap))
	funcListIdx:=0
	for funcname, _ := range currentFunctionMap {
		functionList[funcListIdx]=C.CString(funcname)
		funcListIdx++
	}
	return C.strengthenStringArray(&functionList[0],C.int(len(currentFunctionMap)))
}

//export teardown_self
func teardown_self() {
	SelfTermination<-true
	return
}

func HardInterrupt() {
	if(!isBootstrapped) {
		return
	}
	C.do_interrupt()
}

func Interrupt() {
	if(!isBootstrapped) {
		return
	}
	C.do_sigint_interrupt()
}

func InitReadline() {
	C.init_readline()
}

func Readline(env *environment.PBEnvironment) string {
	if(!isBootstrapped) {
		isBootstrapped=true
	}
	currentFunctionHolder:=env.FunctionHolder.(*function.FunctionHolder)
	currentFunctionMap=currentFunctionHolder.FunctionMap
	readline_cstr:=C.doReadline()
	readline_gstr:=C.GoString(readline_cstr)
	C.free(unsafe.Pointer(readline_cstr))
	return readline_gstr
}
