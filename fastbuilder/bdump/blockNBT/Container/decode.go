package blockNBT_Container

import (
	"fmt"
	"phoenixbuilder/mirror/chunk"
	"strings"
)

/*
检查一个方块是否是已被支持的有效的容器；这里的“有效”指的是可以被 replaceitem 命令生效的容器

如果不是已被支持的容器，则返回 var NotASupportedContainer string = "Not a supported container"
*/
func (c *Container) checkIfIsSupportedContainer() string {
	value, ok := SupportContainerPool[c.BlockEntityDatas.Block.Name]
	if ok {
		return value
	}
	return NotASupportedContainer
}

// 从容器的 NBT 数据提取物品数据；结果会被替换在 c.BlockEntityDatas.Block.NBT[KeyName] 中
func (c *Container) replaceNBTMapToContainerList() error {
	key := c.checkIfIsSupportedContainer()
	if key == NotASupportedContainer {
		return ErrNotASupportedContainer
	}
	// 这里是确定一下这个容器是否是我们支持了的容器
	value, ok := c.BlockEntityDatas.Block.NBT[key]
	if !ok {
		c.BlockEntityDatas.Block.NBT = map[string]interface{}{KeyName: []interface{}{}}
	} else {
		c.BlockEntityDatas.Block.NBT = map[string]interface{}{KeyName: value}
	}
	// 对于唱片机和讲台这种容器，如果它们没有被放物品的话，那么对应的 key 是找不到的
	// 但是这并非是错误
	return nil
	// 返回值
}

