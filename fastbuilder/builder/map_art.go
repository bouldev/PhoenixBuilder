package builder

import (
	_ "embed"
	"fmt"
	"image"
	"os"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"

	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	Height2D = iota
	Height3D_Light
	Height3D_Dark
	Heigth3D_Normal = Height2D
)

type colorBlock struct {
	Name   string
	Meta   int
	Height uint8
}

var colorArray2D []*colorful.Color
var blockArray2D []*colorBlock
var colorArray3D []*colorful.Color
var blockArray3D []*colorBlock

func init() {
	colorArray2D = make([]*colorful.Color, 0)
	blockArray2D = make([]*colorBlock, 0)
	colorArray3D = make([]*colorful.Color, 0)
	blockArray3D = make([]*colorBlock, 0)

	type CData struct {
		Name  string    `json:"block"`
		Meta  int       `json:"meta"`
		Color []float64 `json:"color"`
	}
	ligther := 255.0 / 220.0
	darker := 180.0 / 220.0
	for _, cdata := range ColorTable {
		name := cdata.Block.Name
		meta := int(cdata.Block.Data)
		r, g, b := cdata.Color.R/255.0, cdata.Color.G/255.0, cdata.Color.B/255.0
		//name := cdata[0][0].(string)
		//meta := int(cdata[0][1].(float64))
		//r, g, b := cdata[1][0].(float64)/255.0, cdata[1][1].(float64)/255.0, cdata[1][2].(float64)/255.0
		blockArray2D = append(blockArray2D, &colorBlock{Name: name, Meta: meta, Height: Height2D})
		colorArray2D = append(colorArray2D, &colorful.Color{R: r, G: g, B: b})
		blockArray3D = append(blockArray3D, &colorBlock{Name: name, Meta: meta, Height: Heigth3D_Normal})
		colorArray3D = append(colorArray3D, &colorful.Color{R: r, G: g, B: b})
		blockArray3D = append(blockArray3D, &colorBlock{Name: name, Meta: meta, Height: Height3D_Light})
		colorArray3D = append(colorArray3D, &colorful.Color{R: r * ligther, G: g * ligther, B: b * ligther})
		blockArray3D = append(blockArray3D, &colorBlock{Name: name, Meta: meta, Height: Height3D_Dark})
		colorArray3D = append(colorArray3D, &colorful.Color{R: r * darker, G: g * darker, B: b * darker})
	}
}

func clip(c int64) int64 {
	if c < 0 {
		return 0
	}
	if c > 255 {
		return 255
	}
	return c
}

func Closest(tc [3]float64, colors *[]*colorful.Color) int {
	r, g, b := int64(tc[0]), int64(tc[1]), int64(tc[2])
	r = clip(r)
	g = clip(g)
	b = clip(b)
	delta := int64(2 << 30)
	bestCi := 0
	for ci, c := range *colors {
		pR_, pG_, pB_, _ := c.RGBA()
		pR, pG, pB := int64(pR_>>8), int64(pG_>>8), int64(pB_>>8)
		var d int64
		if r+pR > 256 {
			d = int64(2*(r-pR)*(r-pR) + 4*(g-pG)*(g-pG) + 3*(b-pB)*(b-pB))
		} else {
			d = int64(3*(r-pR)*(r-pR) + 4*(g-pG)*(g-pG) + 2*(b-pB)*(b-pB))
		}
		if d < delta {
			delta = d
			bestCi = ci
		}
	}
	return bestCi
}

func Dither(img image.Image, colors *[]*colorful.Color, blocks *[]*colorBlock) (image.Image, [][]*colorBlock) {
	origBounds := img.Bounds()
	origSize := origBounds.Max
	previewImg := imaging.Clone(img)
	W, H := origSize.X, origSize.Y
	blockImg := make([][]*colorBlock, H)
	// oh no, i need a float matrix to avoid overflow!
	imgMatrix := make([][][3]float64, H)
	for i, _ := range blockImg {
		blockImg[i] = make([]*colorBlock, W)
		imgMatrix[i] = make([][3]float64, W)
	}
	for r := 0; r < H; r++ {
		for c := 0; c < W; c++ {
			origC := previewImg.At(c, r)
			tR_, tG_, tB_, _ := origC.RGBA()
			// 0~255 rgb in flot 64, yes!
			tR, tG, tB := float64(tR_>>8), float64(tG_>>8), float64(tB_>>8)
			imgMatrix[r][c] = [3]float64{tR, tG, tB}
		}
	}
	for r := 0; r < H; r++ {
		for c := 0; c < W; c++ {
			origC := imgMatrix[r][c]
			ci := Closest(origC, colors)
			realC := (*colors)[ci]
			previewImg.Set(c, r, realC)
			blockImg[r][c] = (*blocks)[ci]
			tR, tG, tB := origC[0], origC[1], origC[2]
			rR_, rG_, rB_, _ := realC.RGBA()
			rR, rG, rB := float64(rR_>>8), float64(rG_>>8), float64(rB_>>8)
			dR, dG, dB := tR-rR, tG-rG, tB-rB
			if c != W-1 {
				nearbyC := imgMatrix[r][c+1]
				nR, nG, nB := nearbyC[0], nearbyC[1], nearbyC[2]
				imgMatrix[r][c+1] = [3]float64{
					(nR) + (7.0/16.0)*(dR),
					(nG) + (7.0/16.0)*(dG),
					(nB) + (7.0/16.0)*(dB),
				}
			}
			if r != H-1 {
				nearbyC := imgMatrix[r+1][c]
				nR, nG, nB := nearbyC[0], nearbyC[1], nearbyC[2]
				imgMatrix[r+1][c] = [3]float64{
					(nR) + (5.0/16.0)*(dR),
					(nG) + (5.0/16.0)*(dG),
					(nB) + (5.0/16.0)*(dB),
				}
				if c != 0 {
					nearbyC := imgMatrix[r+1][c-1]
					nR, nG, nB := nearbyC[0], nearbyC[1], nearbyC[2]
					imgMatrix[r+1][c-1] = [3]float64{
						(nR) + (1.0/16.0)*(dR),
						(nG) + (1.0/16.0)*(dG),
						(nB) + (1.0/16.0)*(dB),
					}
				}
				if c != W-1 {
					nearbyC := imgMatrix[r+1][c+1]
					nR, nG, nB := nearbyC[0], nearbyC[1], nearbyC[2]
					imgMatrix[r+1][c+1] = [3]float64{
						(nR) + (3.0/16.0)*(dR),
						(nG) + (3.0/16.0)*(dG),
						(nB) + (3.0/16.0)*(dB),
					}
				}
			}
		}
	}

	return previewImg, blockImg
}

