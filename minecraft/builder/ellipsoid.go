package builder

import "phoenixbuilder/minecraft/mctype"

func Ellipsoid(config *mctype.MainConfig, blc chan *mctype.Module) error {
	Length := config.Length
	Width := config.Width
	Height := config.Height
	point := config.Position
	for i := -Length; i <= Length; i++ {
		for j := -Width; j <= Width; j++ {
			for k := -Height; k <= Height; k++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width)+(k*k)/(Height*Height) <= 1 {
					var b mctype.Module
					b.Point = mctype.Position{point.X + i, point.Y + j, point.Z + k}
					blc <- &b
				}
			}
		}
	}
	return nil
}
