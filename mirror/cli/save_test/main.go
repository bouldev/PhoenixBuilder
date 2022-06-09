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
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/mcdb"
	"time"

	"github.com/df-mc/goleveldb/leveldb/opt"
)

type ChunkPacket struct {
	X, Z           int32
	SubChunksCount uint32
	Payload        []byte
}

func ReadDumpPackets(fileName string) (chunkPackets []*ChunkPacket) {
	TypePool := packet.NewPool()
	fp, err := os.OpenFile(fileName, os.O_RDONLY, 0755)
	if err != nil {
		panic(err)
	}
	chunkPackets = []*ChunkPacket{}
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		panic(err)
	}

	safeDecode := func(pktByte []byte) (pkt packet.Packet) {
		pktID := uint32(pktByte[0])
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(pktID, "decode fail ", pkt)
			}
		}()
		pkt = TypePool[pktID]()
		pkt.Unmarshal(protocol.NewReader(bytes.NewReader(pktByte[1:]), 0))
		return
	}
	pos := 0
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
		if pktByte[0] != packet.IDLevelChunk {
			continue
		}
		pkt := safeDecode(pktByte)
		if pkt != nil {
			switch p := pkt.(type) {
			case *packet.LevelChunk:
				chunkPackets = append(chunkPackets, &ChunkPacket{X: p.ChunkX, Z: p.ChunkZ, SubChunksCount: p.SubChunkCount, Payload: p.RawPayload})
			}
		}
	}
	return chunkPackets
}

func main() {
	chunkPackets := ReadDumpPackets("chunk_packets.bin")
	decodedChunks := []*mirror.ChunkData{}
	for _, chunkPacket := range chunkPackets {
		c, nbts, err := chunk.NEMCNetworkDecode(chunkPacket.Payload, int(chunkPacket.SubChunksCount))

		if err != nil {
			panic(err)
		}
		decodedChunks = append(decodedChunks, &mirror.ChunkData{
			Chunk: c, BlockNbts: nbts,
			ChunkPos:  define.ChunkPos{chunkPacket.X, chunkPacket.Z},
			TimeStamp: time.Now().Unix(),
		})
	}
	provider, err := mcdb.New("testout", opt.FlateCompression)
	if err != nil {
		panic(err)
	}
	provider.D.LevelName = "TestOut"
	for _, chunkData := range decodedChunks {
		fmt.Println("saving chunk @ ", chunkData.ChunkPos.X()<<4, chunkData.ChunkPos.Z()<<4)
		if chunkData == nil {
			fmt.Println("nil chunk")
		}
		err := provider.Write(chunkData)
		if err != nil {
			panic(err)
		}
	}
	provider.Close()
}
