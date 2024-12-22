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
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/string_reader"
)

func is_block_states(str string) bool {
	reader := string_reader.NewStringReader(&str)
	reader.JumpSpace()
	return reader.Next(true) == "["
}

func format_block_states(blockStates string) (string, error) {
	blockStatesMap, err := mcstructure.UnmarshalBlockStates(blockStates)
	if err != nil {
		return "", fmt.Errorf("format_block_states: %v", err)
	}
	blockStatesString, err := mcstructure.MarshalBlockStates(blockStatesMap)
	if err != nil {
		return "", fmt.Errorf("format_block_states: %v", err)
	}
	return blockStatesString, nil
}
