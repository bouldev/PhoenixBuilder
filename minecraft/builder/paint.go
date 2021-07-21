package builder

import "phoenixbuilder/minecraft/mctype"
import "github.com/disintegration/imaging"


func Paint(config mctype.MainConfig) ([]mctype.Module, error) {
	//point := config.Position
	path := config.Path
	width := config.Width
	height := config.Height
	img , err := imaging.Open(path)
	if err != nil {
		return []mctype.Module{}, err
	}
	if width != 0 && height != 0 {
		 imaging.Resize(img,width,height,imaging.Lanczos)
	}
	return nil, err
}
