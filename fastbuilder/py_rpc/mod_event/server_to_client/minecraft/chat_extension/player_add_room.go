package chat_extension

import "fmt"

type RuntimeID string      // 描述玩家的 运行时 ID
type NeteaseUserUID string // 描述玩家的 网易我的世界账户 对应的 UID

// 其他玩家加入游戏时触发
type PlayerAddRoom struct {
	// e.g. "-455266532518": int64(0)
	DimID map[RuntimeID]int64 `json:"id2DimId"`
	// e.g. "-455266532518": "2237342498"
	UIDByRuntimeID map[RuntimeID]NeteaseUserUID `json:"id2Uid"`
	// e.g. "-455266532518": {}
	PrefixInfo map[RuntimeID]map[string]any `json:"prefixInfo"`
	// e.g. ["2237342498"]
	UID []NeteaseUserUID `json:"uids"`
}

// Return the event name of p
func (p *PlayerAddRoom) EventName() string {
	return "PlayerAddRoom"
}

// Convert p to go object which only contains go-built-in types
func (p *PlayerAddRoom) MakeGo() (res any) {
	dim_id := map[string]any{}
	uid_by_runtime_id := map[string]any{}
	prefix_info := map[string]any{}
	uid := make([]string, len(p.UID))
	// prepare
	for key, value := range p.DimID {
		dim_id[string(key)] = value
	}
	for key, value := range p.UIDByRuntimeID {
		uid_by_runtime_id[string(key)] = string(value)
	}
	for key, value := range p.PrefixInfo {
		prefix_info[string(key)] = value
	}
	for key, value := range p.UID {
		uid[key] = string(value)
	}
	// convert data
	return map[string]any{
		"id2DimId":   dim_id,
		"id2Uid":     uid_by_runtime_id,
		"prefixInfo": prefix_info,
		"uids":       uid,
	}
	// return
}

// Sync data to p from obj
func (p *PlayerAddRoom) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	dim_id, success := object["id2DimId"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["id2DimId"] to map[string]interface{}; object["id2DimId"] = %#v`, object["id2DimId"])
	}
	uid_by_runtime_id, success := object["id2Uid"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["id2Uid"] to map[string]interface{}; object["id2Uid"] = %#v`, object["id2Uid"])
	}
	prefix_info, success := object["prefixInfo"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["prefix_info"] to map[string]interface{}; object["prefix_info"] = %#v`, object["prefix_info"])
	}
	uid, success := object["uids"].([]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["uids"] to []interface{}; object["uids"] = %#v`, object["uids"])
	}
	// get data
	p.DimID = make(map[RuntimeID]int64)
	p.UIDByRuntimeID = make(map[RuntimeID]NeteaseUserUID)
	p.PrefixInfo = make(map[RuntimeID]map[string]any)
	p.UID = make([]NeteaseUserUID, len(uid))
	// make map and slice
	for key, value := range dim_id {
		val, success := value.(int64)
		if !success {
			return fmt.Errorf(`FromGo: Failed to convert dim_id[%#v] to int64; dim_id[%#v] = %#v`, key, key, value)
		}
		p.DimID[RuntimeID(key)] = val
	}
	for key, value := range uid_by_runtime_id {
		val, success := value.(string)
		if !success {
			return fmt.Errorf(`FromGo: Failed to convert uid_by_runtime_id[%#v] to string; uid_by_runtime_id[%#v] = %#v`, key, key, value)
		}
		p.UIDByRuntimeID[RuntimeID(key)] = NeteaseUserUID(val)
	}
	for key, value := range prefix_info {
		val, success := value.(map[string]any)
		if !success {
			return fmt.Errorf(`FromGo: Failed to convert prefix_info[%#v] to map[string]interface{}; prefix_info[%#v] = %#v`, key, key, value)
		}
		p.PrefixInfo[RuntimeID(key)] = val
	}
	for key, value := range uid {
		val, success := value.(string)
		if !success {
			return fmt.Errorf(`FromGo: Failed to convert uid[%d] to string; uid[%d] = %#v`, key, key, value)
		}
		p.UID[key] = NeteaseUserUID(val)
	}
	// sync data
	return nil
	// return
}
