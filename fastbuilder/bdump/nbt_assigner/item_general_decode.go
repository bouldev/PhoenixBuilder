package NBTAssigner

import (
	"encoding/gob"
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/mirror/chunk"
	"strings"
)

// 从 singleItem 解码单个物品的基本数据
func (g *GeneralItem) DecodeItemBasicData(singleItem ItemOrigin) error {
	{
		count_origin, ok := singleItem["Count"]
		if !ok {
			return fmt.Errorf(`DecodeItemBasicData: singleItem["Count"] does not exist; singleItem = %#v`, singleItem)
		}
		count_got, normal := count_origin.(byte)
		if !normal {
			return fmt.Errorf(`DecodeItemBasicData: Can not convert count_origin into byte(uint8); singleItem = %#v`, singleItem)
		}
		g.Basic.Count = count_got
	}
	// 物品数量
	{
		name_origin, ok := singleItem["Name"]
		if !ok {
			return fmt.Errorf(`DecodeItemBasicData: singleItem["Name"] does not exist; singleItem = %#v`, singleItem)
		}
		name_got, normal := name_origin.(string)
		if !normal {
			return fmt.Errorf(`DecodeItemBasicData: Can not convert name_origin into string; singleItem = %#v`, singleItem)
		}
		g.Basic.Name = strings.Replace(strings.ToLower(name_got), "minecraft:", "", 1)
	}
	// 物品的英文 ID (已去除命名空间)
	if slot_origin, ok := singleItem["Slot"]; ok {
		slot_got, normal := slot_origin.(byte)
		if !normal {
			return fmt.Errorf(`DecodeItemBasicData: Can not convert slot_origin into byte(uint8); singleItem = %#v`, singleItem)
		}
		g.Basic.Slot = slot_got
	}
	// 物品所在的槽位(对于唱片机等单槽位方块来说，此数据不存在)
	{
		{
			damage_origin, ok := singleItem["Damage"]
			if !ok {
				return fmt.Errorf(`DecodeItemBasicData: singleItem["Damage"] does not exist; singleItem = %#v`, singleItem)
			}
			damage_got, normal := damage_origin.(int16)
			if !normal {
				return fmt.Errorf(`DecodeItemBasicData: Can not convert damage_origin into int16; singleItem = %#v`, singleItem)
			}
			g.Basic.MetaData = uint16(damage_got)
		}
		// Damage
		for i := 0; i < 1; i++ {
			tag_origin, ok := singleItem["tag"]
			if !ok {
				break
			}
			tag_got, normal := tag_origin.(map[string]interface{})
			if !normal {
				return fmt.Errorf(`DecodeItemBasicData: Can not convert tag_origin into map[string]interface{}; singleItem = %#v`, singleItem)
			}
			damage_origin, ok := tag_got["Damage"]
			if !ok {
				break
			}
			damage_got, normal := damage_origin.(int32)
			if !normal {
				return fmt.Errorf(`DecodeItemBasicData: Can not convert damage_origin into int32; singleItem = %#v`, singleItem)
			}
			g.Basic.MetaData = uint16(damage_got)
		}
		// tag["Damage"]
		for i := 0; i < 1; i++ {
			block_origin, ok := singleItem["Block"]
			if !ok {
				break
			}
			block_got, normal := block_origin.(map[string]interface{})
			if !normal {
				return fmt.Errorf(`DecodeItemBasicData: Can not convert block_got into map[string]interface{}; singleItem = %#v`, singleItem)
			}
			if val_origin, ok := block_got["val"]; ok {
				val_got, normal := val_origin.(int16)
				if !normal {
					return fmt.Errorf(`DecodeItemBasicData: Can not convert val_origin into int16; singleItem = %#v`, singleItem)
				}
				g.Basic.MetaData = uint16(val_got)
			} else {
				block_name_origin, ok := block_got["name"]
				if !ok {
					break
				}
				block_name_got, normal := block_name_origin.(string)
				if !normal {
					return fmt.Errorf(`DecodeItemBasicData: Can not convert block_name_origin into string; singleItem = %#v`, singleItem)
				}
				states_origin, ok := block_got["states"]
				if !ok {
					break
				}
				states_got, normal := states_origin.(map[string]interface{})
				if !normal {
					return fmt.Errorf(`DecodeItemBasicData: Can not convert states_origin into map[string]interface{}; singleItem = %#v`, singleItem)
				}
				runtimeId, found := chunk.StateToRuntimeID(block_name_got, states_got)
				if !found {
					return fmt.Errorf(`DecodeItemBasicData: Could not convert legacy block to standard runtime id; singleItem = %#v`, singleItem)
				}
				legacyBlock, found := chunk.RuntimeIDToLegacyBlock(runtimeId)
				if !found {
					return fmt.Errorf(`DecodeItemBasicData: Could not convert standard runtime id to block states; singleItem = %#v`, singleItem)
				}
				g.Basic.MetaData = legacyBlock.Val
			}
		}
		// Block["val"] or Block["states"]
	}
	/*
		物品数据值(附加值)

		以上三个方法都在拿物品数据值(附加值)，而数据值的获取优先级如下。
		Damage < tag["Damage"] < Block["val"]

		顶层复合标签下的 Damage 数据一定存在，但不一定代表物品真实的物品数据值。

		如果当前物品是武器或者工具，其 tag["Damage"] 处会说明其耐久值，
		而这个耐久值才是真正的物品数据值；
		如果当前物品是一个方块，其 block["val"] 处可能会声明其方块数据值，
		而这个方块数据值才是真正的物品数据值。
		当然，如果这个 BDX 文件是从国际版制作的，那么 block["val"] 处将不存在数据，
		此时将需要从 Block["states"] 处得到此物品所对应方块的方块数据值。

		NOTE: 不保证目前已经提供的这三个方法涵盖了所有情况，一切还需要进一步的研究
	*/
	return nil
	// 返回值
}

