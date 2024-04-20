package ai_command

import (
	"fmt"

	"github.com/google/uuid"
)

// 魔法指令 对命令请求的响应体，
// 用于描述命令是否执行成功
type AfterExecuteCommandEvent struct {
	Success          bool      `json:"executeResult"` // e.g. bool
	CommandRequestID uuid.UUID `json:"uuid"`          // e.g. "bd903e4a-8617-4053-8321-821365777777"
}

// Return the event name of a
func (a *AfterExecuteCommandEvent) EventName() string {
	return "AfterExecuteCommandEvent"
}

// Convert a to go object which only contains go-built-in types
func (a *AfterExecuteCommandEvent) MakeGo() (res any) {
	return map[string]any{
		"executeResult": a.Success,
		"uuid":          a.CommandRequestID.String(),
	}
}

// Sync data to a from obj
func (a *AfterExecuteCommandEvent) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	execute_result, success := object["executeResult"].(bool)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["executeResult"] to bool; object["executeResult"] = %#v`, object["executeResult"])
	}
	// get execute result
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
	*a = AfterExecuteCommandEvent{
		Success:          execute_result,
		CommandRequestID: command_request_id,
	}
	// sync data
	return nil
	// return
}
