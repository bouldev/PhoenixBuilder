package builder

import "phoenixbuilder/minecraft/mctype"

func Sphere(config *mctype.MainConfig, blc chan *mctype.Module) error {
	Radius := config.Radius
	Shape := config.Shape
	point := config.Position
	switch Shape {
	default:
	case "hollow":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				for k := -Radius; k <= Radius; k++ {
					if i*i+j*j+k*k <= Radius*Radius && i*i+j*j+k*k >= (Radius-1)*(Radius-1) {
						var b mctype.Module
						b.Point = mctype.Position{point.X + i, point.Y + j, point.Z + k}
						blc <- &b
					}
				}
			}
		}
	case "solid":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				for k := -Radius; k <= Radius; k++ {
					if i*i+j*j+k*k <= Radius*Radius {
						var b mctype.Module
						b.Point = mctype.Position{point.X + i, point.Y + j, point.Z + k}
						blc <- &b
					}
				}
			}
		}
	}
	return nil
}