// 从 singleItem 解码单个物品的增强数据，
// 其中包含物品组件、显示名称和附魔属性。
// 特别地，如果此物品存在 item_lock 物品组件，
// 则只会解析物品组件和附魔的相关数据，
// 因为存在 item_lock 的物品无法使用铁砧修改名称
func (g *GeneralItem) DecodeItemEnhancementData(
	singleItem ItemOrigin,
) error {
	var displayName string
	var enchantments *[]Enchantment
	var itemComponents *ItemComponents
	var nbt_tag_got map[string]interface{}
	var normal bool
	// 初始化
	nbt_tag_origin, ok := singleItem["tag"]
	if ok {
		nbt_tag_got, normal = nbt_tag_origin.(map[string]interface{})
		if !normal {
			return fmt.Errorf(`DecodeItemEnhancementData: Can not convert nbt_tag_origin into map[string]interface{}; singleItem = %#v`, singleItem)
		}
	}
	// 获取当前物品的 tag 数据
	{
		if can_place_on_origin, ok := singleItem["CanPlaceOn"]; ok {
			can_place_on_got, normal := can_place_on_origin.([]interface{})
			if !normal {
				return fmt.Errorf(`DecodeItemEnhancementData: Can not convert can_place_on_origin into []interface{}; singleItem = %#v`, singleItem)
			}
			if itemComponents == nil {
				itemComponents = &ItemComponents{}
			}
			for key, value := range can_place_on_got {
				blockName, normal := value.(string)
				if !normal {
					return fmt.Errorf(`DecodeItemEnhancementData: Can not convert can_place_on_got[%d] into string; singleItem = %#v`, key, singleItem)
				}
				itemComponents.CanPlaceOn = append(itemComponents.CanPlaceOn, blockName)
			}
		}
		// can_place_on
		if can_destroy_origin, ok := singleItem["CanDestroy"]; ok {
			can_destroy_got, normal := can_destroy_origin.([]interface{})
			if !normal {
				return fmt.Errorf(`DecodeItemEnhancementData: Can not convert can_destroy_origin into []interface{}; singleItem = %#v`, singleItem)
			}
			if itemComponents == nil {
				itemComponents = &ItemComponents{}
			}
			for key, value := range can_destroy_got {
				blockName, normal := value.(string)
				if !normal {
					return fmt.Errorf(`DecodeItemEnhancementData: Can not convert can_destroy_got[%d] into string; singleItem = %#v`, key, singleItem)
				}
				itemComponents.CanDestroy = append(itemComponents.CanDestroy, blockName)
			}
		}
		// can_destroy
		if nbt_tag_got != nil {
			if item_lock_origin, ok := nbt_tag_got["minecraft:item_lock"]; ok {
				item_lock_got, normal := item_lock_origin.(byte)
				if !normal {
					return fmt.Errorf(`DecodeItemEnhancementData: Can not convert item_lock_origin into byte(uint8); singleItem = %#v`, singleItem)
				}
				if itemComponents == nil {
					itemComponents = &ItemComponents{}
				}
				switch item_lock_got {
				case 1:
					itemComponents.ItemLock = "lock_in_slot"
				case 2:
					itemComponents.ItemLock = "lock_in_inventory"
				default:
					return fmt.Errorf(`DecodeItemEnhancementData: Unknown value(%d) of item_lock; singleItem = %#v`, item_lock_got, singleItem)
				}
			}
			// item_lock
			if keep_on_death_origin, ok := nbt_tag_got["minecraft:keep_on_death"]; ok {
				keep_on_death_got, normal := keep_on_death_origin.(byte)
				if !normal {
					return fmt.Errorf(`DecodeItemEnhancementData: Can not convert keep_on_death_origin into byte(uint8); singleItem = %#v`, singleItem)
				}
				if keep_on_death_got != 1 {
					return fmt.Errorf(`DecodeItemEnhancementData: Unknown value(%d) of kepp_on_death; singleItem = %#v`, keep_on_death_got, singleItem)
				}
				if itemComponents == nil {
					itemComponents = &ItemComponents{}
				}
				itemComponents.KeepOnDeath = true
			}
			// keep_on_death
		}
		// item_lock and keep_on_death
	}
	// 物品组件
	if ench_origin, ok := nbt_tag_got["ench"]; ok {
		ench_got, normal := ench_origin.([]interface{})
		if !normal {
			return fmt.Errorf(`DecodeItemEnhancementData: Can not convert ench_origin into []interface{}; singleItem = %#v`, singleItem)
		}
		if len(ench_got) > 0 {
			enchantments = &[]Enchantment{}
		}
		for key, value := range ench_got {
			value_got, normal := value.(map[string]interface{})
			if !normal {
				return fmt.Errorf(`DecodeItemEnhancementData: Can not convert ench_got[%d] into map[string]interface{}; singleItem = %#v`, key, singleItem)
			}
			id_origin, ok := value_got["id"]
			if !ok {
				return fmt.Errorf(`DecodeItemEnhancementData: ench_got[%d]["id"] does not exist; singleItem = %#v`, key, singleItem)
			}
			id_got, normal := id_origin.(int16)
			if !normal {
				return fmt.Errorf(`DecodeItemEnhancementData: Can not convert ench_got[%d]["id"] into int16; singleItem = %#v`, key, singleItem)
			}
			lvl_origin, ok := value_got["lvl"]
			if !ok {
				return fmt.Errorf(`DecodeItemEnhancementData: ench_got[%d]["lvl"] does not exist; singleItem = %#v`, key, singleItem)
			}
			lvl_got, normal := lvl_origin.(int16)
			if !normal {
				return fmt.Errorf(`DecodeItemEnhancementData: Can not convert ench_got[%d]["lvl"] into int16; singleItem = %#v`, key, singleItem)
			}
			*enchantments = append(*enchantments, Enchantment{ID: uint8(id_got), Level: lvl_got})
		}
	}
	// 物品的附魔属性
	if itemComponents != nil && len(itemComponents.ItemLock) != 0 {
		g.Enhancement = &ItemEnhancementData{
			DisplayName:    "",
			Enchantments:   enchantments,
			ItemComponents: itemComponents,
		}
		return nil
	}
	// 如果当前物品已经使用了 item_lock 物品组件，
	// 则无需再解析后续的数据，
	// 因为存在 item_lock 的物品无法使用铁砧修改名称
	for i := 0; i < 1; i++ {
		display_origin, ok := nbt_tag_got["display"]
		if !ok {
			break
		}
		display_got, normal := display_origin.(map[string]interface{})
		if !normal {
			return fmt.Errorf(`DecodeItemEnhancementData: Can not convert display_origin into map[string]interface{}; singleItem = %#v`, singleItem)
		}
		name_origin, ok := display_got["Name"]
		if !ok {
			break
		}
		name_got, normal := name_origin.(string)
		if !normal {
			return fmt.Errorf(`DecodeItemEnhancementData: Can not convert name_origin into string; singleItem = %#v`, singleItem)
		}
		displayName = name_got
	}
	// 物品的显示名称
	if len(displayName) != 0 || enchantments != nil || itemComponents != nil {
		g.Enhancement = &ItemEnhancementData{
			DisplayName:    displayName,
			Enchantments:   enchantments,
			ItemComponents: itemComponents,
		}
		return nil
	}
	return nil
	// 返回值
}

