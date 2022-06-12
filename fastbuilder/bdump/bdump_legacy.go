package bdump

import (
	"github.com/andybalholm/brotli"
	"phoenixbuilder/fastbuilder/types"
	"bytes"
	"fmt"
	"os"
	"encoding/binary"
)

type BDumpLegacy struct {
	Author string // Should be empty
	Blocks []*types.Module
}

/*
if(i.cmd=="addToBlockPalette"){
	writebuf(1,1);
	writebuf(i.blockName+"\0");
}else if(i.cmd=="addX"){
	writebuf(2,1);
	writebuf(i.count,2);
}else if(i.cmd=="X++"){
	writebuf(3,1);
}else if(i.cmd=="addY"){
	writebuf(4,1);
	writebuf(i.count,2);
}else if(i.cmd=="Y++"){
	writebuf(5,1);
}else if(i.cmd=="addZ"){
	writebuf(6,1);
	writebuf(i.count,2);
}else if(i.cmd=="placeBlock"){
	writebuf(7,1);
	writebuf(i.blockID,2);
	writebuf(i.blockData,2);
}else if(i.cmd=="Z++"){
	writebuf(8,1);
}else{
	writebuf(9,1);//NOP
}
jumpX 10
jumpY 11
jumpZ 12
reserved 13

*X++  14
*X--  15
*Y++  16
*Y--  17
*Z++  18
*Z--  19
*addX 20
*addBigX 21
*addY 22
*addBigY 23
*addZ 24
*addBigZ 25
assignCommandBlockData 26
placeCommandBlockWithData 27
addSmallX 28
addSmallY 29
addSmallZ 30

end 88
isSigned    90
*/

func (bdump *BDumpLegacy) formatBlocks() {
	
}

func (bdump *BDumpLegacy) writeHeader(w *bytes.Buffer) error {
	_, err:=w.Write([]byte("BDX"))
	if err!=nil {
		return err
	}
	_, err=w.Write([]byte{0})
	if err!=nil {
		return err
	}
	// No author info
	_, err=w.Write([]byte{0})
	return err
}

