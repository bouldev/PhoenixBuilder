package command

// extern void replaceItemRequestInternal(void *preallocatedStr, int x, int y, int z, unsigned char slot, const char *name, unsigned char count, unsigned short damage);
import "C"
import (
	"unsafe"
	"phoenixbuilder/fastbuilder/types"
)


func ReplaceItemRequest(buf *string, module *types.Module, config *types.MainConfig) {
	C.replaceItemRequestInternal(unsafe.Pointer(buf), C.int(module.Point.X), C.int(module.Point.Y), C.int(module.Point.Z), C.uchar(module.ChestSlot.Slot),C.CString(module.ChestSlot.Name),C.uchar(module.ChestSlot.Count), C.ushort(module.ChestSlot.Damage))
	//return fmt.Sprintf("replaceitem block %d %d %d slot.container %d %s %d %d", module.Point.X, module.Point.Y, module.Point.Z, module.ChestSlot.Slot, module.ChestSlot.Name, module.ChestSlot.Count, module.ChestSlot.Damage)
}