package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/mirror/chunk"

	"github.com/google/uuid"
)

// 生成一个新的 uuid 对象并返回
func generateUUID() uuid.UUID {
	for {
		uniqueId, err := uuid.NewUUID()
		if err != nil {
			continue
		}
		return uniqueId
	}
}

// 取得名称为 name 且方块状态为 states 的方块在 NEMC 下的 Block Runtime ID 。
// 特别地，name 需要加上命名空间 minecraft
func blockStatesToNEMCRuntimeID(
	name string,
	states map[string]interface{},
) (uint32, error) {
	standardRuntimeID, found := chunk.StateToRuntimeID(name, states)
	if !found {
		return 0, fmt.Errorf("blockStatesToNEMCRuntimeID: Failed to get the runtimeID of block %v; states = %#v", name, states)
	}
	neteaseBlockRuntimeID := chunk.StandardRuntimeIDToNEMCRuntimeID(standardRuntimeID)
	if neteaseBlockRuntimeID == chunk.AirRID || neteaseBlockRuntimeID == chunk.NEMCAirRID {
		return 0, fmt.Errorf("blockStatesToNEMCRuntimeID: Failed to converse StandardRuntimeID to NEMCRuntimeID; standardRuntimeID = %#v, name = %#v, states = %#v", standardRuntimeID, name, states)
	}
	return neteaseBlockRuntimeID, nil
}
