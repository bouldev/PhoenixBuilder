package signalhandler

import (
	"fmt"
	"os"
	"os/signal"
	"phoenixbuilder/minecraft"
	"syscall"
)

func Init(conn *minecraft.Conn) {
	go func() {
		signalchannel:=make(chan os.Signal)
		signal.Notify(signalchannel, os.Interrupt) // ^C
		signal.Notify(signalchannel, syscall.SIGTERM)
		signal.Notify(signalchannel, syscall.SIGQUIT) // ^\
		<-signalchannel
		conn.Close()
		fmt.Printf("Quit correctly.\n")
		os.Exit(0)
	} ()
}