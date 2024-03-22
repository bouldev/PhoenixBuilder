//go:build windows || (android && arm) || no_readline

package readline

// Windows is not supported, so read from terminal directly.

import (
	"bufio"
	"os"
	"phoenixbuilder/fastbuilder/environment"
	"strings"

	"github.com/pterm/pterm"
)

var SelfTermination chan bool

func HardInterrupt() {
}

func Interrupt() {
	// No readline so exit directly.
	os.Exit(0)
}

func InitReadline() {
	pterm.Warning.Println("Feature readline is not compatible with current platform.")
}

func Readline(env *environment.PBEnvironment) string {
	reader := bufio.NewReader(os.Stdin)
	inp, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	inpl := strings.TrimRight(inp, "\r\n")
	return inpl
}
