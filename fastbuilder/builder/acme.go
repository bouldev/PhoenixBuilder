package builder

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"phoenixbuilder/bridge/bridge_fmt"
	bridge_path "phoenixbuilder/fastbuilder/builder/path"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
	"strconv"
	"strings"
)

func seekBuf(buf *bufio.Reader, seekn int) error {
	seeker := make([]byte, seekn)
	c, err := buf.Read(seeker)
	if c != seekn {
		if err == nil {
			return seekBuf(buf, seekn-c)
		}
		bridge_fmt.Printf("%v\n", err)
		return fmt.Errorf("Early EOF [SEEK]")
	}
	return err
}

func readBig(buf *bufio.Reader, out []byte) error {
	c, err := buf.Read(out)
	if c != len(out) {
		if err != nil {
			return err
		}
		return readBig(buf, out[c:])
	}
	return err
}

func Acme(config *types.MainConfig, blc chan *types.Module) error {
	file, err := bridge_path.ReadFile(config.Path)
	if err != nil {
		return I18n.ProcessSystemFileError(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	buf := bufio.NewReader(gz)
	headerbuf := make([]byte, 4)
	_, err = buf.Read(headerbuf)
	if err != nil {
		return fmt.Errorf("Early EOF[1]")
	}
	if string(headerbuf) != "MCAC" {
		return fmt.Errorf(I18n.T(I18n.NotAnACMEFile))
	}
	{
		versionField1, err := buf.ReadByte()
		versionField2, err := buf.ReadByte()
		if versionField1 != 1 || versionField2 != 2 {
			return fmt.Errorf(I18n.T(I18n.UnsupportedACMEVersion))
		}
		//seeker := make([]byte, 26)
		//_, err = buf.Read(seeker)
		err = seekBuf(buf, 26)
		if err != nil {
			return fmt.Errorf(I18n.T(I18n.ACME_FailedToSeek))
		}
	}
	blocksTable := make(map[string]*types.Block)
	blocksTableSet := false
	for {
		commandStrBuf, err := buf.ReadBytes(0x3a)
		if err != nil {
			return fmt.Errorf(I18n.T(I18n.ACME_FailedToGetCommand))
		}
		commandStr := string(commandStrBuf)
		if commandStr == "dict2strid_:" {
			jsonSizeBuffer := make([]byte, 8)
			c, err := buf.Read(jsonSizeBuffer)
			if err != nil || c != 8 {
				return fmt.Errorf("err?")
			}
			jsonSize := binary.BigEndian.Uint64(jsonSizeBuffer)
			jsonContent := make([]byte, jsonSize)
			err = readBig(buf, jsonContent)
			if err != nil {
				return fmt.Errorf("err?[2]err22")
			}
			var blocksJSON map[string]interface{}
			json.Unmarshal(jsonContent, &blocksJSON)
			for item := range blocksJSON {
				blArr, _ := blocksJSON[item].([]interface{})
				blName, _ := blArr[0].(string)
				blDataF, _ := blArr[1].(float64)
				curBlkSpl := strings.Split(blName, ":")
				blocksTable[item] = &types.Block{
					Name: &(curBlkSpl[1]),
					Data: uint16(blDataF),
				}
			}
			blocksTableSet = true
			continue
		} else if commandStr == "DM3Tab_1id_:" {
			err = seekBuf(buf, 20)
			if err != nil || !blocksTableSet {
				return fmt.Errorf("ERR-SEEK-DM3")
			}
			l1Buffer := make([]byte, 2)
			c, err := buf.Read(l1Buffer)
			if err != nil || c != 2 {
				return fmt.Errorf("ERR RSIZE DM3 l1")
			}
			l1 := int(binary.BigEndian.Uint16(l1Buffer))
			l2Buffer := make([]byte, 2)
			c, err = buf.Read(l2Buffer)
			if err != nil || c != 2 {
				return fmt.Errorf("ERR RSIZE DM3 l2")
			}
			l2 := int(binary.BigEndian.Uint16(l2Buffer))
			l3Buffer := make([]byte, 2)
			c, err = buf.Read(l3Buffer)
			if err != nil || c != 2 {
				return fmt.Errorf("ERR RSIZE DM3 l3")
			}
			l3 := int(binary.BigEndian.Uint16(l3Buffer))
			for p1 := 0; p1 < l1; p1++ {
				for p2 := 0; p2 < l2; p2++ {
					for p3 := 0; p3 < l3; p3++ {
						curBlockId, err := buf.ReadByte()
						if err != nil {
							return fmt.Errorf("%s: %v", I18n.T(I18n.ACME_StructureErrorNotice), err)
						}
						p := config.Position
						p.X += p1
						p.Y += p2
						p.Z += p3
						curBlock := blocksTable[strconv.Itoa(int(curBlockId))]
						if *curBlock.Name != "air" {
							blc <- &types.Module{Point: p, Block: curBlock}
						}
					}
				}
			}
			break
		} else {
			bridge_fmt.Printf("Unknown ACME command!! %s\n", commandStr)
			return fmt.Errorf(I18n.T(I18n.ACME_UnknownCommand))
		}
	}
	return nil
}
