package builder

import (
	"errors"
	"phoenixbuilder/minecraft/mctype"
)

var Builder = map[string]func(config mctype.MainConfig) ([]mctype.Module, error){
	"round":     Round,
	"circle":    Circle,
	"sphere":    Sphere,
	"ellipse":   Ellipse,
	"ellipsoid": Ellipsoid,
	"plot":      Paint,
	"schem":     Schematic,
	"acme":      Acme,
}

func Generate(config mctype.MainConfig) ([]mctype.Module, error) {
	if config.Execute == "" {
		return []mctype.Module{}, errors.New("Command Not Found.")
	} else {
		return Builder[config.Execute](config)
	}
}

func PipeGenerate(configs []mctype.Config) []mctype.Module {
	return []mctype.Module{}
}
