package NBTAssigner

import (
	"fmt"
)

// DefaultItem 结构体用于描述一个完整的 NBT 物品的数据。
// 任何未被支持的 NBT 物品都会被重定向为此结构体
type DefaultItem struct {
	ItemPackage *ItemPackage // 该 NBT 物品的详细数据
}

// 这只是为了保证接口一致而设
func (d *DefaultItem) Decode() error {
	return nil
}

// 这只是为了保证接口一致而设
func (d *DefaultItem) SpecialCheck() (bool, error) {
	return false, nil
}

// 生成目标物品到快捷栏但不写入 NBT 数据
func (d *DefaultItem) WriteData() error {
	item := d.ItemPackage.Item
	// 初始化
	err := d.ItemPackage.ReplaceItemInInventory()
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 获取物品到物品栏，并附加物品组件数据
	if item.Enhancement != nil && item.Enhancement.Enchantments != nil {
		err = d.ItemPackage.AddEnchantments()
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 附加附魔属性
	if item.Enhancement != nil && item.Enhancement.ItemComponents != nil && len(item.Enhancement.ItemComponents.ItemLock) != 0 {
		return nil
	}
	// 如果该物品存在 item_lock 物品组件，
	// 则后续 NBT 无需附加，
	// 因为带有该物品组件的物品不能跨容器移动
	if item.Enhancement != nil && len(item.Enhancement.DisplayName) != 0 {
		err = d.ItemPackage.AddDisplayName()
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
	}
	// 附加物品的自定义显示名称
	return nil
	// 返回值
}
