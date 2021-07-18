package builder

import "gophertunnel/minecraft/mctype"

func Ellipse(config mctype.MainConfig) ([]mctype.Module, error) {
	Length := config.Length
	Width := config.Width
	Facing := config.Facing
	point := config.Position
	var BlockSet []mctype.Module
	switch Facing {
	case "x":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b mctype.Module
					b.Point = mctype.Position{point.X, point.Y + i, point.Z + j}
					b.Block = config.Block
					BlockSet = append(BlockSet, b)
				}
			}
		}
	case "y":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b mctype.Module
					b.Point = mctype.Position{point.X + i, point.Y, point.Z + j}
					b.Block = config.Block
					BlockSet = append(BlockSet, b)
				}
			}
		}
	case "z":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b mctype.Module
					b.Point = mctype.Position{point.X + i, point.Y + j, point.Z}
					b.Block = config.Block
					BlockSet = append(BlockSet, b)
				}
			}
		}
	}
	return BlockSet, nil
}
