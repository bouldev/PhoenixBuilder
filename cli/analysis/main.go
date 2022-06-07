package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/pterm/pterm"
)

func CreatePacketFeeder(fileName string) chan packet.Packet {
	packetChan := make(chan packet.Packet, 1024)
	TypePool := packet.NewPool()
	fp, err := os.OpenFile(fileName, os.O_RDONLY, 0755)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		panic(err)
	}
	safeDecode := func(pktByte []byte) (pkt packet.Packet) {
		pktID := uint32(pktByte[0])
		defer func() {
			if r := recover(); r != nil {
				defer func() {
					if r := recover(); r != nil {
						pterm.Error.Println(pktID, "decode fail ", pkt, " ", r)
					}
				}()
				pkt.Unmarshal(protocol.NewReader(bytes.NewReader(bytes.Join([][]byte{pktByte[1:], []byte{0}}, []byte{})), 0))
			}
		}()
		pkt = TypePool[pktID]()
		pkt.Unmarshal(protocol.NewReader(bytes.NewReader(pktByte[1:]), 0))
		return
	}
	pos := 0
	go func() {
		for pos < len(data) {
			dataLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
			if dataLen == 0 {
				continue
			}
			pos += 4
			if pos+dataLen > len(data) {
				break
			}
			pktByte := data[pos : pos+dataLen]
			pos += dataLen
			pkt := safeDecode(pktByte)
			if pkt != nil {
				packetChan <- pkt
			}
		}
		close(packetChan)
	}()
	return packetChan
}

func CheckPackets(fileName string) {
	pterm.Info.Println()
	fmt.Println(fileName)
	for pkt := range CreatePacketFeeder(fileName) {
		if pkt.ID() < 200 {
			continue
		}
		fmt.Println(pkt)
	}
}

func main() {
	fileName := ""
	fileName = "failtologin1.bin"
	CheckPackets(fileName)
	fileName = "failtologin2.bin"
	CheckPackets(fileName)
	fileName = "canlogin1.bin"
	CheckPackets(fileName)
	fileName = "canlogin2.bin"
	CheckPackets(fileName)
}
