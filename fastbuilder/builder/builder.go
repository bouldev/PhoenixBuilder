package builder

import (
	"errors"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/i18n"
)

var Builder = map[string]func(config *types.MainConfig, blc chan *types.Module) error{
	"round":     Round,
	"circle":    Circle,
	"sphere":    Sphere,
	"ellipse":   Ellipse,
	"ellipsoid": Ellipsoid,
	"paint":     Paint,
	"schematic": Schematic,
	"acme":      Acme,
	"bdump":     BDump,
	"mapart":    MapArt,
}

func Generate(config *types.MainConfig, blc chan *types.Module) error {
	if config.Execute == "" {
		return errors.New(I18n.T(I18n.CommandNotFound))
	} else {
		return Builder[config.Execute](config, blc)
	}
}
