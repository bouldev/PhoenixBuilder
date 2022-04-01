// +build windows android,arm no_readline

package readline

// Windows is not supported, so read from terminal directly.

import (
	"os"
	"fmt"
	"bufio"
	"strings"
	"phoenixbuilder/fastbuilder/environment"
)

var SelfTermination chan bool

func HardInterrupt() {
}

func Interrupt() {
	// No readline so exit directly.
	os.Exit(0)
}

func InitReadline() {
	fmt.Printf("Warning: Feature readline is not compatible with platform Windows.\n")
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