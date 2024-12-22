package NBTAssigner

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

import "fmt"

// DefaultBlock 结构体用于描述一个完整的方块实体数据。
// 任何未被支持的方块实体都会被重定向为此结构体
type DefaultBlock struct {
	BlockEntity *BlockEntity // 该方块实体的详细数据
}

// 这只是为了保证接口一致而设
func (d *DefaultBlock) Decode() error {
	return nil
}

// 放置目标方块但不写入方块实体数据
func (d *DefaultBlock) WriteData() error {
	err := d.BlockEntity.Interface.SetBlockAsync(d.BlockEntity.AdditionalData.Position, d.BlockEntity.Block.Name, d.BlockEntity.AdditionalData.BlockStates)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	return nil
}
