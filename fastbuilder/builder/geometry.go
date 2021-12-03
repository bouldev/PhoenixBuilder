package builder

import "phoenixbuilder/fastbuilder/types"

func Circle(config *types.MainConfig, blc chan *types.Module)error {
	Radius := config.Radius
	Facing := config.Facing
	point := config.Position
	switch Facing {
	case "x":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius && i*i+j*j >= (Radius-1)*(Radius-1) {
					var b types.Module
					b.Point = types.Position{point.X, point.Y + i, point.Z + j}
					blc <- &b
				}
			}
		}
	case "y":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius && i*i+j*j >= (Radius-1)*(Radius-1) {
					var b types.Module
					b.Point = types.Position{point.X + i, point.Y, point.Z + j}
					blc <- &b
				}
			}
		}
	case "z":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius && i*i+j*j >= (Radius-1)*(Radius-1) {
					var b types.Module
					b.Point = types.Position{point.X + i, point.Y + j, point.Z}
					blc <- &b
				}
			}
		}
	}
	return nil
}


func Ellipse(config *types.MainConfig, blc chan *types.Module) error {
	Length := config.Length
	Width := config.Width
	Facing := config.Facing
	point := config.Position
	switch Facing {
	case "x":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b types.Module
					b.Point = types.Position{point.X, point.Y + i, point.Z + j}
					blc <- &b
				}
			}
		}
	case "y":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b types.Module
					b.Point = types.Position{point.X + i, point.Y, point.Z + j}
					blc <- &b
				}
			}
		}
	case "z":
		for i := -Length; i <= Length; i++ {
			for j := -Width; j <= Width; j++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width) < 1 {
					var b types.Module
					b.Point = types.Position{point.X + i, point.Y + j, point.Z}
					blc <- &b
				}
			}
		}
	}
	return nil
}

func Ellipsoid(config *types.MainConfig, blc chan *types.Module) error {
	Length := config.Length
	Width := config.Width
	Height := config.Height
	point := config.Position
	for i := -Length; i <= Length; i++ {
		for j := -Width; j <= Width; j++ {
			for k := -Height; k <= Height; k++ {
				if (i*i)/(Length*Length)+(j*j)/(Width*Width)+(k*k)/(Height*Height) <= 1 {
					var b types.Module
					b.Point = types.Position{point.X + i, point.Y + j, point.Z + k}
					blc <- &b
				}
			}
		}
	}
	return nil
}

func Round(config *types.MainConfig, blc chan *types.Module) error {
	Radius := config.Radius
	Facing := config.Facing
	point := config.Position
	switch Facing {
	case "x":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius {
					var b types.Module
					b.Point = types.Position{X: point.X, Y: point.Y + i, Z: point.Z + j}
					blc <- &b
				}
			}
		}
	case "y":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius {
					var b types.Module
					b.Point = types.Position{X: point.X + i, Y: point.Y, Z: point.Z + j}
					blc <- &b
				}
			}
		}
	case "z":
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius {
					var b types.Module
					b.Point = types.Position{point.X + i, point.Y + j, point.Z}
					blc <- &b
				}
			}
		}
	}
	return nil
}

func Sphere(config *types.MainConfig, blc chan *types.Module) error {
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
						var b types.Module
						b.Point = types.Position{point.X + i, point.Y + j, point.Z + k}
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
						var b types.Module
						b.Point = types.Position{point.X + i, point.Y + j, point.Z + k}
						blc <- &b
					}
				}
			}
		}
	}
	return nil
}
