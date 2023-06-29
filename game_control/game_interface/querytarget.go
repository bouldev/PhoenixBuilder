package GameInterface

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 用于描述单个 querytarget 结果的结构体
type TargetQueryingInfo struct {
	Dimension byte
	Position  [3]float32
	UniqueId  string
	YRot      float32
}

// 解析 querytarget 命令的返回值为列表，因为同一时刻可以查询多个实体的相关数据。
// 列表内单个数据的数据类型为 QuerytargetInfo 结构体
func (g *GameInterface) ParseTargetQueryingInfo(pk packet.CommandOutput) ([]TargetQueryingInfo, error) {
	res := []TargetQueryingInfo{}
	// 初始化
	if pk.SuccessCount <= 0 || len(pk.OutputMessages[0].Parameters) <= 0 {
		return []TargetQueryingInfo{}, nil
	}
	// 如果命令失败或者未能找到任何可以解析的信息
	datas := pk.OutputMessages[0].Parameters[0]
	var datasDecodeAns []interface{}
	err := json.Unmarshal([]byte(datas), &datasDecodeAns)
	if err != nil {
		return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: %v", err)
	}
	// 解析 JSON 数据
	for _, value := range datasDecodeAns {
		newStruct := TargetQueryingInfo{}
		// 初始化
		val, normal := value.(map[string]interface{})
		if !normal {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Could not convert value into map[string]interface{}; value = %#v", value)
		}
		// 将列表中的被遍历元素解析为 map[string]interface{}
		_, ok := val["dimension"]
		if !ok {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"dimension\"]; val = %#v", val)
		}
		dimension, normal := val["dimension"].(float64)
		if !normal {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"dimension\"]; val = %#v", val)
		}
		newStruct.Dimension = byte(dimension)
		// dimension
		_, ok = val["position"]
		if !ok {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"]; val = %#v", val)
		}
		position, normal := val["position"].(map[string]interface{})
		if normal {
			_, ok := position["x"]
			if !ok {
				return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"][\"x\"]; val[\"position\"] = %#v", position)
			}
			x, normal := position["x"].(float64)
			if !normal {
				return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"][\"x\"]; val[\"position\"] = %#v", position)
			}
			newStruct.Position = [3]float32{float32(x), 0, 0}
			// posx
			_, ok = position["y"]
			if !ok {
				return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"][\"y\"]; val[\"position\"] = %#v", position)
			}
			y, normal := position["y"].(float64)
			if !normal {
				return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"][\"y\"]; val[\"position\"] = %#v", position)
			}
			newStruct.Position[1] = float32(y)
			// posy
			_, ok = position["z"]
			if !ok {
				return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"][\"z\"]; val[\"position\"] = %#v", position)
			}
			z, normal := position["z"].(float64)
			if !normal {
				return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"][\"z\"]; val[\"position\"] = %#v", position)
			}
			newStruct.Position[2] = float32(z)
			// posz
		} else {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"position\"]; val = %#v", val)
		}
		// position
		_, ok = val["uniqueId"]
		if !ok {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"uniqueId\"]; val = %#v", val)
		}
		uniqueId, normal := val["uniqueId"].(string)
		if !normal {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"uniqueId\"]; val = %#v", val)
		}
		newStruct.UniqueId = uniqueId
		// uniqueId
		_, ok = val["yRot"]
		if !ok {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"yRot\"]; val = %#v", val)
		}
		yRot, normal := val["yRot"].(float64)
		if !normal {
			return []TargetQueryingInfo{}, fmt.Errorf("ParseTargetQueryingInfo: Crashed in val[\"yRot\"]; val = %#v", val)
		}
		newStruct.YRot = float32(yRot)
		// yRot
		res = append(res, newStruct)
		// append struct
	}
	return res, nil
	// 返回值
}
