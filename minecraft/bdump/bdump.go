package bdump

import (
	"github.com/andybalholm/brotli"
	"phoenixbuilder/minecraft/mctype"
	"fmt"
	"os"
	"encoding/binary"
)

type BDump struct {
	Author string
	Blocks []*mctype.Module
}

func (bdump *BDump) formatBlocks() {
	min:=[]int{2147483647,2147483647,2147483647}
	for _, mdl := range bdump.Blocks {
		if mdl.Point.X<min[0] {
			min[0]=mdl.Point.X
		}
		if mdl.Point.Y<min[1] {
			min[1]=mdl.Point.Y
		}
		if mdl.Point.Z<min[2] {
			min[2]=mdl.Point.Z
		}
	}
	for _, mdl := range bdump.Blocks {
		mdl.Point.X-=min[0]
		mdl.Point.Y-=min[1]
		mdl.Point.Z-=min[2]
	}
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

end 88
*/

func (bdump *BDump) writeHeader(w *brotli.Writer) error {
	_, err:=w.Write([]byte("BDX"))
	if err!=nil {
		return err
	}
	_, err=w.Write([]byte{0})
	if err!=nil {
		return err
	}
	_, err=w.Write([]byte(bdump.Author))
	if err!=nil {
		return err
	}
	_, err=w.Write([]byte{0})
	return err
}

func (bdump *BDump) writeBlocks(w *brotli.Writer) error {
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
					_, err:=w.Write([]byte{3})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else{
					wrap:=mdl.Point.X-brushPosition[0]
					if wrap > 65535 {
						_, err:=w.Write([]byte{10})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,4)
						binary.BigEndian.PutUint32(writeO,uint32(wrap))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else{
						_, err:=w.Write([]byte{2})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,2)
						binary.BigEndian.PutUint16(writeO,uint16(wrap))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}
				}
				brushPosition[0]=mdl.Point.X
				brushPosition[1]=0
				brushPosition[2]=0
				continue
			}else if(mdl.Point.Y!=brushPosition[1]) {
				if(mdl.Point.Y-brushPosition[1]==1){
					_, err:=w.Write([]byte{5})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else{
					wrap:=mdl.Point.Y-brushPosition[1]
					if wrap > 65535 {
						_, err:=w.Write([]byte{11})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,4)
						binary.BigEndian.PutUint32(writeO,uint32(wrap))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else{
						_, err:=w.Write([]byte{4})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,2)
						binary.BigEndian.PutUint16(writeO,uint16(wrap))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}
				}
				brushPosition[1]=mdl.Point.Y
				brushPosition[2]=0
				continue
			}else if(mdl.Point.Z!=brushPosition[2]) {
				if(mdl.Point.Z-brushPosition[2]==1){
					_, err:=w.Write([]byte{8})
					if err != nil {
						return fmt.Errorf("Failed to write command")
					}
				}else{
					wrap:=mdl.Point.Z-brushPosition[2]
					if wrap > 65535 {
						_, err:=w.Write([]byte{12})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,4)
						binary.BigEndian.PutUint32(writeO,uint32(wrap))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}else{
						_, err:=w.Write([]byte{6})
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
						writeO:=make([]byte,2)
						binary.BigEndian.PutUint16(writeO,uint16(wrap))
						_, err=w.Write(writeO)
						if err != nil {
							return fmt.Errorf("Failed to write command")
						}
					}
				}
				brushPosition[2]=mdl.Point.Z
			}
			break
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
	w.Write([]byte("XE"))
	return nil
}

func (bdump *BDump) WriteToFile(path string) error {
	file, err:=os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE,0666)
	if err!=nil {
		return fmt.Errorf("Failed to open file: %v", err)
	}
	defer file.Close()
	_, err=file.Write([]byte("BD@"))
	if err!=nil {
		return fmt.Errorf("Failed to write BRBDP file header")
	}
	brw := brotli.NewWriter(file)
	defer brw.Close()
	err=bdump.writeHeader(brw)
	if err!=nil {
		return err
	}
	err=bdump.writeBlocks(brw)
	return err
}