package command

import (
	"io"
	"fmt"
	"encoding/json"
	"encoding/binary"
)

// See [#107](https://github.com/LNSSPsd/PhoenixBuilder/issues/107) for details

type PlaceBlockWithBlockStates struct {
	BlockConstantStringID uint16
	BlockStatesJSONString string
}

func (_ *PlaceBlockWithBlockStates) ID() uint16 {
	return 13
}

func (_ *PlaceBlockWithBlockStates) Name() string {
	return "PlaceBlockWithBlockStatesCommand"
}

func (cmd *PlaceBlockWithBlockStates) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.BlockConstantStringID)
	_, err:=writer.Write(buf)
	if err!=nil {
		return err
	}
	_, err=writer.Write([]byte(cmd.BlockStatesJSONString))
	return err
}

func (cmd *PlaceBlockWithBlockStates) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.BlockConstantStringID=binary.BigEndian.Uint16(buf)
	blockStates, err:=readString(reader)
	if err!=nil {
		return err
	}
	if !json.Valid([]byte(blockStates)) {
		return fmt.Errorf("Invalid blockStates JSON (Not a valid JSON string)")
	}
	cmd.BlockStatesJSONString=blockStates
	return nil
}