func GetYMap(blocks [][]*colorBlock, MapY int) [][]int {
	H := len(blocks)
	W := len(blocks[0])
	YMap := make([][]int, H)
	for i := 0; i < H; i++ {
		YMap[i] = make([]int, W)
	}

	for x := 0; x < W; x++ {
		colYmap := make([]int, 1)
		colYmap[0] = 0
		min := 0
		for z := 0; z < H; z++ {
			height := blocks[z][x].Height
			lastY := colYmap[len(colYmap)-1]
			switch height {
			case Height3D_Dark:
				colYmap = append(colYmap, lastY+2)
				break
			case Height3D_Light:
				colYmap = append(colYmap, lastY-2)
				if lastY-2 < min {
					min = lastY - 2
				}
				break
			case Heigth3D_Normal:
				colYmap = append(colYmap, lastY)
				break
			}
		}
		for z := 0; z < H; z++ {
			YMap[z][x] = (colYmap[z] - min) % MapY
		}
	}

	return YMap
}

func MapArt(config *types.MainConfig, blc chan *types.Module) error {
	//path := config.Path
	MapX := config.MapX
	MapZ := config.MapZ
	MapY := config.MapY
	if MapY != 0 {
		if MapY < 20 || MapY > 255 {
			return fmt.Errorf(I18n.T(I18n.Error_MapY_Exceed), MapY)
		}

	}
	pos := config.Position
	file, err := os.Open(config.Path)
	img, err := imaging.Decode(file)
	if err != nil {
		return I18n.ProcessSystemFileError(err)
	}
	origBounds := img.Bounds()
	origSize := origBounds.Max
	origRaito := float64(origSize.X) / float64(origSize.Y)
	targetRaito := float64(MapX) / float64(MapZ)
	if origRaito > targetRaito {
		lX := int(float64(origSize.Y) * targetRaito)
		sX := (origSize.X - lX) / 2
		eX := sX + lX
		if sX < 0 {
			sX = 0
		}
		if eX > origSize.X {
			eX = origSize.X
		}
		img = imaging.Crop(img, image.Rect(sX, 0, eX, origSize.Y))
	} else if origRaito < targetRaito {
		lY := int(float64(origSize.X) / targetRaito)
		sY := (origSize.Y - lY) / 2
		eY := sY + lY
		if sY < 0 {
			sY = 0
		}
		if eY > origSize.Y {
			eY = origSize.Y
		}
		img = imaging.Crop(img, image.Rect(0, sY, origSize.X, eY))
	}
	img = imaging.Resize(img, MapX*128, MapZ*128, imaging.Lanczos)

	if MapY == 0 {
		_, blockImg := Dither(img, &colorArray2D, &blockArray2D)
		for Z, row := range blockImg {
			for X, blk := range row {
				blc <- &types.Module{
					Point: types.Position{
						X: X + pos.X,
						Y: pos.Y,
						Z: Z + pos.Z,
					},
					Block: &types.Block{
						Name: &blk.Name,
						Data: uint16(blk.Meta),
					},
				}
			}
		}
	} else {
		_, blockImg := Dither(img, &colorArray3D, &blockArray3D)
		YMap := GetYMap(blockImg, MapY)
		for Z, row := range blockImg {
			for X, blk := range row {
				blc <- &types.Module{
					Point: types.Position{
						X: X + pos.X,
						Y: pos.Y + YMap[Z][X],
						Z: Z + pos.Z,
					},
					Block: &types.Block{
						Name: &blk.Name,
						Data: uint16(blk.Meta),
					},
				}
			}
		}
	}

	return nil
}