// 从 c.BlockEntityDatas.Block.NBT 提取物品数据并保存在 c.Items 中
func (c *Container) Decode() error {
	var correct []interface{} = []interface{}{}
	// 初始化
	err := c.replaceNBTMapToContainerList()
	if err != nil {
		return fmt.Errorf("Decode: %v", err)
	}
	// 替换 NBT 数据为物品数据
	got, normal := c.BlockEntityDatas.Block.NBT[KeyName].([]interface{})
	if !normal {
		got, normal := c.BlockEntityDatas.Block.NBT[KeyName].(map[string]interface{})
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
		correct = append(correct, got)
	} else {
		correct = got
	}
	// 把物品放入 correct 中
	// 如果这个物品是一个唱片机或者讲台，那么传入的 container 是一个 map[string]interface{} 而非 []interface{}
	// 为了更好的兼容性(更加方便)，这里都会把 map[string]interface{} 处理成通常情况下的 []interface{}
	// correct 就是处理结果
	for key, value := range correct {
		var count uint8 = uint8(0)
		var itemData uint16 = uint16(0)
		var name string = ""
		var slot uint8 = uint8(0)
		// 初始化
		containerData, normal := value.(map[string]interface{})
		if !normal {
			return fmt.Errorf("Decode: Crashed in correct[%v]; correct[%v] = %#v", key, key, value)
		}
		// correct 这个列表中的每一项都必须是一个复合标签，也就得是 map[string]interface{} 才行
		_, ok := containerData["Count"]
		if !ok {
			return fmt.Errorf("Decode: Crashed in correct[%v][\"Count\"]; correct[%v] = %#v", key, key, containerData)
		}
		count_got, normal := containerData["Count"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in correct[%v][\"Count\"]; correct[%v] = %#v", key, key, containerData)
		}
		count = uint8(count_got)
		// 拿一下物品数量(物品数量是一定存在的)
		_, ok = containerData["Name"]
		if !ok {
			return fmt.Errorf("Decode: Crashed in correct[%v][\"Name\"]; correct[%v] = %#v", key, key, containerData)
		}
		got, normal := containerData["Name"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed in correct[%v][\"Name\"]; correct[%v] = %#v", key, key, containerData)
		}
		name = strings.Replace(strings.ToLower(got), "minecraft:", "", 1)
		// 拿一下这个物品的物品名称(命名空间 minecraft 已移除; 此数据必定存在)
		_, ok = containerData["Damage"]
		if !ok {
			return fmt.Errorf("Decode: Crashed in correct[%v][\"Damage\"]; correct[%v] = %#v", key, key, containerData)
		}
		damage_got, normal := containerData["Damage"].(int16)
		if !normal {
			return fmt.Errorf("Decode: Crashed in correct[%v][\"Damage\"]; correct[%v] = %#v", key, key, containerData)
		}
		itemData = uint16(damage_got)
		// 拿一下物品的 Damage 值; Damage 值不一定就是物品的数据值(附加值); 此数据必定存在 [需要验证]
		_, ok = containerData["tag"]
		if ok {
			tag, normal := containerData["tag"].(map[string]interface{})
			if !normal {
				return fmt.Errorf("Decode: Crashed in correct[%v][\"tag\"]; correct[%v] = %#v", key, key, containerData)
			}
			// 这个 container["tag"] 一定是一个复合标签 [需要验证]
			_, ok = tag["Damage"]
			if ok {
				got, normal := tag["Damage"].(int32)
				if !normal {
					return fmt.Errorf("Decode: Crashed in correct[%v][\"tag\"]; correct[%v] = %#v", key, key, containerData)
				}
				itemData = uint16(got)
			}
		}
		// 拿一下这个工具的耐久值（当然也可能是别的，甚至它都不是个工具）
		// 这个 tag 里的 Damage 实际上也不一定就是物品的数据值(附加值)
		// 需要说明的是，tag 不一定存在，且 tag 存在，Damage 也不一定存在
		_, ok = containerData["Block"]
		if ok {
			Block, normal := containerData["Block"].(map[string]interface{})
			if !normal {
				return fmt.Errorf("Decode: Crashed in correct[%v][\"Block\"]; correct[%v] = %#v", key, key, containerData)
			}
			// 这个 container["Block"] 一定是一个复合标签；如果 Block 找得到则说明这个物品是一个方块
			_, ok = Block["val"]
			if ok {
				got, normal := Block["val"].(int16)
				if !normal {
					return fmt.Errorf("Decode: Crashed in correct[%v][\"Block\"][\"val\"]; correct[%v][\"Block\"] = %#v", key, key, Block)
				}
				itemData = uint16(got)
				// 如果这个物品是个方块，也就是 Block 找得到的话
				// 那在 Block 里面一定有一个 val 去声明这个方块的方块数据值(附加值) [仅限 Netease MC]
			} else {
				_, ok = Block["states"]
				if !ok {
					itemData = 0
				} else {
					got, normal := Block["states"].(map[string]interface{})
					if !normal {
						itemData = 0
					} else {
						runtimeId, found := chunk.StateToRuntimeID(name, got)
						if !found {
							return fmt.Errorf("Decode: Could not convert legacy block to standard runtime id; got = %#v", got)
						}
						legacyBlock, found := chunk.RuntimeIDToLegacyBlock(runtimeId)
						if !found {
							return fmt.Errorf("Decode: Could not convert standard runtime id to block states; got = %#v", got)
						}
						itemData = legacyBlock.Val
					}
				}
			}
		}
		// 拿一下这个方块的方块数据值(附加值)
		// 这个 Block 里的 val 一定是这个物品对应的方块的方块数据值(附加值)
		// 需要说明的是，Block 不一定存在，但如果 Block 存在，则 val 一定存在 [仅 Netease MC]
		/*
			以上三个都在拿物品数据值(附加值)
			需要说明的是，数据值的获取优先级是这样的
			Damage < tag["Damage"] < Block["val"]
			需要说明的是，以上列举的三个情况不能涵盖所有的物品数据值(附加值)的情况，所以我希望可以有个人看一下普世情况是长什么样的，请帮帮我！
		*/
		_, ok = containerData["Slot"]
		if ok {
			got, normal := containerData["Slot"].(byte)
			if !normal {
				return fmt.Errorf("Decode: Crashed in correct[%v][\"Slot\"]; correct[%v] = %#v", key, key, containerData)
			}
			slot = uint8(got)
		}
		// 拿一下这个物品所在的栏位(槽位)
		// 这个栏位(槽位)不一定存在，例如唱片机和讲台这种就不存在了(这种方块就一个物品，就不需要这个数据了)
		c.Items = append(c.Items, Item{
			Name:   name,
			Count:  count,
			Damage: itemData,
			Slot:   slot,
		})
		// 提交数据
	}
	// get datas
	return nil
	// return
}
