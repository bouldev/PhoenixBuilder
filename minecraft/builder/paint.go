package builder

import (
	"errors"
	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
	"phoenixbuilder/minecraft/mctype"
)

type ColorBlock struct {
	Color colorful.Color
	Block mctype.Block
}

func Paint(config mctype.MainConfig) ([]mctype.Module, error) {
	path := config.Path
	width := config.Width
	height := config.Height
	facing := config.Facing
	pos := config.Position
	img, err := imaging.Open(path)
	if err != nil {
		return nil, err
	}
	if width != 0 && height != 0 {
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	}
	Max := img.Bounds().Max
	X, Y := Max.X, Max.Y
	BlockSet := make([]mctype.Module, X*Y)
	index := 0
	for x := 0; x < X; x++ {
		for y := 0; y < Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			c := colorful.Color{
				R: float64(r & 0xff),
				G: float64(g & 0xff),
				B: float64(b & 0xff),
			}
			switch facing {
			default:
				return nil, errors.New("Facing (-f) not defined")
			case "x":
				BlockSet[index] = mctype.Module{
					Point: mctype.Position{
						X: pos.X,
						Y: x + pos.Y,
						Z: y + pos.Z,
					},
					Block: getBlock(c),
				}
			case "y":
				BlockSet[index] = mctype.Module{
					Point: mctype.Position{
						X: x + pos.X,
						Y: pos.Y,
						Z: y + pos.Z,

					},
					Block: getBlock(c),
				}
			case "z":
				BlockSet[index] = mctype.Module{
					Point: mctype.Position{
						X: x + pos.X,
						Y: y + pos.Y,
						Z: pos.Z,
					},
					Block: getBlock(c),
				}
			}

			index++
		}
	}
	return BlockSet, nil
}

func getBlock(c colorful.Color) mctype.Block {
	if _, _, _, a := c.RGBA(); a == 0 {
		return mctype.Block{
			Name: "air",
			Data: 0,
		}
	}
	var List []float64
	for _, v := range ColorTable {
		s := c.DistanceRgb(v.Color)
		List = append(List, s)
	}
	return ColorTable[getMin(List)].Block
}

func getMin(t []float64) int {
	min := t[0]
	index := 0
	for i, v := range t {
		if v < min {
			min = v
			index = i
		}
	}
	return index
}

