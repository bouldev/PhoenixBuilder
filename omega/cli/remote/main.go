package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/uqHolder"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/mainframe"
)

type RemoteConnection struct {
	initUQ            *uqHolder.UQHolder
	conn              *connection.Client
	onConnectionError func()
}

type RemoteCliAdaptor struct {
	*RemoteConnection
	BackendCmdFeeder chan string
}

func (rca *RemoteCliAdaptor) GetBackendCommandFeeder() chan string {
	return rca.BackendCmdFeeder
}

func (rca *RemoteCliAdaptor) FBEval(cmd string) {
	rca.conn.SendFBCmd(cmd)
}

func (rca *RemoteCliAdaptor) pumpInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, _, _ := reader.ReadLine()
		rca.BackendCmdFeeder <- string(line)
	}
}

func (rc *RemoteConnection) Read() mc_packet.Packet {
	if pkt, err := rc.conn.RecvDecodedGamePacket(); err != nil {
		if rc.conn.IsClosed() {
			fmt.Println("Connection Error: " + err.Error())
			rc.onConnectionError()
		}
		fmt.Println("Error: " + err.Error())
		return nil
	} else {
		return pkt
	}
}

func (rc *RemoteConnection) Write(pkt mc_packet.Packet) {
	err := rc.conn.SendMCPacket(pkt)
	if err != nil {
		if rc.conn.IsClosed() {
			fmt.Println("Connection Error: " + err.Error())
			rc.onConnectionError()
		}
		fmt.Println("Error: " + err.Error())
		return
	}
}

func (rc *RemoteConnection) GetInitUQHolderCopy() *uqHolder.UQHolder {
	return rc.initUQ
}

func main() {
	// make connection
	// 这里以远程连接的方式工作
	// 但是，理论上，移植到 fb 内嵌只需要实现 adaptor 声明的 interface 即可
	conn := connection.NewClient("localhost:3456")
	if conn != nil {
		fmt.Println("Connect Success")
	} else {
		panic("Connection Fail")
	}
	conn.ReducePacket(111, 20)
	conn.ReducePacket(39, 20)
	conn.ReducePacket(40, 20)
	uqholder, err := conn.RequestUQHolder("*")
	if err != nil {
		panic(err)
	}

	// wrapper interface
	adaptor := &RemoteCliAdaptor{
		RemoteConnection: &RemoteConnection{
			initUQ: uqholder,
			conn:   conn,
			onConnectionError: func() {
				panic("Connection Error")
			},
		},
		BackendCmdFeeder: make(chan string),
	}
	go adaptor.pumpInput()

	// create mainframe
	var mainFrame defines.System
	mainFrame = mainframe.NewOmega()

	// configuration
	stop := func() {
		mainFrame.Stop()
		os.Exit(0)
	}
	adaptor.onConnectionError = stop
	defer stop()

	// activate
	mainFrame.Bootstrap(adaptor)
	go mainFrame.Activate()

	// catch signal
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	fmt.Println("Aborting...")
}
