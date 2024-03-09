package ai_command

import (
	"fmt"
)

// 魔法指令 对命令请求在前置检查失败时的响应体，
// 通常在 作弊未开启 或 机器人不具备操作员权限 时触发
type AvailableCheckFailed struct {
	Reason string `json:"reason"` // e.g. "执行失败：您没有操作员权限"
}

// Return the event name of a
func (a *AvailableCheckFailed) EventName() string {
	return "AvailableCheckFailed"
}

// Convert a to go object which only contains go-built-in types
func (a *AvailableCheckFailed) MakeGo() (res any) {
	return map[string]any{"reason": a.Reason}
}

// Sync data to a from obj
func (a *AvailableCheckFailed) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	reason, success := object["reason"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["reason"] to string; object["reason"] = %#v`, object["reason"])
	}
	// get reason
	*a = AvailableCheckFailed{Reason: reason}
	// sync data
	return nil
	// return
}
