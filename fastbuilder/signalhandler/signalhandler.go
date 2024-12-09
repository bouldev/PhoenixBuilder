package signalhandler

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
	"fmt"
	"os"
	"os/signal"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/readline"
	"phoenixbuilder/minecraft"
	"syscall"
)

func Install(conn *minecraft.Conn, env *environment.PBEnvironment) {
	if(!args.NoReadline) {
		go func() {
			readline.SelfTermination = make(chan bool)
			<-readline.SelfTermination
			readline.HardInterrupt()
			conn.Close()
			fmt.Printf("%s.\n", I18n.T(I18n.QuitCorrectly))
			os.Exit(0)
		}()
		go func() {
			for {
				sigintchannel := make(chan os.Signal)
				signal.Notify(sigintchannel, os.Interrupt) // ^C
				<-sigintchannel
				readline.Interrupt()
			}
		}()
	}
	go func() {
		signalchannel := make(chan os.Signal)
		signal.Notify(signalchannel, syscall.SIGTERM)
		signal.Notify(signalchannel, syscall.SIGQUIT) // ^\
		if args.NoReadline {
			signal.Notify(signalchannel, os.Interrupt)
		}
		<-signalchannel
		readline.HardInterrupt()
		conn.Close()
		fmt.Printf("%s.\n", I18n.T(I18n.QuitCorrectly))
		os.Exit(0)
	}()
}
