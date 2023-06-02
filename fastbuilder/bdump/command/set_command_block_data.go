package command

import (
	"io"
	"encoding/binary"
	"phoenixbuilder/fastbuilder/types"
)

type SetCommandBlockData struct {
	CommandBlockData *types.CommandBlockData
}

func (_ *SetCommandBlockData) ID() uint16 {
	return 26
}

func (_ *SetCommandBlockData) Name() string {
	return "SetCommandBlockDataCommand"
}

func (cmd *SetCommandBlockData) Marshal(writer io.Writer) error {
	uint32_buf:=make([]byte, 4)
	binary.BigEndian.PutUint32(uint32_buf, cmd.CommandBlockData.Mode)
	_, err:=writer.Write(uint32_buf)
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

func (cmd *SetCommandBlockData) Unmarshal(reader io.Reader) error {
	cmd.CommandBlockData = &types.CommandBlockData{}
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
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