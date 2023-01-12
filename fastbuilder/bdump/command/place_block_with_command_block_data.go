package command

import (
	"io"
	"encoding/binary"
	"phoenixbuilder/fastbuilder/types"
)

type PlaceBlockWithCommandBlockData struct {
	BlockConstantStringID uint16
	BlockData uint16
	CommandBlockData *types.CommandBlockData
}

func (_ *PlaceBlockWithCommandBlockData) ID() uint16 {
	return 27
}

func (_ *PlaceBlockWithCommandBlockData) Name() string {
	return "PlaceBlockWithCommandBlockDataCommand"
}

func (cmd *PlaceBlockWithCommandBlockData) Marshal(writer io.Writer) error {
	uint16_buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(uint16_buf, cmd.BlockConstantStringID)
	_, err:=writer.Write(uint16_buf)
	if err!=nil {
		return err
	}
	binary.BigEndian.PutUint16(uint16_buf, cmd.BlockData)
	_, err=writer.Write(uint16_buf)
	if err!=nil {
		return err
	}
	uint32_buf:=make([]byte, 4)
	binary.BigEndian.PutUint32(uint32_buf, cmd.CommandBlockData.Mode)
	_, err=writer.Write(uint32_buf)
	if err!=nil {
		return err
	}
	_, err=writer.Write(append([]byte(cmd.CommandBlockData.Command), 0))
	if err!=nil {
		return err
	}
	_, err=writer.Write(append([]byte(cmd.CommandBlockData.CustomName), 0))
	if err!=nil {
		return err
	}
	_, err=writer.Write(append([]byte(cmd.CommandBlockData.LastOutput), 0))
	if err!=nil {
		return err
	}
	binary.BigEndian.PutUint32(uint32_buf, uint32(cmd.CommandBlockData.TickDelay))
	_, err=writer.Write(uint32_buf)
	if err!=nil {
		return err
	}
	binary.BigEndian.PutUint32(uint32_buf, 0) // cleanup the buffer
	if cmd.CommandBlockData.ExecuteOnFirstTick {
		uint32_buf[0]=1
	}
	if cmd.CommandBlockData.TrackOutput {
		uint32_buf[1]=1
	}
	if cmd.CommandBlockData.Conditional {
		uint32_buf[2]=1
	}
	if cmd.CommandBlockData.NeedsRedstone {
		uint32_buf[3]=1
	}
	// ELSE statements are not required as the buffer was initiated w/ 0
	_, err=writer.Write(uint32_buf)
	return err
}

func (cmd *PlaceBlockWithCommandBlockData) Unmarshal(reader io.Reader) error {
	cmd.CommandBlockData=&types.CommandBlockData{}
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
	if err != nil {
		return err
	}
	cmd.BlockConstantStringID=binary.BigEndian.Uint16(buf[0:2])
	cmd.BlockData=binary.BigEndian.Uint16(buf[2:])
	_, err=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	cmd.CommandBlockData.Mode=binary.BigEndian.Uint32(buf)
	str, err:=readString(reader)
	if err!=nil {
		return err
	}
	cmd.CommandBlockData.Command=str
	str, err=readString(reader)
	if err!=nil {
		return err
	}
	cmd.CommandBlockData.CustomName=str
	str, err=readString(reader)
	if err!=nil {
		return err
	}
	cmd.CommandBlockData.LastOutput=str
	_, err=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	cmd.CommandBlockData.TickDelay=int32(binary.BigEndian.Uint32(buf))
	_, err=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	if buf[0]==0 {
		cmd.CommandBlockData.ExecuteOnFirstTick=false
	}else{
		cmd.CommandBlockData.ExecuteOnFirstTick=true
	}
	if buf[1]==0 {
		cmd.CommandBlockData.TrackOutput=false
	}else{
		cmd.CommandBlockData.TrackOutput=true
	}
	if buf[2]==0 {
		cmd.CommandBlockData.Conditional=false
	}else{
		cmd.CommandBlockData.Conditional=true
	}
	if buf[3]==0 {
		cmd.CommandBlockData.NeedsRedstone=false
	}else{
		cmd.CommandBlockData.NeedsRedstone=true
	}
	return nil
}