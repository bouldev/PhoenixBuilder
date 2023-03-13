package blockNBT

import (
	"fmt"
	blockNBT_API "phoenixbuilder/fastbuilder/bdump/blockNBT/API"
	blockNBT_CommandBlock "phoenixbuilder/fastbuilder/bdump/blockNBT/CommandBlock"
	blockNBT_Container "phoenixbuilder/fastbuilder/bdump/blockNBT/Container"
	blockNBT_global "phoenixbuilder/fastbuilder/bdump/blockNBT/Global"
	blockNBT_Sign "phoenixbuilder/fastbuilder/bdump/blockNBT/Sign"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/types"
	"strings"
	"sync"
)

// 将 types.Module 转换为 blockNBT_depends.GeneralBlock
func parseBlockModule(singleBlock *types.Module) (blockNBT_global.GeneralBlock, error) {
	// init var
	got, err := mcstructure.ParseStringNBT(singleBlock.Block.BlockStates, true)
	if err != nil {
		return blockNBT_global.GeneralBlock{}, fmt.Errorf("parseBlockModule: Could not parse block states; singleBlock.Block.BlockStates = %#v", singleBlock.Block.BlockStates)
	}
	blockStates, normal := got.(map[string]interface{})
	if !normal {
		return blockNBT_global.GeneralBlock{}, fmt.Errorf("parseBlockModule: The target block states is not map[string]interface{}; got = %#v", got)
	}
	// get block states
	return blockNBT_global.GeneralBlock{
		Name:   strings.Replace(strings.ToLower(strings.ReplaceAll(*singleBlock.Block.Name, " ", "")), "minecraft:", "", 1),
		States: blockStates,
		NBT:    singleBlock.NBTMap,
	}, nil
	// return
}

// 检查这个方块实体是否已被支持
func checkIfIsEffectiveNBTBlock(blockName string) string {
	value, ok := blockNBT_global.SupportBlocksPool[blockName]
	if ok {
		return value
	}
	return ""
}

/*
带有 NBT 数据放置方块

如果你也想参与更多方块实体的支持，可以去看看这个库 https://github.com/df-mc/dragonfly

这个库也依然基于 gophertunnel
*/
func placeBlockWithNBTData(pack *blockNBT_global.BlockEntityDatas) error {
	switch pack.Datas.Type {
	case "CommandBlock":
		newStruct := blockNBT_CommandBlock.CommandBlock{BlockEntityDatas: pack}
		err := newStruct.Main()
		if err != nil {
			return fmt.Errorf("placeBlockWithNBTData: %v", err)
		}
		// 命令方块
	case "Container":
		newStruct := blockNBT_Container.Container{BlockEntityDatas: pack}
		err := newStruct.Main()
		if err != nil {
			return fmt.Errorf("placeBlockWithNBTData: %v", err)
		}
		// 各类已支持的且可被 replaceitem 生效的容器
	case "Sign":
		newStruct := blockNBT_Sign.Sign{BlockEntityDatas: pack}
		err := newStruct.Main()
		if err != nil {
			return fmt.Errorf("placeBlockWithNBTData: %v", err)
		}
		// 告示牌
	default:
		err := pack.API.SetBlockFastly(pack.Datas.Position, pack.Block.Name, pack.Datas.StatesString)
		if err != nil {
			return fmt.Errorf("placeBlockWithNBTData: %v", err)
		}
		return nil
		// 其他没有支持的方块实体
	}
	return nil
}

var apiIsUsing sync.Mutex

// 此函数是 package blockNBT 的主函数
func PlaceBlockWithNBTDataRun(api *blockNBT_API.GlobalAPI, blockInfo *types.Module, datas *blockNBT_global.Datas) error {
	defer apiIsUsing.Unlock()
	apiIsUsing.Lock()
	// lock(or unlock) api
	generalBlock, err := parseBlockModule(blockInfo)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTDataRun: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", *blockInfo.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	// get general block
	newRequest := blockNBT_global.BlockEntityDatas{
		API:   api,
		Block: generalBlock,
		Datas: datas,
	}
	newRequest.Datas.StatesString = blockInfo.Block.BlockStates
	newRequest.Datas.Position = [3]int32{int32(blockInfo.Point.X), int32(blockInfo.Point.Y), int32(blockInfo.Point.Z)}
	newRequest.Datas.Type = checkIfIsEffectiveNBTBlock(newRequest.Block.Name)
	// get new request of place nbt block
	err = placeBlockWithNBTData(&newRequest)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTDataRun: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", newRequest.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	// place block with nbt datas
	return nil
	// return
}
