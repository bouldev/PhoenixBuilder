package builder

import (
	"compress/gzip"
	"phoenixbuilder/minecraft/mctype"
	"bufio"
	"os"
	"fmt"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"strings"
)

func seekBuf(buf *bufio.Reader,seekn int) error {
	seeker := make([]byte, seekn)
	c, err := buf.Read(seeker)
	if c!=seekn {
		return fmt.Errorf("Early EOF [SEEK]")
	}
	return err
}

func Acme(config mctype.MainConfig) ([]mctype.Module, error) {
	file, err := os.Open(config.Path)
	if err != nil {
		return []mctype.Module{}, err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return []mctype.Module{}, err
	}
	defer gz.Close()
	buf := bufio.NewReader(gz)
	headerbuf := make([]byte, 4)
	_, err = buf.Read(headerbuf)
	if err != nil {
		return []mctype.Module{}, fmt.Errorf("Early EOF[1]")
	}
	if string(headerbuf) != "MCAC" {
		return []mctype.Module{}, fmt.Errorf("Invalid file, not an ACME structure.")
	}
	{
		versionField1, err := buf.ReadByte()
		versionField2, err := buf.ReadByte()
		if versionField1 != 1 || versionField2 != 2 {
			return []mctype.Module{}, fmt.Errorf("Unsupported ACME structure version.Only acme file version 1.2 is supported.")
		}
		//seeker := make([]byte, 26)
		//_, err = buf.Read(seeker)
		err = seekBuf(buf,26)
		if err != nil {
			return []mctype.Module{}, fmt.Errorf("Failed to seek.Invalid acme file.")
		}
	}
	var blocksTable map[string]interface{}
	blocksTableSet := false
	var BlockSet []mctype.Module
	for {
		commandStrBuf, err := buf.ReadBytes(0x3a)
		if err != nil {
			return []mctype.Module{}, fmt.Errorf("Failed to get acme command.")
		}
		commandStr := string(commandStrBuf)
		if commandStr=="dict2strid_:" {
			jsonSizeBuffer:=make([]byte,8)
			c,err:=buf.Read(jsonSizeBuffer)
			if err != nil || c != 8 {
				return []mctype.Module{}, fmt.Errorf("err?")
			}
			jsonSize := binary.BigEndian.Uint64(jsonSizeBuffer)
			jsonContent := make([]byte, jsonSize)
			c, err=buf.Read(jsonContent)
			if err != nil || c != int(jsonSize) {
				return []mctype.Module{}, fmt.Errorf("err?[2]")
			}
			json.Unmarshal(jsonContent,&blocksTable)
			blocksTableSet=true
			continue
		}else if commandStr=="DM3Tab_1id_:" {
			err=seekBuf(buf,20)
			if err != nil || !blocksTableSet {
				return []mctype.Module{}, fmt.Errorf("ERR-SEEK-DM3")
			}
			l1Buffer:=make([]byte,2)
			c,err:=buf.Read(l1Buffer)
			if err!=nil || c!=2 {
				return []mctype.Module{}, fmt.Errorf("ERR RSIZE DM3 l1")
			}
			l1:=int(binary.BigEndian.Uint16(l1Buffer))
			l2Buffer:=make([]byte,2)
			c,err=buf.Read(l2Buffer)
			if err!=nil || c!=2 {
				return []mctype.Module{}, fmt.Errorf("ERR RSIZE DM3 l2")
			}
			l2:=int(binary.BigEndian.Uint16(l2Buffer))
			l3Buffer:=make([]byte,2)
			c,err=buf.Read(l3Buffer)
			if err!=nil || c!=2 {
				return []mctype.Module{}, fmt.Errorf("ERR RSIZE DM3 l3")
			}
			l3:=int(binary.BigEndian.Uint16(l3Buffer))
			for p1:=0;p1<l1;p1++ {
				for p2:=0;p2<l2;p2++ {
					for p3:=0;p3<l3;p3++ {
						curBlockId,err:=buf.ReadByte()
						if err != nil {
							return []mctype.Module{},fmt.Errorf("Structure error: %v",err)
						}
						p := config.Position
						p.X += p1
						p.Y += p2
						p.Z += p3
						curBlock,_:=blocksTable[strconv.Itoa(int(curBlockId))].([]interface{})
						curBlockName,_:=curBlock[0].(string)
						curBlockData,_:=curBlock[1].(float64)
						curBlkSpl:=strings.Split(curBlockName,":")
						b:=mctype.Block {
							Name:curBlkSpl[1],
							Data:int(curBlockData),
						}
						if b.Name!="air" {
							BlockSet = append(BlockSet,mctype.Module{Point: p, Block: b})
						}
					}
				}
			}
			break
		}else{
			fmt.Println("Unknown ACME command!! %s",commandStr)
			return []mctype.Module{}, fmt.Errorf("Unknown ACME command: %s",commandStr)
		}
	}
	return BlockSet, nil
}