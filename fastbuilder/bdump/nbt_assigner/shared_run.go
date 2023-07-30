package NBTAssigner

import (
	"fmt"
	env_interfaces "phoenixbuilder/fastbuilder/environment/interfaces"
	"phoenixbuilder/fastbuilder/types"
	"sync"
)

var interfaceLock sync.Mutex

// 带有 NBT 数据放置方块。
// 若你也想参与对于方块实体的其他支持，
// 另见 https://github.com/df-mc/dragonfly
func PlaceBlockWithNBTData(
	intf env_interfaces.GameInterface,
	blockInfo *types.Module,
	additionalData *BlockAdditionalData,
) error {
	defer interfaceLock.Unlock()
	interfaceLock.Lock()
	// lock(or unlock) api
	generalBlock, err := ParseBlockModule(blockInfo)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", *blockInfo.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	// get general block
	newRequest := BlockEntity{
		Interface:      intf,
		Block:          generalBlock,
		AdditionalData: *additionalData,
	}
	newRequest.AdditionalData.BlockStates = blockInfo.Block.BlockStates
	newRequest.AdditionalData.Position = [3]int32{int32(blockInfo.Point.X), int32(blockInfo.Point.Y), int32(blockInfo.Point.Z)}
	newRequest.AdditionalData.Type = IsNBTBlockSupported(newRequest.Block.Name)
	// get new request of place nbt block
	var placeBlockMethod GeneralBlockNBT
	if additionalData.Settings.AssignNBTData || newRequest.AdditionalData.Type == "CommandBlock" {
		placeBlockMethod = GetPlaceBlockMethod(&newRequest)
		err = placeBlockMethod.Decode()
		if err != nil {
			return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", *blockInfo.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
		}
		// if the user wants us to assign NBT data,
		// or the target block is a command block
	} else {
		placeBlockMethod = &DefaultBlock{BlockEntity: &newRequest}
		// if the user does not want us to assign NBT data
	}
	// get method and decode nbt data into golang struct
	err = placeBlockMethod.WriteData()
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTData: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", *blockInfo.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	// assign nbt data
	return nil
	// return
}

// 生成带有 NBT 数据的物品。
// 若你也想参与对于 NBT 物品的其他支持，
// 另见 https://github.com/df-mc/dragonfly
func GenerateItemWithNBTData(
	intf env_interfaces.GameInterface,
	singleItem ItemOrigin,
	additionalData *ItemAdditionalData,
) error {
	defer interfaceLock.Unlock()
	interfaceLock.Lock()
	// lock(or unlock) api
	newRequest := ItemPackage{
		Interface:      intf,
		Item:           GeneralItem{},
		AdditionalData: *additionalData,
	}
	err := newRequest.ParseItemFromNBT(singleItem)
	if err != nil {
		return fmt.Errorf("GenerateItemWithNBTData: Failed to generate the NBT item in hotbar %d, and the error log is %v", additionalData.HotBarSlot, err)
	}
	// get new request to generate new NBT item
	generateNBTItemMethod := GetGenerateItemMethod(&newRequest)
	if !additionalData.Decoded {
		err = generateNBTItemMethod.Decode()
		if err != nil {
			return fmt.Errorf("GenerateItemWithNBTData: Failed to generate the NBT item in hotbar %d, and the error log is %v", additionalData.HotBarSlot, err)
		}
	}
	// get method and decode nbt data into golang struct
	err = generateNBTItemMethod.WriteData()
	if err != nil {
		return fmt.Errorf("GenerateItemWithNBTData: Failed to generate the NBT item in hotbar %d, and the error log is %v", additionalData.HotBarSlot, err)
	}
	// assign nbt data
	return nil
	// return
}
