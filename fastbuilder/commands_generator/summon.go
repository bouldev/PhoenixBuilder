package commands_generator

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
)


func SummonRequest(module *types.Module, config *types.MainConfig) string {
	entity := config.Entity
	point := module.Point
	return fmt.Sprintf("summon %s %d %d %d", entity, point.X, point.Y, point.Z)
}

