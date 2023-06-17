package blockNBT

import (
	"fmt"
	GlobalAPI "phoenixbuilder/GameControl/GlobalAPI"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/types"
	"strings"
	"sync"
)

// 将 types.Module 转换为 blockNBT_depends.GeneralBlock
func parseBlockModule(singleBlock *types.Module) (GeneralBlock, error) {
	// init var
	got, err := mcstructure.ParseStringNBT(singleBlock.Block.BlockStates, true)
	if err != nil {
		return GeneralBlock{}, fmt.Errorf("parseBlockModule: Could not parse block states; singleBlock.Block.BlockStates = %#v", singleBlock.Block.BlockStates)
	}
	blockStates, normal := got.(map[string]interface{})
	if !normal {
		return GeneralBlock{}, fmt.Errorf("parseBlockModule: The target block states is not map[string]interface{}; got = %#v", got)
	}
	// get block states
	return GeneralBlock{
		Name:   strings.Replace(strings.ToLower(strings.ReplaceAll(*singleBlock.Block.Name, " ", "")), "minecraft:", "", 1),
		States: blockStates,
		NBT:    singleBlock.NBTMap,
	}, nil
	// return
}

var apiIsUsing sync.Mutex

// 带有 NBT 数据放置方块。
// 若你也想参与对于方块实体的其他支持，
// 另见 https://github.com/df-mc/dragonfly
func PlaceBlockWithNBTData(
	api *GlobalAPI.GlobalAPI,
	blockInfo *types.Module,
	datas *Datas,
) error {
	defer apiIsUsing.Unlock()
	apiIsUsing.Lock()
	// lock(or unlock) api
	generalBlock, err := parseBlockModule(blockInfo)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", *blockInfo.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	// get general block
	newRequest := Package{
		API:   api,
		Block: generalBlock,
		Datas: datas,
	}
	newRequest.Datas.StatesString = blockInfo.Block.BlockStates
	newRequest.Datas.Position = [3]int32{int32(blockInfo.Point.X), int32(blockInfo.Point.Y), int32(blockInfo.Point.Z)}
	newRequest.Datas.Type = CheckIfIsEffectiveNBTBlock(newRequest.Block.Name)
	// get new request of place nbt block
	placeBlockMethod := GetMethod(newRequest)
	err = placeBlockMethod.Decode()
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", newRequest.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	err = placeBlockMethod.WriteDatas()
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", newRequest.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	// place block with nbt datas
	return nil
	// return
}
