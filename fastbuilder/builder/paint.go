package builder

import (
	"errors"
	"os"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"

	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
)

type ColorBlock struct {
	Color colorful.Color
	Block *types.ConstBlock
}

func Paint(config *types.MainConfig, blc chan *types.Module) error {
	width := config.Width
	height := config.Height
	facing := config.Facing
	pos := config.Position
	file, err := os.Open(config.Path)
	img, err := imaging.Decode(file)
	if err != nil {
		return I18n.ProcessSystemFileError(err)
	}
	if width != 0 && height != 0 {
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	}
	Max := img.Bounds().Max
	X, Y := Max.X, Max.Y
	//BlockSet := make([]*types.Module, X*Y)
	index := 0
	for x := 0; x < X; x++ {
		for y := 0; y < Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			c := colorful.Color{
				R: float64(r >> 8),
				G: float64(g >> 8),
				B: float64(b >> 8),
			}
			switch facing {
			default:
				return errors.New("Facing (-f) not defined")
			case "x":
				blc <- &types.Module{
					Point: types.Position{
						X: pos.X,
						Y: x + pos.Y,
						Z: y + pos.Z,
					},
					Block: getBlock(c),
				}
			case "y":
				blc <- &types.Module{
					Point: types.Position{
						X: x + pos.X,
						Y: pos.Y,
						Z: y + pos.Z,
					},
					Block: getBlock(c),
				}
			case "z":
				blc <- &types.Module{
					Point: types.Position{
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
	return nil
}

func getBlock(c colorful.Color) *types.Block {
	if _, _, _, a := c.RGBA(); a == 0 {
		return AirBlock.Take()
	}
	var List []float64
	for _, v := range ColorTable {
		s := c.DistanceRgb(v.Color)
		List = append(List, s)
	}
	return ColorTable[getMin(List)].Block.Take()
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
