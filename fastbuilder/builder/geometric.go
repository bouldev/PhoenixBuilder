package builder

import (
	"phoenixbuilder/fastbuilder/types"
	"math"
)

func Circle(config *types.MainConfig, blc chan *types.Module)error {
	Radius := config.Radius
	Facing := config.Facing
	point := config.Position
	radius_squared:=math.Pow(float64(Radius), 2)
	var push_to_channel func(int, int)
	switch Facing {
	case "x":
		push_to_channel=func (x int, y int) {
			blc<-&types.Module {
				Point: types.Position {point.X, point.Y + y, point.Z + x},
			}
		}
	case "y":
		push_to_channel=func (x int, y int) {
			blc<-&types.Module {
				Point: types.Position {point.X+x, point.Y, point.Z + y},
			}
		}
	case "z":
		push_to_channel=func (x int, y int) {
			blc<-&types.Module {
				Point: types.Position {point.X+x, point.Y+y, point.Z},
			}
		}
	}
	for i:=0;i<=Radius;i++ {
		first_quadrant_val:=int(math.Sqrt(radius_squared-math.Pow(float64(i), 2)))
		push_to_channel(i, first_quadrant_val)
		if(first_quadrant_val!=0) {
			push_to_channel(i, -first_quadrant_val)
		}
		if(i!=0) {
			push_to_channel(-i,first_quadrant_val)
		}
		if(first_quadrant_val!=0&&i!=0) {
			push_to_channel(-i,-first_quadrant_val)
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
