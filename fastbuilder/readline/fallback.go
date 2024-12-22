//go:build windows || (android && arm) || no_readline

package readline

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

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
