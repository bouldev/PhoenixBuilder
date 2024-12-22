package GameInterface

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
	"phoenixbuilder/mirror/blocks"
	"strings"

	"github.com/google/uuid"
)

// 返回 uniqueID 在字符串化之后的安全形式，
// 因为我们得考虑 NEMC 的屏蔽词机制
func uuid_to_safe_string(uniqueID uuid.UUID) string {
	str := uniqueID.String()
	for key, value := range StringUUIDReplaceMap {
		str = strings.ReplaceAll(str, key, value)
	}
	return str
}

// 取得名称为 name 且方块状态为 states 的方块的 Block Runtime ID 。
// 特别地，name 需要加上命名空间 minecraft
func blockStatesToRuntimeID(
	name string,
	states map[string]interface{},
) (uint32, error) {
	runtimeID, found := blocks.BlockNameAndStateToRuntimeID(name, states)
	if !found {
		return 0, fmt.Errorf("blockStatesToRuntimeID: Failed to get the runtimeID of block %v; states = %#v", name, states)
	}
	return runtimeID, nil
}
