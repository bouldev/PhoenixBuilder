package builder

import "phoenixbuilder/minecraft/mctype"

func Circle(config *mctype.MainConfig, blc chan *mctype.Module)error {
	Radius := config.Radius
	Facing := config.Facing
	point := config.Position
	switch Facing {
	case "x":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius && i*i+j*j >= (Radius-1)*(Radius-1) {
					var b mctype.Module
					b.Point = mctype.Position{point.X, point.Y + i, point.Z + j}
					blc <- &b
				}
			}
		}
	case "y":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius && i*i+j*j >= (Radius-1)*(Radius-1) {
					var b mctype.Module
					b.Point = mctype.Position{point.X + i, point.Y, point.Z + j}
					blc <- &b
				}
			}
		}
	case "z":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius && i*i+j*j >= (Radius-1)*(Radius-1) {
					var b mctype.Module
					b.Point = mctype.Position{point.X + i, point.Y + j, point.Z}
					blc <- &b
				}
			}
		}
	}
	return nil
}
