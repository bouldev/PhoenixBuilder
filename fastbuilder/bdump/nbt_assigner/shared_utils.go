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

import (
	"encoding/json"
	"fmt"
	"math"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/mirror/blocks"
	"strings"
)

// 从 SupportBlocksPool 检查这个方块实体是否已被支持。
// 如果尚未被支持，则返回空字符串，否则返回这种方块的类型。
// 以告示牌为例，所有的告示牌都可以写作为 Sign
func IsNBTBlockSupported(blockName string) string {
	value, ok := SupportBlocksPool[blockName]
	if ok {
		return value
	}
	return ""
}

// 从 SupportItemsPool 检查这个 NBT 物品是否已被支持。
// 如果尚未被支持，则返回空字符串，否则返回这种物品的类型。
// 以物品展示框为例，物品展示框和荧光物品展示框都可以写作为 Frame
func IsNBTItemSupported(itemName string) string {
	value, ok := SupportItemsPool[itemName]
	if ok {
		return value
	}
	return ""
}

// 从 ContainerCouldOpenPool 检查这个容器是否可以打开。
// 如果不能打开，则返回假，否则返回真。
// 如果提供的 blockName 不是容器，则始终返回假
func ContainerCouldOpen(blockName string) bool {
	return SupportContainerPool[blockName].CouldOpen
}

// 将 itemComponents 编码为游戏支持的 JSON 格式。
// 如果传入的 itemComponents 为空指针，则返回空字符串
func MarshalItemComponents(itemComponents *ItemComponents) string {
	type can_place_on_or_can_destroy struct {
		Blocks []string `json:"blocks"`
	}
	type item_lock struct {
		Mode string `json:"mode"`
	}
	res := map[string]interface{}{}
	// 初始化
	if itemComponents == nil {
		return ""
	}
	// 如果物品组件不存在，那么应该返回空字符串而非 {}
	if len(itemComponents.CanPlaceOn) > 0 {
		res["can_place_on"] = can_place_on_or_can_destroy{Blocks: itemComponents.CanPlaceOn}
	}
	if len(itemComponents.CanDestroy) > 0 {
		res["can_destroy"] = can_place_on_or_can_destroy{Blocks: itemComponents.CanDestroy}
	}
	if itemComponents.KeepOnDeath {
		res["keep_on_death"] = struct{}{}
	}
	if len(itemComponents.ItemLock) != 0 {
		res["item_lock"] = item_lock{Mode: itemComponents.ItemLock}
	}
	// 赋值
	bytes, _ := json.Marshal(res)
	return string(bytes)
	// 返回值
}

// 取得名称为 blockName 且数据值(附加值)为 metaData 的方块的方块状态。
// 特别地，name **不**需要加上命名空间 minecraft
func get_block_states_from_legacy_block(
	blockName string,
	metaData uint16,
) (map[string]interface{}, error) {
	// What the fuck?
	if (blockName == "chest" || blockName == "minecraft:chest") && metaData == 0 {
		return map[string]any{"facing_direction": int32(2)}, nil
	}
	// ...
	standardRuntimeID, found := blocks.LegacyBlockToRuntimeID(blockName, metaData)
	if !found {
		return nil, fmt.Errorf("get_block_states_from_legacy_block: Failed to get the runtimeID of block %s; metaData = %d", blockName, metaData)
	}
	generalBlock, found := blocks.RuntimeIDToBlock(standardRuntimeID)
	if !found {
		return nil, fmt.Errorf("get_block_states_from_legacy_block: Failed to converse StandardRuntimeID to NEMCRuntimeID; standardRuntimeID = %d, blockName = %s, metaData = %d", standardRuntimeID, blockName, metaData)
	}
	return generalBlock.States().ToNBT(), nil
}

// 取得名称为 blockName 且方块状态为 blockStates 的旧方块的
// 新名称及它的新方块状态。
// 特别地，name 需要加上命名空间 minecraft
func get_new_block_states_from_older(
	blockName string,
	blockStates map[string]any,
) (string, map[string]any, error) {
	runtimeID, found := blocks.BlockNameAndStateToRuntimeID(blockName, blockStates)
	if !found {
		return "", nil, fmt.Errorf("get_new_block_states_from_older: Failed to get the runtimeID of block %s; blockStates = %#v", blockName, blockStates)
	}
	blockName, states, found := blocks.RuntimeIDToState(runtimeID)
	if !found {
		return "", nil, fmt.Errorf("get_new_block_states_from_older: Failed to converse runtimeID to block with states; runtimeID = %d, blockName = %s, blockStates = %#v", runtimeID, blockName, blockStates)
	}
	return blockName, states, nil
}

