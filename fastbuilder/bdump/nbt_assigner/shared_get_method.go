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

// 取得用于放置目标方块实体的 接口/方法
func GetPlaceBlockMethod(block *BlockEntity) GeneralBlockNBT {
	switch block.AdditionalData.Type {
	case "CommandBlock":
		return &CommandBlock{BlockEntity: block, ShouldPlaceBlock: true}
	case "Container":
		return &Container{BlockEntity: block}
	case "Sign":
		return &Sign{BlockEntity: block}
	case "Frame":
		return &Frame{BlockEntity: block}
	case "StructureBlock":
		return &StructureBlock{BlockEntity: block}
	case "Banner":
		return &BannerBlock{BlockEntity: block}
	default:
		return &DefaultBlock{BlockEntity: block}
		// 其他尚且未被支持的方块实体
	}
}

// 取得用于生成目标 NBT 物品的 接口/方法
func GetGenerateItemMethod(item *ItemPackage) GeneralItemNBT {
	switch item.AdditionalData.Type {
	case "Book":
		return &Book{ItemPackage: item}
	case "Banner":
		return &BannerItem{ItemPackage: item}
	case "Shield":
		return &Shield{ItemPackage: item}
	default:
		return &DefaultItem{ItemPackage: item}
		// 其他尚且未被支持的 NBT 物品
	}
}
