package main

import (
	"C"
	"fmt"
	"phoenixbuilder/fastbuilder/external/connection"
)

var DontGCMe []connection.ReliableConnection

func getConnID(conn connection.ReliableConnection) int {
	for i, conn := range DontGCMe {
		if conn == nil {
			return i
		}
	}
	i := len(DontGCMe)
	DontGCMe = append(DontGCMe, conn)
	return i
}

//export ReleaseConnByID
func ReleaseConnByID(id int) {
	if id >= len(DontGCMe) {
		fmt.Printf("id %v out of range %v\n", id, len(DontGCMe))
		return
	}
	conn := DontGCMe[id]
	if conn != nil {
		DontGCMe[id] = nil
	}
	for {
		if len(DontGCMe) != 0 && DontGCMe[len(DontGCMe)-1] == nil {
			DontGCMe = DontGCMe[0 : len(DontGCMe)-1]
		}
	}
}

func startClient(address string) (int, error) {
	conn, err := connection.KCPDial(address)
	if err != nil {
		return 0, err
	}
	return getConnID(conn), nil
}

//export ConnectFB
func ConnectFB(address *C.char) int {
	str := C.GoString(address)
	// fmt.Println(str)
	connID, err := startClient(str)
	if err != nil {
		// fmt.Printf("Connection Fail")
		// fmt.Println(err)
		connID = -1
	}
	// fmt.Println("Connection Success")
	return connID
}

//export RecvFrame
func RecvFrame(connID int) *C.char {
	bs, err := DontGCMe[connID].RecvFrame()
	// fmt.Println(bs, err)
	if err != nil {
		// fmt.Println(err)
		bs = []byte{}
	} else {
		bs = bs[1:]
	}
	return C.CString(string(bs))
}

func main() {
	//go build -o fb_conn.so -buildmode=c-shared main.go
}
