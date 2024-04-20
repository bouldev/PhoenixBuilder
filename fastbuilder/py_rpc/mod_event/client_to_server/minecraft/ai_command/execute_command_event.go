package ai_command

import (
	"fmt"

	"github.com/google/uuid"
)

// 用于执行 魔法指令 。
// 与 请求指令 不同，
// 此操作不会消耗 魔法指令 的 使用次数 ，
// 因此在使用次数上不存在限制
type ExecuteCommandEvent struct {
	CommandLine      string    `json:"cmd"`  // e.g. "list"
	CommandRequestID uuid.UUID `json:"uuid"` // e.g. "bd903e4a-8617-4053-8321-821365777777"
}

// Return the event name of e
func (e *ExecuteCommandEvent) EventName() string {
	return "ExecuteCommandEvent"
}

// Convert e to go object which only contains go-built-in types
func (e *ExecuteCommandEvent) MakeGo() (res any) {
	return map[string]any{
		"cmd":  e.CommandLine,
		"uuid": e.CommandRequestID.String(),
	}
}

// Sync data to e from obj
func (e *ExecuteCommandEvent) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	command_line, success := object["cmd"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["cmd"] to string; object["cmd"] = %#v`, object["cmd"])
	}
	// get command line
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
	*e = ExecuteCommandEvent{
		CommandLine:      command_line,
		CommandRequestID: command_request_id,
	}
	// sync data
	return nil
	// return
}
