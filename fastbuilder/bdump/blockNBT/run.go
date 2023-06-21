package blockNBT

import (
	"fmt"
	GlobalAPI "phoenixbuilder/GameControl/GlobalAPI"
	"phoenixbuilder/fastbuilder/types"
	"sync"
)

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
	newRequest.Datas.Type = checkIfIsEffectiveNBTBlock(newRequest.Block.Name)
	// get new request of place nbt block
	var placeBlockMethod GeneralBlockNBT
	if datas.Settings.AssignNBTData {
		placeBlockMethod = getMethod(newRequest)
		err = placeBlockMethod.Decode()
		if err != nil {
			return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", newRequest.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
		}
		// if the user wants us to assign NBT data
	} else {
		if newRequest.Datas.Type == "CommandBlock" {
			placeBlockMethod = &CommandBlock{Package: &newRequest, NeedToPlaceBlock: true}
		} else {
			placeBlockMethod = &Default{Package: &newRequest}
		}
		// uf the user does not want us to assign NBT data
	}
	// get method and decode nbt data into golang struct
	err = placeBlockMethod.WriteDatas()
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", newRequest.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	// assign nbt data
	return nil
	// return
}
