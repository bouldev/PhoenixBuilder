package signalhandler

import (
	"fmt"
	"os"
	"os/signal"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/minecraft"
	"syscall"
)

func Init(conn *minecraft.Conn) {
	go func() {
		defer func() {
			r:=recover()
			if r!=nil{
				fmt.Println("go routine @ fastbuilder.signalhandler crashed ",r)
			}
		}()
		signalchannel:=make(chan os.Signal)
		signal.Notify(signalchannel, os.Interrupt) // ^C
		signal.Notify(signalchannel, syscall.SIGTERM)
		signal.Notify(signalchannel, syscall.SIGQUIT) // ^\
		<-signalchannel
		conn.Close()
		fmt.Printf("%s.\n",I18n.T(I18n.QuitCorrectly))
		os.Exit(0)
	} ()
}