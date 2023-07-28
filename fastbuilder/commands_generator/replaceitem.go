package commands_generator

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
