package commands_generator

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
	"fmt"
	"phoenixbuilder/fastbuilder/types"
)

func SetBlockRequest(module *types.Module, config *types.MainConfig) string {
	Block := module.Block
	Point := module.Point
	Method := config.Method
	if Block != nil {
		if len(Block.BlockStates) != 0 {
			return fmt.Sprintf("setblock %d %d %d %s %s %s", Point.X, Point.Y, Point.Z, *Block.Name, Block.BlockStates, Method)
		} else {
			return fmt.Sprintf("setblock %d %d %d %s %d %s", Point.X, Point.Y, Point.Z, *Block.Name, Block.Data, Method)
		}
	} else {
		return fmt.Sprintf("setblock %d %d %d %s %d %s", Point.X, Point.Y, Point.Z, config.Block.Name, config.Block.Data, Method)
	}

}
