package universe_import

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"os"
	"path"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/omega/utils/structure"
	"strconv"

	"github.com/Tnze/go-mc/nbt"

	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/pterm/pterm"
)

const (
	Height2D = iota
	Height3D_Light
	Height3D_Dark
	Heigth3D_Normal = Height2D
)

type colorBlock struct {
	RuntimeID uint32
	Height    uint8
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
		Meta  uint16    `json:"meta"`
		Color []float64 `json:"color"`
	}
	// ligther := 255.0 / 220.0
	// darker := 180.0 / 220.0
	for _, cdata := range chunk.ColorTable {
		r, g, b := cdata.Color.R/255.0, cdata.Color.G/255.0, cdata.Color.B/255.0
		//name := cdata[0][0].(string)
		//meta := int(cdata[0][1].(float64))
		//r, g, b := cdata[1][0].(float64)/255.0, cdata[1][1].(float64)/255.0, cdata[1][2].(float64)/255.0
		rtid, found := chunk.LegacyBlockToRuntimeID(cdata.Block.Name, cdata.Block.Val)
		if !found || rtid == chunk.AirRID {
			panic(fmt.Errorf("missing color block mapping %v", cdata))
		}
		javaBlock, found := chunk.RuntimeIDToJava(rtid)
		if !found {
			javaBlock = fmt.Sprintf("omega:as_legacy_block[name=%v,val=%v]", cdata.Block.Name, cdata.Block.Val)
		}
		rtidJ, found := chunk.JavaToRuntimeID(javaBlock)
		if !found || rtidJ != rtid {
			panic(fmt.Errorf("%v %v %v %v", cdata.Block.Name, cdata.Block.Val, found, rtidJ))
		}
		if blk, found := chunk.RuntimeIDToLegacyBlock(rtid); !found || blk.Name != cdata.Block.Name || blk.Val != cdata.Block.Val {
			panic(fmt.Errorf("missing color block mapping %v", cdata))
		}
		blockArray2D = append(blockArray2D, &colorBlock{RuntimeID: rtid, Height: Height2D})
		colorArray2D = append(colorArray2D, &colorful.Color{R: r, G: g, B: b})
		// blockArray3D = append(blockArray3D, &colorBlock{Name: name, Meta: meta, Height: Heigth3D_Normal})
		// colorArray3D = append(colorArray3D, &colorful.Color{R: r, G: g, B: b})
		// blockArray3D = append(blockArray3D, &colorBlock{Name: name, Meta: meta, Height: Height3D_Light})
		// colorArray3D = append(colorArray3D, &colorful.Color{R: r * ligther, G: g * ligther, B: b * ligther})
		// blockArray3D = append(blockArray3D, &colorBlock{Name: name, Meta: meta, Height: Height3D_Dark})
		// colorArray3D = append(colorArray3D, &colorful.Color{R: r * darker, G: g * darker, B: b * darker})
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

func PreProcessImage(img image.Image, dir string, cmds []string) (structureFile string, err error) {
	origBounds := img.Bounds()
	origSize := origBounds.Max
	origRaito := float64(origSize.X) / float64(origSize.Y)

	if len(cmds) < 2 {
		return "", fmt.Errorf("这是一张图片,导入时应该为 [路径] [x] [y] [z] [x方向地图数] [z方向地图数]")
	}
	MapX := 0
	MapZ := 0
	if val, err := strconv.Atoi(cmds[0]); err != nil || val < 1 {
		return "", fmt.Errorf("参数 [x方向地图数] %v 不是正整数，应该为一个正整数", cmds[0])
	} else {
		MapX = val
	}
	if val, err := strconv.Atoi(cmds[1]); err != nil || val < 1 {
		return "", fmt.Errorf("参数 [x方向地图数] %v 不是正整数，应该为一个正整数", cmds[1])
	} else {
		MapZ = val
	}

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
	previewImg, blockImg := Dither(img, &colorArray2D, &blockArray2D)
	previewImagePath := path.Join(dir, "preview.png")
	err = imaging.Save(previewImg, previewImagePath)
	if err != nil {
		return "", err
	} else {
		pterm.Success.Println("生成的效果预览图位于 " + previewImagePath)
	}
	schemFile := &structure.SchemFileStructrue{
		Palette: make(map[string]int32),
		Metadata: structure.WEOffset{
			WEOffsetX: 0,
			WEOffsetY: 0,
			WEOffsetZ: 0,
		},
		DataVersion:   2975,
		Offset:        []int32{0, 0, 0},
		PaletteMax:    0,
		Version:       2,
		Length:        int16(len(blockImg)),
		Height:        1,
		Width:         int16(len(blockImg[0])),
		BlockEntities: structure.NbtBlocks{},
	}
	writerBlocks := bytes.NewBuffer(make([]byte, 0, uint32(float64(len(blockImg)*len(blockImg[0]))*1.3)))
	writeVarUint32 := func(u uint32) {
		for u&128 != 0 {
			_ = writerBlocks.WriteByte(byte(u) | 128)
			u >>= 7
		}
		_ = writerBlocks.WriteByte(byte(u))
	}
	convertor := structure.NewRuntimeIDToPaletteConvertor()
	convertor.AcquirePaletteFN = func(u uint32) string {
		javaStr, found := chunk.RuntimeIDToJava(u)
		if found {
			return javaStr
		} else {
			if block, found := chunk.RuntimeIDToLegacyBlock(u); found {
				return fmt.Sprintf("omega:as_legacy_block[name=%v,val=%v]", block.Name, block.Val)
			} else {
				return chunk.JavaAirBlock
			}

		}

	}
	convertor.Convert(chunk.AirRID) // force air to be 0
	// woodRTID, _ := chunk.LegacyBlockToRuntimeID("wood", 0)
	// for _, row := range blockImg {
	// 	for _, _ = range row {
	// 		writeVarUint32(convertor.Convert(woodRTID))
	// 	}
	// }
	for _, row := range blockImg {
		for _, blk := range row {
			rtid := convertor.Convert(blk.RuntimeID)
			if rtid == chunk.AirRID {
				panic(blk)
			}
			writeVarUint32(rtid)
		}
	}
	schemFile.BlockDataIn = writerBlocks.Bytes()
	for paletteI, paletteStr := range convertor.Palette {
		schemFile.Palette[paletteStr] = int32(paletteI)
	}
	schemFile.PaletteMax = int32(len(schemFile.Palette))
	schemImageDir := path.Join(dir, "image.schem")
	fp, err := os.OpenFile(schemImageDir, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	writer := gzip.NewWriter(fp)
	defer writer.Close()
	// nbtEncoder := standard_nbt.NewEncoderWithEncoding(writer, standard_nbt.BigEndian)
	// err = nbtEncoder.EncodeWithRootTag(*schemFile, "Schematic")
	err = nbt.NewEncoder(writer).Encode(*schemFile, "Schematic")
	if err != nil {
		return "", err
	}
	return schemImageDir, nil
}