// 从 singleItem 解码单个物品的自定义 NBT 数据。
// 此函数应当后于 DecodeItemBasicData 执行
func (i *ItemPackage) DecodeItemCustomData(
	singleItem ItemOrigin,
) error {
	var blockStates map[string]interface{}
	var blockStatesString string
	var err error
	// 初始化
	nbt_tag_origin, ok := singleItem["tag"]
	if !ok || !i.AdditionalData.Settings.AssignNBTData || i.AdditionalData.FastMode {
		return nil
	}
	nbt_tag_got, normal := nbt_tag_origin.(map[string]interface{})
	if !normal {
		return fmt.Errorf(`DecodeItemCustomData: Can not convert nbt_tag_origin into map[string]interface{}; singleItem = %#v`, singleItem)
	}
	// 获取当前物品的 tag 数据
	{
		blockName := ItemNameToBlockNamePool[i.Item.Basic.Name]
		blockType := IsNBTBlockSupported(blockName)
		// 取得该方块实体的类型
		if len(blockType) != 0 {
			var copyOne map[string]interface{}
			ResourcesControl.DeepCopy(
				&nbt_tag_got,
				&copyOne,
				func() {
					gob.Register(map[string]interface{}{})
					gob.Register([]interface{}{})
				},
			)
			delete(copyOne, "ench")
			delete(copyOne, "display")
			delete(copyOne, "RepairCost")
			delete(copyOne, "minecraft:item_lock")
			delete(copyOne, "minecraft:keep_on_death")
			if len(copyOne) == 0 {
				return nil
			}
			// 检查当前方块实体是否真的需要注入 NBT 数据
			blockStates, err = get_block_states_from_legacy_block(
				blockName, i.Item.Basic.MetaData,
			)
			if err != nil {
				blockStates = map[string]interface{}{}
			}
			blockStatesString, err = mcstructure.ConvertCompoundToString(blockStates, true)
			if err != nil {
				blockStates = map[string]interface{}{}
				blockStatesString = "[]"
			}
			// 取得当前方块实体的方块状态及其字符串形式
			i.Item.Custom = &ItemCustomData{
				SubBlockData: GetPlaceBlockMethod(
					&BlockEntity{
						Interface: i.Interface,
						Block: GeneralBlock{
							Name:   blockName,
							States: blockStates,
							NBT:    nbt_tag_got,
						},
						AdditionalData: BlockAdditionalData{
							BlockStates: blockStatesString,
							Position:    i.AdditionalData.Position,
							Type:        blockType,
							Settings:    i.AdditionalData.Settings,
							FastMode:    false,
							Others:      i.AdditionalData.Others,
						},
					},
				),
				ItemTag: nil,
			}
			return nil
		}
		// 赋值并返回
	}
	// 如果该物品是一个 NBT 方块
	{
		i.Item.Custom = &ItemCustomData{
			SubBlockData: nil,
			ItemTag:      nbt_tag_got,
		}
		i.AdditionalData.Type = IsNBTItemSupported(i.Item.Basic.Name)
		needSpecialTreatment, err := GetGenerateItemMethod(i).SpecialCheck()
		if err != nil {
			return fmt.Errorf("DecodeItemCustomData: %v", err)
		}
		if !needSpecialTreatment {
			i.Item.Custom = nil
			return nil
		}
	}
	// 如果该物品是一个 NBT 物品，例如通过工作台合成的烟花
	return nil
	// 返回值
}
