package parse

import (
	"flag"
	"gophertunnel/minecraft/builder"
	"gophertunnel/minecraft/mctype"
	"strconv"
	"strings"
)

func Parse(Message string, defaultConfig mctype.MainConfig) mctype.MainConfig {
	SLC := strings.Split(Message," ")
	Config := mctype.MainConfig{
		Execute:   "",
		Block:     mctype.Block{},
		OldBlock:  mctype.Block{},
		Begin:     mctype.Position{},
		End:       mctype.Position{},
		Position:  defaultConfig.Position,
		Radius:    0,
		Length:    0,
		Width:     0,
		Height:    0,
		Method:    "replace",
		OldMethod: "keep",
	}

	FlagSet := flag.NewFlagSet("Parser", 0)
	//Length,  Width and Height
	FlagSet.IntVar(&Config.Length,"length",defaultConfig.Length,"The length")
	FlagSet.IntVar(&Config.Length,"l",defaultConfig.Length,"The length")
	FlagSet.IntVar(&Config.Width,"width",defaultConfig.Width,"The width")
	FlagSet.IntVar(&Config.Width,"w",defaultConfig.Width,"The width")
	FlagSet.IntVar(&Config.Height,"height",defaultConfig.Height,"The height")
	FlagSet.IntVar(&Config.Height,"h",defaultConfig.Height,"The height")
	//Radius
	FlagSet.IntVar(&Config.Radius,"radius",defaultConfig.Radius,"The radius")
	FlagSet.IntVar(&Config.Radius,"r",defaultConfig.Radius,"The radius")
	//Facing, Path, Shape
	FlagSet.StringVar(&Config.Facing,"facing",defaultConfig.Facing,"Building's facing")
	FlagSet.StringVar(&Config.Facing,"f",defaultConfig.Facing,"Building's facing")
	FlagSet.StringVar(&Config.Path,"path",defaultConfig.Path,"The path of file")
	FlagSet.StringVar(&Config.Path,"p",defaultConfig.Path,"The path of file")
	FlagSet.StringVar(&Config.Shape,"shape",defaultConfig.Shape,"The path of file")
	FlagSet.StringVar(&Config.Shape,"s",defaultConfig.Shape,"The path of file")
	//Block
	FlagSet.StringVar(&Config.Block.Name,"block",defaultConfig.Block.Name,"Blocks that make up the building")
	FlagSet.StringVar(&Config.Block.Name,"b",defaultConfig.Block.Name,"Blocks that make up the building")
	FlagSet.IntVar(&Config.Block.Data,"data",defaultConfig.Block.Data,"The data of Block")
	FlagSet.IntVar(&Config.Block.Data,"d",defaultConfig.Block.Data,"The data of Block")
	//OldBlock
	FlagSet.StringVar(&Config.OldBlock.Name,"old_block",defaultConfig.OldBlock.Name,"Blocks that make up the building")
	FlagSet.StringVar(&Config.OldBlock.Name,"ob",defaultConfig.OldBlock.Name,"Blocks that make up the building")
	FlagSet.IntVar(&Config.OldBlock.Data,"old_data",defaultConfig.OldBlock.Data,"The data of Block")
	FlagSet.IntVar(&Config.OldBlock.Data,"od",defaultConfig.OldBlock.Data,"The data of Block")

	FlagSet.Parse(SLC[1:])
	for k, _ := range builder.Builder {
		if k == SLC[0] {
			Config.Execute = k
		}
	}
	for index, v := range SLC {
		if v == "-p" || v == "--position" {
			x, xe := strconv.Atoi(SLC[index + 1])
			y, ye := strconv.Atoi(SLC[index + 2])
			z, ze := strconv.Atoi(SLC[index + 3])
			if xe == nil && ye == nil && ze == nil {
				Config.Position = mctype.Position{X: x, Y: y, Z: z}
			}
		}
	}
	return Config
}

func PipeParse(Message string, config mctype.MainConfig) []mctype.MainConfig{
	ChatSlice := strings.Split(Message,"|")
	var Configs []mctype.MainConfig
	for _, v := range ChatSlice {
		Configs = append(Configs,Parse(v,config))
	}
	return Configs
}