var ColorTable = []ColorBlock{
	{Block: mctype.Block{Name: "dirt", Data: 0}, Color: colorful.Color{134, 96, 67}},
	{Block: mctype.Block{Name: "cobblestone", Data: 0}, Color: colorful.Color{123, 123, 123}},
	{Block: mctype.Block{Name: "bedrock", Data: 0}, Color: colorful.Color{84, 84, 84}},
	{Block: mctype.Block{Name: "quartz_block", Data: 0}, Color: colorful.Color{237, 235, 228}},
	{Block: mctype.Block{Name: "emerald_block", Data: 0}, Color: colorful.Color{81, 217, 117}},
	{Block: mctype.Block{Name: "glowstone", Data: 0}, Color: colorful.Color{144, 118, 70}},
	{Block: mctype.Block{Name: "gold_block", Data: 0}, Color: colorful.Color{249, 236, 79}},
	{Block: mctype.Block{Name: "lapis_block", Data: 0}, Color: colorful.Color{39, 67, 138}},
	{Block: mctype.Block{Name: "log", Data: 14}, Color: colorful.Color{207, 206, 201}},
	{Block: mctype.Block{Name: "log", Data: 15}, Color: colorful.Color{87, 68, 27}},
	{Block: mctype.Block{Name: "melon_block", Data: 0}, Color: colorful.Color{151, 154, 37}},
	{Block: mctype.Block{Name: "netherrack", Data: 0}, Color: colorful.Color{111, 54, 53}},
	{Block: mctype.Block{Name: "purpur_block", Data: 0}, Color: colorful.Color{170, 123, 170}},
	{Block: mctype.Block{Name: "quartz_ore", Data: 0}, Color: colorful.Color{125, 85, 80}},
	{Block: mctype.Block{Name: "redstone_block", Data: 0}, Color: colorful.Color{171, 28, 9}},
	{Block: mctype.Block{Name: "sponge", Data: 0}, Color: colorful.Color{195, 196, 85}},
	{Block: mctype.Block{Name: "slime", Data: 0}, Color: colorful.Color{121, 200, 101}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 0}, Color: colorful.Color{210, 178, 162}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 10}, Color: colorful.Color{120, 72, 88}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 11}, Color: colorful.Color{75, 61, 93}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 12}, Color: colorful.Color{78, 52, 37}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 13}, Color: colorful.Color{104, 119, 54}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 15}, Color: colorful.Color{37, 23, 17}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 1}, Color: colorful.Color{163, 85, 39}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 2}, Color: colorful.Color{151, 90, 111}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 4}, Color: colorful.Color{188, 135, 37}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 3}, Color: colorful.Color{114, 110, 140}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 5}, Color: colorful.Color{104, 119, 54}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 6}, Color: colorful.Color{164, 80, 80}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 7}, Color: colorful.Color{58, 43, 37}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 8}, Color: colorful.Color{136, 108, 99}},
	{Block: mctype.Block{Name: "stone", Data: 1}, Color: colorful.Color{153, 114, 99}},
	{Block: mctype.Block{Name: "stone", Data: 2}, Color: colorful.Color{159, 115, 98}},
	{Block: mctype.Block{Name: "stone", Data: 5}, Color: colorful.Color{180, 180, 183}},
	{Block: mctype.Block{Name: "stone", Data: 0}, Color: colorful.Color{125, 125, 125}},
	{Block: mctype.Block{Name: "wool", Data: 0}, Color: colorful.Color{223, 223, 223}},
	{Block: mctype.Block{Name: "wool", Data: 10}, Color: colorful.Color{125, 60, 180}},
	{Block: mctype.Block{Name: "stained_hardened_clay", Data: 9}, Color: colorful.Color{87, 91, 91}},
	{Block: mctype.Block{Name: "wool", Data: 11}, Color: colorful.Color{44, 54, 134}},
	{Block: mctype.Block{Name: "wool", Data: 13}, Color: colorful.Color{54, 72, 28}},
	{Block: mctype.Block{Name: "wool", Data: 14}, Color: colorful.Color{151, 52, 49}},
	{Block: mctype.Block{Name: "wool", Data: 15}, Color: colorful.Color{37, 23, 17}},
	{Block: mctype.Block{Name: "wool", Data: 12}, Color: colorful.Color{82, 52, 32}},
	{Block: mctype.Block{Name: "wool", Data: 1}, Color: colorful.Color{219, 126, 64}},
	{Block: mctype.Block{Name: "wool", Data: 2}, Color: colorful.Color{179, 79, 188}},
	{Block: mctype.Block{Name: "wool", Data: 3}, Color: colorful.Color{107, 138, 201}},
	{Block: mctype.Block{Name: "wool", Data: 4}, Color: colorful.Color{189, 177, 43}},
	{Block: mctype.Block{Name: "wool", Data: 5}, Color: colorful.Color{69, 184, 59}},
	{Block: mctype.Block{Name: "wool", Data: 6}, Color: colorful.Color{211, 141, 160}},
	{Block: mctype.Block{Name: "wool", Data: 7}, Color: colorful.Color{65, 65, 65}},
	{Block: mctype.Block{Name: "wool", Data: 8}, Color: colorful.Color{158, 164, 164}},
	{Block: mctype.Block{Name: "wool", Data: 9}, Color: colorful.Color{47, 111, 138}},
	{Block: mctype.Block{Name: "concrete", Data: 0}, Color: colorful.Color{207, 213, 214}},
	{Block: mctype.Block{Name: "concrete", Data: 1}, Color: colorful.Color{224, 97, 0}},
	{Block: mctype.Block{Name: "concrete", Data: 2}, Color: colorful.Color{169, 48, 159}},
	{Block: mctype.Block{Name: "concrete", Data: 3}, Color: colorful.Color{35, 137, 198}},
	{Block: mctype.Block{Name: "concrete", Data: 4}, Color: colorful.Color{241, 175, 21}},
	{Block: mctype.Block{Name: "concrete", Data: 5}, Color: colorful.Color{94, 169, 25}},
	{Block: mctype.Block{Name: "concrete", Data: 6}, Color: colorful.Color{213, 101, 142}},
	{Block: mctype.Block{Name: "concrete", Data: 7}, Color: colorful.Color{55, 58, 62}},
	{Block: mctype.Block{Name: "concrete", Data: 8}, Color: colorful.Color{125, 125, 115}},
	{Block: mctype.Block{Name: "concrete", Data: 9}, Color: colorful.Color{21, 119, 136}},
	{Block: mctype.Block{Name: "concrete", Data: 10}, Color: colorful.Color{100, 32, 156}},
	{Block: mctype.Block{Name: "concrete", Data: 11}, Color: colorful.Color{45, 47, 143}},
	{Block: mctype.Block{Name: "concrete", Data: 12}, Color: colorful.Color{96, 60, 32}},
	{Block: mctype.Block{Name: "concrete", Data: 13}, Color: colorful.Color{73, 91, 36}},
	{Block: mctype.Block{Name: "concrete", Data: 14}, Color: colorful.Color{142, 33, 33}},
	{Block: mctype.Block{Name: "concrete", Data: 15}, Color: colorful.Color{8, 10, 15}},
	{Block: mctype.Block{Name: "coal_block", Data: 0}, Color: colorful.Color{19, 19, 19}},
	{Block: mctype.Block{Name: "diamond_block", Data: 0}, Color: colorful.Color{98, 219, 214}},
	{Block: mctype.Block{Name: "dried_kelp_block", Data: 0}, Color: colorful.Color{50, 59, 39}},
	{Block: mctype.Block{Name: "furnace", Data: 0}, Color: colorful.Color{96, 96, 96}},
	{Block: mctype.Block{Name: "hay_block", Data: 0}, Color: colorful.Color{169, 140, 16}},
	{Block: mctype.Block{Name: "iron_block", Data: 0}, Color: colorful.Color{219, 219, 219}},
	{Block: mctype.Block{Name: "stripped_birch_log", Data: 0}, Color: colorful.Color{185, 161, 104}},
	{Block: mctype.Block{Name: "stripped_acacia_log", Data: 0}, Color: colorful.Color{167, 92, 59}},
	{Block: mctype.Block{Name: "stripped_jungle_log", Data: 0}, Color: colorful.Color{171, 134, 85}},
	{Block: mctype.Block{Name: "stripped_oak_log", Data: 0}, Color: colorful.Color{164, 134, 81}},
	{Block: mctype.Block{Name: "stripped_spruce_log", Data: 0}, Color: colorful.Color{106, 83, 48}},
	{Block: mctype.Block{Name: "brick_block", Data: 0}, Color: colorful.Color{147, 100, 87}},
	{Block: mctype.Block{Name: "clay", Data: 0}, Color: colorful.Color{159, 164, 177}},
	{Block: mctype.Block{Name: "crafting_table", Data: 0}, Color: colorful.Color{107, 71, 43}},
	{Block: mctype.Block{Name: "end_stone", Data: 0}, Color: colorful.Color{221, 224, 165}},
	{Block: mctype.Block{Name: "red_glazed_terracotta", Data: 0}, Color: colorful.Color{182, 60, 53}},
	{Block: mctype.Block{Name: "noteblock", Data: 0}, Color: colorful.Color{101, 68, 51}},
	{Block: mctype.Block{Name: "sealantern", Data: 0}, Color: colorful.Color{172, 200, 190}},
	{Block: mctype.Block{Name: "soul_sand", Data: 0}, Color: colorful.Color{85, 64, 52}},
	{Block: mctype.Block{Name: "prismarine", Data: 0}, Color: colorful.Color{100, 152, 142}},
	{Block: mctype.Block{Name: "pink_glazed_terracotta", Data: 0}, Color: colorful.Color{235, 155, 182}},
	{Block: mctype.Block{Name: "purple_glazed_terracotta", Data: 0}, Color: colorful.Color{110, 48, 152}},
	{Block: mctype.Block{Name: "magenta_glazed_terracotta", Data: 0}, Color: colorful.Color{208, 100, 192}},
	{Block: mctype.Block{Name: "gray_glazed_terracotta", Data: 0}, Color: colorful.Color{83, 90, 94}},
	{Block: mctype.Block{Name: "yellow_glazed_terracotta", Data: 0}, Color: colorful.Color{234, 192, 89}},
	{Block: mctype.Block{Name: "blue_glazed_terracotta", Data: 0}, Color: colorful.Color{47, 65, 139}},
	{Block: mctype.Block{Name: "obsidian", Data: 0}, Color: colorful.Color{20, 18, 30}},
	{Block: mctype.Block{Name: "sponge", Data: 1}, Color: colorful.Color{160, 159, 63}},
	{Block: mctype.Block{Name: "bone_block", Data: 0}, Color: colorful.Color{206, 201, 178}},
}