// 将 types.Module 解析为 GeneralBlock
func ParseBlockModule(singleBlock *types.Module) (GeneralBlock, error) {
	blockStates, err := mcstructure.UnmarshalBlockStates(singleBlock.Block.BlockStates)
	if err != nil {
		return GeneralBlock{}, fmt.Errorf("ParseBlockModule: Failed to parse the block states from string; singleBlock.Block.BlockStates = %#v, err = %v", singleBlock.Block.BlockStates, err)
	}
	// get block states
	return GeneralBlock{
		Name:   strings.Replace(strings.ToLower(strings.ReplaceAll(*singleBlock.Block.Name, " ", "")), "minecraft:", "", 1),
		States: blockStates,
		NBT:    singleBlock.NBTMap,
	}, nil
	// return
}

/*
将 singleItem 解析为 GeneralItem 。

特别地，如果此物品存在 item_lock 物品组件，
则只会解析该物品的物品组件和附魔属性，
因为存在 item_lock 的物品无法使用铁砧修改名称，
也不能跨容器移动；

如果此物品是一个 NBT 方块，
则附魔属性和物品组件数据将被丢弃；

如果该物品是一个 NBT 物品，
则物品组件数据将被丢弃
*/
func (i *ItemPackage) ParseItemFromNBT(singleItem ItemOrigin) error {
	err := i.Item.DecodeItemBasicData(singleItem)
	if err != nil {
		return fmt.Errorf("ParseItemFromNBT: %v", err)
	}
	i.AdditionalData.Type = IsNBTItemSupported(i.Item.Basic.Name)
	// basic
	err = i.Item.DecodeItemEnhancementData(singleItem)
	if err != nil {
		return fmt.Errorf("ParseItemFromNBT: %v", err)
	}
	// enhancement
	if i.Item.Enhancement != nil && i.Item.Enhancement.ItemComponents != nil && len(i.Item.Enhancement.ItemComponents.ItemLock) != 0 {
		return nil
	}
	// 如果此物品使用了物品组件 item_lock ，
	// 则后续数据将不被解析。
	// 因为存在 item_lock 的物品无法使用铁砧修改名称，
	// 也不能跨容器移动
	err = i.DecodeItemCustomData(singleItem)
	if err != nil {
		return fmt.Errorf("ParseItemFromNBT: %v", err)
	}
	// custom
	if i.Item.Custom != nil && i.Item.Custom.SubBlockData != nil && i.Item.Enhancement != nil {
		i.Item.Enhancement.Enchantments = nil
		i.Item.Enhancement.ItemComponents = nil
	}
	// 如果此物品是一个 NBT 方块，
	// 则附魔属性和物品组件将被丢弃
	if i.Item.Custom != nil && i.Item.Custom.NBTItemData != nil && i.Item.Enhancement != nil {
		i.Item.Enhancement.ItemComponents = nil
	}
	// 如果该物品是一个 NBT 物品，
	// 则物品组件数据将被丢弃
	return nil
	// return
}

// 计算两个 RGB 颜色 colorA 和 colorB 的加权欧式距离
func CalculateColorDistance(colorA [3]uint8, colorB [3]uint8) float64 {
	rmean := float64(colorA[0]+colorB[0]) / 2
	deltaR := float64(colorA[0] - colorB[0])
	deltaG := float64(colorA[1] - colorB[1])
	deltaB := float64(colorA[2] - colorB[2])
	return math.Sqrt((2+rmean/256)*deltaR*deltaR + 4*deltaG*deltaG + (2+(255-rmean)/256)*deltaB*deltaB)
}

// 从 mapping 中选出距离 color 最近的 RGB 颜色
func SearchForBestColor(color [3]uint8, mapping [][3]uint8) (result [3]uint8) {
	distance := math.Inf(1)
	for _, value := range mapping {
		if deltaC := CalculateColorDistance(color, value); deltaC < distance {
			distance = deltaC
			result = value
		}
	}
	return
}

// 从 x 解码一个 RGBA 颜色
func DecodeVarRGBA(x int32) (RGB [3]uint8, RGBA [4]uint8) {
	R, G, B := uint8(x>>16), uint8(x>>8), uint8(x)
	A := uint8(x >> 24)
	return [3]uint8{R, G, B}, [4]uint8{R, G, B, A}
}
