package parsing

import (
	"flag"
	"fmt"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
	"strings"
)

func Parse(Message string, defaultConfig *types.MainConfig) (*types.MainConfig, error) {
	//SLC := strings.Split(Message," ")
	isEscaping := false
	isInQuote := false
	var SLC []string
	curmsg := ""
	for _, c := range Message {
		if isEscaping {
			isEscaping = false
			curmsg += string(c)
			continue
		}
		if c == '\\' {
			isEscaping = true
			continue
		}
		if c == '"' {
			isInQuote = (!isInQuote)
			continue
		}
		if c == ' ' && !isInQuote {
			SLC = append(SLC, curmsg)
			curmsg = ""
			continue
		}
		if c == '#' {
			break
		}
		curmsg += string(c)
	}
	if len(curmsg) > 0 {
		SLC = append(SLC, curmsg)
		curmsg = ""
	}
	//fmt.Printf("%v\n",SLC)
	if isInQuote {
		return nil, fmt.Errorf(I18n.T(I18n.Parsing_UnterminatedQuotedString))
	} else if isEscaping {
		return nil, fmt.Errorf(I18n.T(I18n.Parsing_UnterminatedEscape))
	}
	Config := &types.MainConfig{
		Execute:  "",
		Block:    &types.ConstBlock{},
		OldBlock: &types.ConstBlock{},
		//Begin:     types.Position{},
		End:                defaultConfig.End,
		Position:           defaultConfig.Position,
		Radius:             0,
		Length:             0,
		Width:              0,
		Height:             0,
		Method:             "replace",
		OldMethod:          "keep",
		AssignNBTData:      false,
		ExcludeCommands:    false,
		InvalidateCommands: false,
	}

	FlagSet := flag.NewFlagSet("Parser", 0)
	var tempBlockData int
	var tempOldBlockData int
	//Length,  Width and Height
	FlagSet.BoolVar(&Config.AssignNBTData, "assignnbtdata", defaultConfig.AssignNBTData, "Assign NBT data to blocks by lawful means while importing")
	FlagSet.BoolVar(&Config.AssignNBTData, "nbt", defaultConfig.AssignNBTData, "Assign NBT data to blocks by lawful means while importing")
	FlagSet.BoolVar(&Config.ExcludeCommands, "excludecommands", defaultConfig.ExcludeCommands, "Exclude commands in command blocks")
	FlagSet.BoolVar(&Config.InvalidateCommands, "invalidatecommands", defaultConfig.InvalidateCommands, "Invalidate commands in command blocks")
	FlagSet.BoolVar(&Config.Strict, "strict", defaultConfig.Strict, "Break if the file isn't signed")
	FlagSet.BoolVar(&Config.Strict, "S", defaultConfig.Strict, "Break if the file isn't signed")

	FlagSet.IntVar(&Config.Length, "length", defaultConfig.Length, "The length")
	FlagSet.IntVar(&Config.Length, "l", defaultConfig.Length, "The length")
	FlagSet.IntVar(&Config.Width, "width", defaultConfig.Width, "The width")
	FlagSet.IntVar(&Config.Width, "w", defaultConfig.Width, "The width")
	FlagSet.IntVar(&Config.Height, "height", defaultConfig.Height, "The height")
	FlagSet.IntVar(&Config.Height, "h", defaultConfig.Height, "The height")
	//Radius
	FlagSet.IntVar(&Config.Radius, "radius", defaultConfig.Radius, "The radius")
	FlagSet.IntVar(&Config.Radius, "r", defaultConfig.Radius, "The radius")
	// Map Art Configuration
	FlagSet.IntVar(&Config.MapX, "mapX", defaultConfig.MapX, "Take X maps in map art")
	FlagSet.IntVar(&Config.MapZ, "mapZ", defaultConfig.MapZ, "Take Z maps in map art")
	FlagSet.IntVar(&Config.MapY, "mapY", defaultConfig.MapY, "Available Height (blocks) for 3D map art")
	//Facing, Path, Shape
	FlagSet.StringVar(&Config.Facing, "facing", defaultConfig.Facing, "Building's facing")
	FlagSet.StringVar(&Config.Facing, "f", defaultConfig.Facing, "Building's facing")
	FlagSet.StringVar(&Config.Path, "path", defaultConfig.Path, "The path of file")
	FlagSet.StringVar(&Config.Path, "p", defaultConfig.Path, "The path of file")
	FlagSet.StringVar(&Config.Shape, "shape", defaultConfig.Shape, "The shape of geometric structure")
	FlagSet.StringVar(&Config.Shape, "s", defaultConfig.Shape, "The shape of geometric structure")
	//Block
	FlagSet.StringVar(&Config.Block.Name, "block", defaultConfig.Block.Name, "Blocks making up the structure")
	FlagSet.StringVar(&Config.Block.Name, "b", defaultConfig.Block.Name, "Blocks making up the structure")
	FlagSet.StringVar(&Config.Entity, "entity", "", "")
	FlagSet.StringVar(&Config.Entity, "e", "", "")
	FlagSet.IntVar(&tempBlockData, "data", int(defaultConfig.Block.Data), "The data of Block")
	FlagSet.IntVar(&tempBlockData, "d", int(defaultConfig.Block.Data), "The data of Block")
	//OldBlock
	FlagSet.StringVar(&Config.OldBlock.Name, "old_block", defaultConfig.OldBlock.Name, "Blocks that make up the building")
	FlagSet.StringVar(&Config.OldBlock.Name, "ob", defaultConfig.OldBlock.Name, "Blocks that make up the building")
	FlagSet.IntVar(&tempOldBlockData, "old_data", int(defaultConfig.OldBlock.Data), "The data of Block")
	FlagSet.IntVar(&tempOldBlockData, "od", int(defaultConfig.OldBlock.Data), "The data of Block")
	// Resume
	FlagSet.Float64Var(&Config.ResumeFrom, "resume", float64(defaultConfig.ResumeFrom), "Resume Construction from percentage, async only")

	FlagSet.Parse(SLC[1:])
	/*for k, _ := range builder.Builder {
		if k == SLC[0] {
			Config.Execute = k
		}
	}*/
	Config.Execute = SLC[0]
	// Since the function system exists ^^

	//for index, v := range SLC {
	//	if v == "-p" || v == "--position" {
	//		x, xe := strconv.Atoi(SLC[index + 1])
	//		y, ye := strconv.Atoi(SLC[index + 2])
	//		z, ze := strconv.Atoi(SLC[index + 3])
	//		if xe == nil && ye == nil && ze == nil {
	//			Config.Position = types.Position{X: x, Y: y, Z: z}
	//		}
	//	}
	//}
	Config.Block.Data = uint16(tempBlockData)
	Config.OldBlock.Data = uint16(tempOldBlockData)
	return Config, nil
}

func PipeParse(Message string, config *types.MainConfig) ([]*types.MainConfig, error) {
	ChatSlice := strings.Split(Message, "|")
	var Configs []*types.MainConfig
	for _, v := range ChatSlice {
		pv, err := Parse(v, config)
		if err != nil {
			return nil, err
		}
		Configs = append(Configs, pv)
	}
	return Configs, nil
}
