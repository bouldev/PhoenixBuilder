// +build !windows,!android android,!arm
// +build !no_readline

package readline

/*
#cgo CFLAGS: -I${SRCDIR}/../../depends/readline-master/include
#cgo LDFLAGS: -L${SRCDIR}/../../depends/readline-master/prebuilt -L${SRCDIR}/../../depends/ncurses-6.3-20220319/prebuilt
#cgo !android,!darwin,!ios,!windows LDFLAGS: -lreadline
#cgo android,arm LDFLAGS: -lreadline-armv7-android -lncurses-arm-android
#cgo android,arm64 LDFLAGS: -lreadline-aarch64-android -lncurses-aarch64-android
#cgo android,386 LDFLAGS: -lreadline-i686-android -lncurses-i686-android
#cgo android,amd64 LDFLAGS: -lreadline-x86_64-android -lncurses-x86_64-android
#cgo darwin,!ios,arm64 LDFLAGS: -lreadline-aarch64-macos -lncurses
#cgo darwin,amd64 LDFLAGS: -lreadline-x86_64-macos -lncurses
#cgo ios,arm64 LDFLAGS: -lreadline-aarch64-ios -lncurses
#cgo windows,386 LDFLAGS: -lreadline-i686-mingw32
#cgo windows,amd64 LDFLAGS: -lreadline-x86_64-mingw32
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
	C.do_interrupt()
}

func Interrupt() {
	C.do_sigint_interrupt()
}

func InitReadline() {
	C.init_readline()
}

func Readline(env *environment.PBEnvironment) string {
	currentFunctionHolder:=env.FunctionHolder.(*function.FunctionHolder)
	currentFunctionMap=currentFunctionHolder.FunctionMap
	readline_cstr:=C.doReadline()
	readline_gstr:=C.GoString(readline_cstr)
	C.free(unsafe.Pointer(readline_cstr))
	return readline_gstr
}