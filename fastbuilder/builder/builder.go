package builder

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"errors"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
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
