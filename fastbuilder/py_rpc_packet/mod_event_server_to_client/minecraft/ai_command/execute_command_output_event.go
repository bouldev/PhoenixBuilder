package ai_command

import (
	"fmt"

	"github.com/google/uuid"
)

// 对 魔法指令 的响应体，
// 其作用可能是用于描述其命令输出。
// 当且仅当 魔法命令 成功时收到
type ExecuteCommandOutputEvent struct {
	AIModel          string    `json:"ai_model"`      // e.g. "-1"
	CommandLine      string    `json:"cmd"`           // e.g. "list"
	GenerateType     int64     `json:"generate_type"` // e.g. int64(3)
	IsReExecute      bool      `json:"isReExecute"`   // e.g. False
	IsLast           bool      `json:"is_last"`       // e.g. False
	CommandOutput    string    `json:"msg"`           // e.g. "命令输出：目前有 1/10 个玩家在线："
	CommandRequestID uuid.UUID `json:"uuid"`          // e.g. "bd903e4a-8617-4053-8321-821365777777"
}

// Return the event name of e
func (e *ExecuteCommandOutputEvent) EventName() string {
	return "ExecuteCommandOutputEvent"
}

// Convert e to go object which only contains go-built-in types
func (e *ExecuteCommandOutputEvent) MakeGo() (res any) {
	return map[string]any{
		"ai_model":      e.AIModel,
		"cmd":           e.CommandLine,
		"generate_type": e.GenerateType,
		"isReExecute":   e.IsReExecute,
		"is_last":       e.IsLast,
		"msg":           e.CommandOutput,
		"uuid":          e.CommandRequestID.String(),
	}
}

// Sync data to e from obj
func (e *ExecuteCommandOutputEvent) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	ai_model, success := object["ai_model"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["ai_model"] to string; object["ai_model"] = %#v`, object["ai_model"])
	}
	command_line, success := object["cmd"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["cmd"] to string; object["cmd"] = %#v`, object["cmd"])
	}
	generate_type, success := object["generate_type"].(int64)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["generate_type"] to int64; object["generate_type"] = %#v`, object["generate_type"])
	}
	is_re_execute, success := object["isReExecute"].(bool)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["isReExecute"] to bool; object["isReExecute"] = %#v`, object["isReExecute"])
	}
	is_last, success := object["is_last"].(bool)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["is_last"] to bool; object["is_last"] = %#v`, object["is_last"])
	}
	command_output, success := object["msg"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["msg"] to string; object["msg"] = %#v`, object["msg"])
	}
	// get data
	command_request_id_origin, success := object["uuid"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["uuid"] to string; object["uuid"] = %#v`, object["uuid"])
	}
	command_request_id := uuid.UUID{}
	err := command_request_id.UnmarshalText([]byte(command_request_id_origin))
	if err != nil {
		return fmt.Errorf("FromGo: %v", err)
	}
	// get command request id
	*e = ExecuteCommandOutputEvent{
		AIModel:          ai_model,
		CommandLine:      command_line,
		GenerateType:     generate_type,
		IsReExecute:      is_re_execute,
		IsLast:           is_last,
		CommandOutput:    command_output,
		CommandRequestID: command_request_id,
	}
	// sync data
	return nil
	// return
}
