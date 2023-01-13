package nbttranslatinginterface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	"strings"
)

// 检查一个方块是否是有效的容器；这里的有效指的是可以被 replaceitem 命令生效的容器
func CheckIfIsEffectiveContainer(name string) (string, error) {
	index := map[string]string{
		"blast_furnace":      "Items",
		"lit_blast_furnace":  "Items",
		"smoker":             "Items",
		"lit_smoker":         "Items",
		"furnace":            "Items",
		"lit_furnace":        "Items",
		"chest":              "Items",
		"barrel":             "Items",
		"trapped_chest":      "Items",
		"lectern":            "book",
		"hopper":             "Items",
		"dispenser":          "Items",
		"dropper":            "Items",
		"cauldron":           "Items",
		"lava_cauldron":      "Items",
		"jukebox":            "RecordItem",
		"brewing_stand":      "Items",
		"undyed_shulker_box": "Items",
		"shulker_box":        "Items",
	}
	value, ok := index[name]
	if ok {
		return value, nil
	}
	return "", fmt.Errorf("CheckIfIsEffectiveContainer: \"%v\" not found", name)
}

// 将 Interface NBT 转换为 types.ChestData
func GetContainerData(container interface{}) (types.ChestData, error) {
	var correct []interface{} = make([]interface{}, 0)
	// 初始化
	got, normal := container.([]interface{})
	if !normal {
		got, normal := container.(map[string]interface{})
		if !normal {
			return types.ChestData{}, fmt.Errorf("Crashed in input")
		}
		correct = append(correct, got)
	} else {
		correct = got
	}
	// 把物品丢入 correct 里面
	// 如果这个物品是一个唱片机或者讲台，那么传入的 container 是一个 map[string]interface{} 而非 []interface{}
	// 为了更好的兼容性(更加方便)，这里都会把 map[string]interface{} 处理成通常情况下的 []interface{}
	// correct 就是处理结果
	ans := make(types.ChestData, 0)
	for key, value := range correct {
		var count uint8 = uint8(0)
		var itemData uint16 = uint16(0)
		var name string = ""
		var slot uint8 = uint8(0)
		// 初始化
		containerData, normal := value.(map[string]interface{})
		if !normal {
			return types.ChestData{}, fmt.Errorf("Crashed in input[%v]", key)
		}
		// correct 这个列表中的每一项都必须是一个复合标签，也就得是 map[string]interface{} 才行
		_, ok := containerData["Count"]
		if !ok {
			return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Count\"]", key)
		}
		count_got, normal := containerData["Count"].(byte)
		if !normal {
			return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Count\"]", key)
		}
		count = uint8(count_got)
		// 拿一下物品数量
		// 这个物品数量是一定存在的，拿不到必须报错哦
		_, ok = containerData["Damage"]
		if !ok {
			return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Damage\"]", key)
		}
		damage_got, normal := containerData["Damage"].(int16)
		if !normal {
			return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Damage\"]", key)
		}
		itemData = uint16(damage_got)
		// 拿一下物品的 Damage 值
		// 这里的 Damage 值不一定就是物品的数据值(附加值)
		// 不过这个 Damage 值是一定存在的，拿不到必须报错哦
		// 真的一定存在吗（？
		_, ok = containerData["tag"]
		if ok {
			tag, normal := containerData["tag"].(map[string]interface{})
			if !normal {
				return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"tag\"]", key)
			}
			// 这个 input["tag"] 一定是一个复合标签，如果不是就必须报错哦
			// 真的是吗（？
			_, ok = tag["Damage"]
			if ok {
				got, normal := tag["Damage"].(int32)
				if !normal {
					return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"tag\"]", key)
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
				return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Block\"]", key)
			}
			// 这个 input["Block"] 一定是一个复合标签，如果不是就必须报错哦
			// 如果 Block 找得到则说明这个物品是一个方块
			_, ok = Block["val"]
			if !ok {
				return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Block\"][\"val\"]", key)
			}
			got, normal := Block["val"].(int16)
			if !normal {
				return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Block\"][\"val\"]", key)
			}
			itemData = uint16(got)
			// 如果这个物品是个方块，也就是 Block 找得到的话
			// 那在 Block 里面一定有一个 val 去声明这个方块的方块数据值(附加值)
		}
		// 拿一下这个方块的方块数据值(附加值)
		// 这个 Block 里的 val 一定是这个物品对应的方块的方块数据值(附加值)
		// 需要说明的是，Block 不一定存在，但如果 Block 存在，则 val 一定存在
		// 除非网易尝试打击我们，把 val 扣掉了

		// 以上三个都在拿物品数据值(附加值)
		// 需要说明的是，数据值的获取优先级是这样的
		// Damage < tag["Damage"] < Block["val"]
		// 需要说明的是，以上列举的三个情况不能涵盖所有的物品数据值(附加值)的情况，所以我希望可以有个人看一下普世情况是长什么样的，请帮帮我！
		_, ok = containerData["Name"]
		if !ok {
			return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Name\"]", key)
		}
		got, normal := containerData["Name"].(string)
		if !normal {
			return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Name\"]", key)
		}
		name = strings.Replace(got, "minecraft:", "", 1)
		// 拿一下这个物品的物品名称
		// 可以看到，我这里是把命名空间删了的
		// 这个物品名称是一定存在的，拿不到必须报错哦
		_, ok = containerData["Slot"]
		if ok {
			got, normal := containerData["Slot"].(byte)
			if !normal {
				return types.ChestData{}, fmt.Errorf("Crashed in input[%v][\"Slot\"]", key)
			}
			slot = uint8(got)
		}
		// 拿一下这个物品所在的栏位(槽位)
		// 这个栏位(槽位)不一定存在，例如唱片机和讲台这种就不存在了(这种方块就一个物品，就不需要这个数据了)
		ans = append(ans, types.ChestSlot{
			Name:   name,
			Count:  count,
			Damage: itemData,
			Slot:   slot,
		})
		// 提交数据
	}
	return ans, nil
}

// 主函数
func GetContainerDataRun(blockNBT map[string]interface{}, blockName string) (types.ChestData, error) {
	key, err := CheckIfIsEffectiveContainer(blockName)
	if err != nil {
		return types.ChestData{}, fmt.Errorf("GetContainerDataRun: Not a container")
	}
	got, ok := blockNBT[key]
	// 这里是确定一下这个容器是否是我们支持了的容器
	if ok {
		ans, err := GetContainerData(got)
		if err != nil {
			return types.ChestData{}, fmt.Errorf("GetContainerData(Started by GetContainerDataRun): %v", err)
		}
		return ans, nil
	}
	// 如果这是个容器且对应的 key 可以找到，那么就去拿一下对应的 types.ChestData 结构体
	return types.ChestData{}, nil
	// 对于唱片机和讲台这种容器，如果它们没有被放物品的话，那么对应的 key 是找不到的
	// 但是这不是一个错误，所以我们返回一个空的 types.ChestData 和一个空的 error
}
