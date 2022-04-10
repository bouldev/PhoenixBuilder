package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/external/packet"
	"phoenixbuilder/minecraft/protocol"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"unsafe"
)

var DontGCMe []*connection.Client

func objAvailable(id int) (*connection.Client, error) {
	if id < 0 || id >= len(DontGCMe) {
		return nil, fmt.Errorf("id %v out of range %v\n", id, len(DontGCMe))
	}
	if c := DontGCMe[id]; c == nil {
		return nil, fmt.Errorf("id %v has been released", id)
	} else {
		return c, nil
	}
}

func toCErrStr(err error) *C.char {
	if err == nil {
		return nil
	}
	return C.CString(err.Error())
}

//export FreeMem
func FreeMem(address unsafe.Pointer) {
	C.free(address)
}

//export ConnectFB
func ConnectFB(address *C.char) (connID int, err *C.char) {
	str := C.GoString(address)
	// fmt.Println(str)
	client := connection.NewClient(str)
	if client == nil {
		return -1, C.CString("connect fail")
	}
	for i, c := range DontGCMe {
		if c == nil {
			DontGCMe[i] = client
			return i, nil
		}
	}
	i := len(DontGCMe)
	DontGCMe = append(DontGCMe, client)
	return i, nil
}

//export ReleaseConnByID
func ReleaseConnByID(id int) (err *C.char) {
	if _, _err := objAvailable(id); _err != nil {
		return C.CString(_err.Error())
	} else {
		DontGCMe[id] = nil
	}
	for {
		if len(DontGCMe) != 0 && DontGCMe[len(DontGCMe)-1] == nil {
			DontGCMe = DontGCMe[0 : len(DontGCMe)-1]
		} else {
			break
		}
	}
	return nil
}

func bytesToCharArr(goByteSlice []byte) *C.char {
	ptr := C.malloc(C.size_t(len(goByteSlice)))
	C.memmove(ptr, (unsafe.Pointer)(&goByteSlice[0]), C.size_t(len(goByteSlice)))
	return (*C.char)(ptr)
}

//export RecvGamePacket
func RecvGamePacket(connID int) (pktBytes *C.char, l int, err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		return nil, 0, C.CString(_err.Error())
	}
	bs, _err := obj.RecvGamePacket()
	if _err != nil {
		bs = []byte{}
		ReleaseConnByID(connID)
		return nil, 0, C.CString(_err.Error())
	}
	//fmt.Println(bs)
	return bytesToCharArr(bs), len(bs), nil
}

//export SendGamePacketBytes
func SendGamePacketBytes(connID int, content []byte) (err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return C.CString(_err.Error())
	}
	_err = obj.Send(&packet.GamePacket{Content: content})
	return toCErrStr(_err)
}

//export SendFBCommand
func SendFBCommand(connID int, cmd *C.char) (err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return C.CString(_err.Error())
	}
	_err = obj.SendFBCmd(C.GoString(cmd))
	return toCErrStr(_err)
}

//export SendWSCommand
func SendWSCommand(connID int, cmd *C.char) (uuid *C.char, err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return nil, C.CString(_err.Error())
	}
	uid, _err := obj.SendWSCmd(C.GoString(cmd))
	return C.CString(uid.String()), toCErrStr(_err)
}

//export SendMCCommand
func SendMCCommand(connID int, cmd *C.char) (uuid *C.char, err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return nil, C.CString(_err.Error())
	}
	uid, _err := obj.SendMCCmd(C.GoString(cmd))
	return C.CString(uid.String()), toCErrStr(_err)
}

//export SendNoResponseCommand
func SendNoResponseCommand(connID int, cmd *C.char) (err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return C.CString(_err.Error())
	}
	_err = obj.SendNoResponseMCCmd(C.GoString(cmd))
	return toCErrStr(_err)
}

type NoEOFByteReader struct {
	s []byte
	i int
}

func (nbr *NoEOFByteReader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	n = copy(b, nbr.s[nbr.i:])
	nbr.i += n
	return
}

func (nbr *NoEOFByteReader) ReadByte() (b byte, err error) {
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	b = nbr.s[nbr.i]
	nbr.i++
	return b, nil
}
func safeDecode(pktByte []byte) (pkt mc_packet.Packet) {
	pktID := uint32(pktByte[0])
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(pktID, "decode fail")
		}
		return
	}()
	pkt = connection.TypePool[pktID]()
	pkt.Unmarshal(protocol.NewReader(&NoEOFByteReader{s: pktByte[1:]}, 0))
	return
}

//export GamePacketBytesAsIsJsonStr
func GamePacketBytesAsIsJsonStr(pktBytes []byte) (jsonStr *C.char, err *C.char) {
	pk := safeDecode(pktBytes)
	marshal, _err := json.Marshal(pk)
	if _err != nil {
		return nil, C.CString(_err.Error())
	}
	return C.CString(string(marshal)), toCErrStr(_err)
}

//export JsonStrAsIsGamePacketBytes
func JsonStrAsIsGamePacketBytes(packetID int, jsonStr *C.char) (pktBytes *C.char, l int, err *C.char) {
	pk := connection.TypePool[uint32(packetID)]()
	_err := json.Unmarshal([]byte(C.GoString(jsonStr)), &pk)
	if _err != nil {
		return nil, 0, C.CString(_err.Error())
	}
	b := &bytes.Buffer{}
	w := protocol.NewWriter(b, 0)
	hdr := pk.ID()
	w.Varuint32(&hdr)
	pk.Marshal(w)
	bs := b.Bytes()
	l = len(bs)
	return bytesToCharArr(bs), l, nil
}

func main() {
	//Windows: go build  -tags fbconn -o fbconn.dll -buildmode=c-shared main.go
	//Linux: go build -tags fbconn -o libfbconn.so -buildmode=c-shared main.go
	//Macos: go build -tags fbconn -o fbconn.dylib -buildmode=c-shared main.go
	//将生成的文件 (fbconn.dll 或 libfbconn.so 或 fbconn.dylib) 放在 conn.py 同一个目录下
}
