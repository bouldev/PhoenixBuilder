package signalhandler

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
			env.Stop()
			conn.Close()
			fmt.Printf("%s.\n", I18n.T(I18n.QuitCorrectly))
			env.WaitStopped()
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
		env.Stop()
		conn.Close()
		fmt.Printf("%s.\n", I18n.T(I18n.QuitCorrectly))
		env.WaitStopped()
		os.Exit(0)
	}()
}
