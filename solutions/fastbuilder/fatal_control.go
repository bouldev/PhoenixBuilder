package fastbuilder

import (
	"bufio"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/readline"
	"phoenixbuilder/omega/suggest"
	"runtime"
	"runtime/debug"

	I18n "phoenixbuilder/fastbuilder/i18n"

	"github.com/pterm/pterm"
	_ "unsafe"
)

var PassFatal bool = false

//go:linkname onFatal args_hook_on_fatal
func onFatal(string)

func Fatal() {
	if PassFatal {
		return
	}
	if err := recover(); err != nil {
		if !args.NoReadline {
			readline.HardInterrupt()
		}
		debug.PrintStack()
		pterm.Error.Println(I18n.T(I18n.Crashed_Tip))
		pterm.Error.Println(I18n.T(I18n.Crashed_StackDump_And_Error))
		pterm.Error.Println(err)
		if args.ShouldEnableOmegaSystem {
			omegaSuggest := suggest.GetOmegaErrorSuggest(fmt.Sprintf("%v", err))
			fmt.Print(omegaSuggest)
		}
		if runtime.GOOS == "windows" {
			pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		os.Exit(1)
	}
	os.Exit(0)
}