func (bdump *BDumpLegacy) writeBlocks(w *bytes.Buffer) error {
	bdump.formatBlocks()
	brushPosition:=[]int{0,0,0}
	blocksPalette:=make(map[string]int)
	cursor := 0
	for _, mdl := range bdump.Blocks {
		blknm:=*mdl.Block.Name
		_, found := blocksPalette[blknm]
		if found {
			continue
		}
		_, err:=w.Write([]byte{1}) //addToPalette
		if (err != nil) {
			return fmt.Errorf("Failed to write palette")
		}
		_, err=w.Write([]byte(blknm))
		if (err != nil) {
			return fmt.Errorf("Failed to write palette p2")
		}
		_, err=w.Write([]byte{0})
		if (err != nil) {
			return fmt.Errorf("Failed to write palette p3")
		}
		blocksPalette[blknm]=cursor;
		cursor++
	}
	for _,mdl := range bdump.Blocks {
		for {
			if(mdl.Point.X!=brushPosition[0]) {
				if(mdl.Point.X-brushPosition[0]==1){
					_, err:=w.Write([]byte{14})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else if(mdl.Point.X-brushPosition[0]==-1){
					_, err:=w.Write([]byte{15})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else{
					wrap:=mdl.Point.X-brushPosition[0]
					if (wrap < -32768||wrap > 32767) {
						_, err:=w.Write([]byte{21})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,4)
						binary.BigEndian.PutUint32(writeO,uint32(int32(wrap)))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else if(wrap < -127||wrap > 127){
						_, err:=w.Write([]byte{20})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,2)
						binary.BigEndian.PutUint16(writeO,uint16(int16(wrap)))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else{
						_, err:=w.Write([]byte{28})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						_, err=w.Write([]byte{uint8(int8(wrap))})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}
				}
				brushPosition[0]=mdl.Point.X
				continue
			}else if(mdl.Point.Y!=brushPosition[1]) {
				if(mdl.Point.Y-brushPosition[1]==1){
					_, err:=w.Write([]byte{16})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else if(mdl.Point.Y-brushPosition[1]==-1){
					_, err:=w.Write([]byte{17})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else{
					wrap:=mdl.Point.Y-brushPosition[1]
					if (wrap > 32767||wrap < -32768) {
						_, err:=w.Write([]byte{23})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,4)
						binary.BigEndian.PutUint32(writeO,uint32(int32(wrap)))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else if(wrap > 127||wrap < -127){
						_, err:=w.Write([]byte{22})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,2)
						binary.BigEndian.PutUint16(writeO,uint16(int16(wrap)))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else{
						_, err:=w.Write([]byte{29})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						_, err=w.Write([]byte{uint8(int8(wrap))})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}
				}
				brushPosition[1]=mdl.Point.Y
				continue
			}else if(mdl.Point.Z!=brushPosition[2]) {
				if(mdl.Point.Z-brushPosition[2]==1){
					_, err:=w.Write([]byte{18})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else if(mdl.Point.Z-brushPosition[2]==1){
					_, err:=w.Write([]byte{19})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else{
					wrap:=mdl.Point.Z-brushPosition[2]
					if (wrap > 32767||wrap < -32768) {
						_, err:=w.Write([]byte{25})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,4)
						binary.BigEndian.PutUint32(writeO,uint32(int32(wrap)))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else if(wrap > 127||wrap < -127){
						_, err:=w.Write([]byte{24})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,2)
						binary.BigEndian.PutUint16(writeO,uint16(int16(wrap)))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else{
						_, err:=w.Write([]byte{30})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						_, err=w.Write([]byte{uint8(int8(wrap))})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}
				}
				brushPosition[2]=mdl.Point.Z
			}
			break
		}
		if mdl.CommandBlockData != nil {
			_, err:=w.Write([]byte{27})
			writeA:=make([]byte,2)
			wac, _ := blocksPalette[*mdl.Block.Name]
			binary.BigEndian.PutUint16(writeA,uint16(wac))
			_, err1 := w.Write(writeA)
			writeB:=make([]byte,2)
			binary.BigEndian.PutUint16(writeB,uint16(mdl.Block.Data))
			_, err2 := w.Write(writeB)
			if(err!=nil||err1!=nil||err2!=nil){
				return fmt.Errorf("Failed to write line230")
			}
			dt:=mdl.CommandBlockData
			wMode:=make([]byte,4)
			binary.BigEndian.PutUint32(wMode,dt.Mode)
			_, err1=w.Write(wMode)
			_, err2=w.Write([]byte(dt.Command))
			_, err3:=w.Write([]byte{0})
			_, err4:=w.Write([]byte(dt.CustomName))
			_, err5:=w.Write([]byte{0})
			_, err6:=w.Write([]byte(dt.LastOutput))
			_, err7:=w.Write([]byte{0})
			wTickDelay:=make([]byte,4)
			binary.BigEndian.PutUint32(wTickDelay,uint32(int32(dt.TickDelay)))
			_, err8:=w.Write(wTickDelay)
			fBools:=make([]byte,4)
			if dt.ExecuteOnFirstTick {
				fBools[0]=1
			}else{
				fBools[0]=0
			}
			if dt.TrackOutput {
				fBools[1]=1
			}else{
				fBools[1]=0
			}
			if dt.Conditional {
				fBools[2]=1
			}else{
				fBools[2]=0
			}
			if dt.NeedRedstone {
				fBools[3]=1
			}else{
				fBools[3]=0
			}
			_, err9:=w.Write(fBools)
			if(err!=nil||err1!=nil||err2!=nil||err3!=nil||err4!=nil||err5!=nil||err6!=nil||err7!=nil||err8!=nil||err9!=nil){
				return fmt.Errorf("Failed to write cbcmd")
			}
			continue
		}
		_, err:=w.Write([]byte{7})
		writeA:=make([]byte,2)
		wac, _ := blocksPalette[*mdl.Block.Name]
		binary.BigEndian.PutUint16(writeA,uint16(wac))
		_, err1 := w.Write(writeA)
		writeB:=make([]byte,2)
		binary.BigEndian.PutUint16(writeB,uint16(mdl.Block.Data))
		_, err2 := w.Write(writeB)
		if(err!=nil||err1!=nil||err2!=nil){
			return fmt.Errorf("Failed to write line230")
		}
	}
	return nil
}

func (bdump *BDumpLegacy) WriteToFile(path string, localCert string, localKey string) (error, error) {
	file, err:=os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE,0666)
	if err!=nil {
		return fmt.Errorf("Failed to open file: %v", err), nil
	}
	defer file.Close()
	_, err=file.Write([]byte("BD@"))
	if err!=nil {
		return fmt.Errorf("Failed to write BRBDP file header"), nil
	}
	buffer:=&bytes.Buffer{}
	brw := brotli.NewWriter(file)
	err=bdump.writeHeader(buffer)
	if err!=nil {
		return err, nil
	}
	err=bdump.writeBlocks(buffer)
	if err!=nil {
		return err, nil
	}
	bts:=buffer.Bytes()
	_, err=brw.Write(bts)
	if(err!=nil) {
		return err, nil
	}
	sign, signerr:=SignBDX(bts, localKey, localCert)
	if(signerr!=nil) {
		brw.Write([]byte("XE"))
	}else{
		brw.Write(append([]byte{88}, sign...))
		if(len(sign)>=255) {
			realLength:=make([]byte,2)
			binary.BigEndian.PutUint16(realLength,uint16(len(sign)+2))
			brw.Write(realLength)
			brw.Write([]byte{uint8(255)})
		}else{
			brw.Write([]byte{uint8(len(sign))})
		}
		brw.Write([]byte{90})
	}
	err=brw.Close()
	return err, signerr
}
