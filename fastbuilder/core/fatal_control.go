package core

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

import (
	"bufio"
	"os"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/readline"
	"runtime"
	"runtime/debug"

	I18n "phoenixbuilder/fastbuilder/i18n"

	_ "unsafe"

	"github.com/pterm/pterm"
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
		if runtime.GOOS == "windows" {
			pterm.Error.Println(I18n.T(I18n.Crashed_OS_Windows))
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		os.Exit(1)
	}
	os.Exit(0)
}
