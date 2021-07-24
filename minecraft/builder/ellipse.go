package builder

import "phoenixbuilder/minecraft/mctype"

func Ellipse(config *mctype.MainConfig, blc chan *mctype.Module) error {
	Length := config.Length
	Width := config.Width
	Facing := config.Facing
	point := config.Position
	switch Facing {
	case "x":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b mctype.Module
					b.Point = mctype.Position{point.X, point.Y + i, point.Z + j}
					blc <- &b
				}
			}
		}
	case "y":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b mctype.Module
					b.Point = mctype.Position{point.X + i, point.Y, point.Z + j}
					blc <- &b
				}
			}
		}
	case "z":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b mctype.Module
					b.Point = mctype.Position{point.X + i, point.Y + j, point.Z}
					blc <- &b
				}
			}
		}
	}
	return nil
}
