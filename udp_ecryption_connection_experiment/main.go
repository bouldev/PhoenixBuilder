package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
	"udp_ecryption_connection_experiment/connection"
)

func startServer() {
	server := &connection.KCPConnetionServerHandler{}
	server.SetOnServerDown(func(r interface{}) {
		if r != nil {
			panic(fmt.Errorf("server down: %v", r))
		}
		fmt.Print("server down\n")
	})
	server.SetOnAcceptNewConnectionFail(func(e error) {
		fmt.Println("server accept new connection fail", e)
	})
	server.SetOnNewConnection(func(conn connection.ReliableConnetion) {
		fmt.Println("server accept new connection")
		go func() {
			for {
				data, err := conn.RecvFrame()
				if err != nil {
					fmt.Println("server recv fail", err)
					return
				}
				// fmt.Printf("server recv (%v):%v\n", len(data), string(data))
				err = conn.SendFrame(data)
				if err != nil {
					fmt.Println("server echo fail", err)
					return
				}
				// fmt.Println("server echo", string(data))
			}
		}()
	})
	if err := server.Listen("0.0.0.0:7000"); err != nil {
		panic(err)
	}
	time.Sleep(time.Second)
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func startClient() {
	conn, err := connection.KCPDial("localhost:7000")
	if err != nil {
		panic(err)
	}
	testEcho := func(data string) {
		send_data_bytes := []byte(data)

		// warning! SendFrame with encryption will change the data
		conn.SendFrame(send_data_bytes)
		recv_data_bytes, err := conn.RecvFrame()
		if err != nil {
			panic(err)
		}
		if !bytes.Equal([]byte(data), recv_data_bytes) {
			panic(fmt.Errorf("send:%v,recv:%v", data, string(recv_data_bytes)))
		}
		fmt.Println("test echo success")
	}
	testEcho("hello")
	testEcho("world")
	for i := 0; i < 19; i++ {
		fmt.Println(i)
		testEcho(String(1 << i))
	}
}

func startMuxServer() {
	server := &connection.KCPConnetionServerHandler{}
	server.SetOnServerDown(func(r interface{}) {
		if r != nil {
			panic(fmt.Errorf("server down: %v", r))
		}
		fmt.Print("server down\n")
	})
	server.SetOnAcceptNewConnectionFail(func(e error) {
		fmt.Println("server accept new connection fail", e)
	})
	server.SetOnNewConnection(func(conn connection.ReliableConnetion) {
		fmt.Println("server accept new connection")
		mux := connection.NewMux(conn)
		connectChan := mux.GetSubChannel('C')
		dataChan := mux.GetSubChannel('D')
		go func() {
			for {
				ctrl, err := connectChan.RecvFrame()
				// fmt.Println(string(ctrl))
				if err != nil {
					fmt.Println("connectChan recv fail", err)
					return
				}
				err = connectChan.SendFrame(ctrl)
				if err != nil {
					fmt.Println("connectChan echo fail", err)
					return
				}
			}
		}()
		go func() {
			for {
				data, err := dataChan.RecvFrame()
				// fmt.Println(string(data))
				if err != nil {
					fmt.Println("dataChan recv fail", err)
					return
				}
				err = dataChan.SendFrame(data)
				if err != nil {
					fmt.Println("dataChan echo fail", err)
					return
				}
			}
		}()
	})
	if err := server.Listen("0.0.0.0:7001"); err != nil {
		panic(err)
	}
	time.Sleep(time.Second)
}

func startMuxClient() {
	conn, err := connection.KCPDial("localhost:7001")
	if err != nil {
		panic(err)
	}
	mux := connection.NewMux(conn)
	ctrlChan := mux.GetSubChannel('C')
	dataChan := mux.GetSubChannel('D')
	testEcho := func(data string, conn connection.ReliableConnetion) {
		send_data_bytes := []byte(data)

		// warning! SendFrame with encryption will change the data
		conn.SendFrame(send_data_bytes)
		recv_data_bytes, err := conn.RecvFrame()
		if err != nil {
			panic(err)
		}
		if !bytes.Equal([]byte(data), recv_data_bytes) {
			panic(fmt.Errorf("send:%v,recv:%v", data, string(recv_data_bytes)))
		}
		fmt.Println("test echo success")
	}
	testEcho("ctrl hello", ctrlChan)
	testEcho("data world", dataChan)
	for i := 1; i < 19; i++ {
		fmt.Println(i)
		testEcho(String(i), ctrlChan)
		testEcho(String(1<<i), dataChan)
	}
}

func main() {
	startServer()
	startClient()
	startMuxServer()
	startMuxClient()
}
