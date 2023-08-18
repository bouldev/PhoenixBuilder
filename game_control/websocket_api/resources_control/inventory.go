package ResourcesControl

/*
Func List

func (*inventoryContents).GetInventoryInfo(windowID uint32) (map[uint8]protocol.ItemInstance, error)
func (*inventoryContents).GetItemStackInfo(windowID uint32, slotLocation uint8) (protocol.ItemInstance, error)
func (*inventoryContents).ListSlot(windowID uint32, filter *[]int32) ([]uint8, error)
func (*inventoryContents).ListWindowID() []uint32
*/

type Inventory_GII struct {
	WindowID uint32 `json:"window_id"`
}
