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

func ReplaceItemInContainerRequest(module *types.Module, method string) string {
	command := fmt.Sprintf(
		"replaceitem block %d %d %d slot.container %d %s %d %d",
		module.Point.X,
		module.Point.Y,
		module.Point.Z,
		module.ChestSlot.Slot,
		module.ChestSlot.Name,
		module.ChestSlot.Count,
		module.ChestSlot.Damage,
	)
	if len(method) == 0 {
		return command
	} else {
		return fmt.Sprintf("%v %v", command, method)
	}
}

func ReplaceItemInInventoryRequest(
	itemInfo *types.ChestSlot,
	target string,
	generateLocation string,
	method string,
) string {
	command := fmt.Sprintf(
		"replaceitem entity %s %s %s %d %d",
		target,
		generateLocation,
		itemInfo.Name,
		itemInfo.Count,
		itemInfo.Damage,
	)
	if len(method) == 0 {
		return command
	} else {
		return fmt.Sprintf("%v %v", command, method)
	}
}
