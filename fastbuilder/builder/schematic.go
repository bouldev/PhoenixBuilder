package builder

import (
	"errors"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
)

func Schematic(config *types.MainConfig, blc chan *types.Module) error {
	return errors.New(I18n.T(I18n.Warning_Schem_Deprecated))
	//
	//file, err := os.Open(config.Path)
	//if err != nil {
	//	return I18n.ProcessSystemFileError(err)
	//}
	//defer file.Close()
	//gzip, err := gzip.NewReader(file)
	//if err != nil {
	//	return err
	//}
	//defer gzip.Close()
	//buffer, err := ioutil.ReadAll(gzip)
	//
	//var SchematicModule struct {
	//	Blocks    []byte `nbt:"Blocks"`
	//	Data      []byte `nbt:"Data"`
	//	Width     int    `nbt:"Width"`
	//	Length    int    `nbt:"Length"`
	//	Height    int    `nbt:"Height"`
	//	WEOffsetX int    `nbt:"WEOffsetX"`
	//	WEOffsetY int    `nbt:"WEOffsetY"`
	//	WEOffsetZ int    `nbt:"WEOffsetZ"`
	//}
	//
	//if err := nbt.Unmarshal(buffer, &SchematicModule); err != nil {
	//	// Won't return the error `err` since it contains a large content that can
	//	// crash the server after being sent.
	//	return fmt.Errorf(I18n.T(I18n.Sch_FailedToResolve))
	//}
	//if len(SchematicModule.Blocks) == 0 {
	//	return fmt.Errorf("Invalid structure.")
	//}
	//Size := [3]int{SchematicModule.Width, SchematicModule.Height, SchematicModule.Length}
	//Offset := [3]int{SchematicModule.WEOffsetX, SchematicModule.WEOffsetY, SchematicModule.WEOffsetZ}
	//X, Y, Z := 0, 1, 2
	//BlockIndex := 0
	//
	//for y := 0; y < Size[Y]; y++ {
	//	for z := 0; z < Size[Z]; z++ {
	//		for x := 0; x < Size[X]; x++ {
	//			p := config.Position
	//			p.X += x + Offset[X]
	//			p.Y += y + Offset[Y]
	//			p.Z += z + Offset[Z]
	//			var b types.Block
	//			b.Name = &BlockStr[SchematicModule.Blocks[BlockIndex]]
	//			b.Data = uint16(SchematicModule.Data[BlockIndex])
	//			if BlockIndex-188 <= 5 && BlockIndex-188 >= 0 {
	//				b.Name = &FenceName
	//				b.Data = uint16(BlockIndex - 188)
	//			}
	//			if BlockIndex == 3 && b.Data == 2 {
	//				b.Name = &PodzolName
	//			}
	//			if *b.Name != "air" {
	//				blc <- &types.Module{Point: p, Block: &b}
	//			}
	//			BlockIndex++
	//		}
	//	}
	//}
	//return nil
}
