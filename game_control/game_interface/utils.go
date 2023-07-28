package GameInterface

import (
	"fmt"
	"phoenixbuilder/mirror/chunk"
	"strings"

	"github.com/google/uuid"
)

// 返回 uniqueID 在字符串化之后的安全形式，
// 因为我们得考虑 NEMC 的屏蔽词机制
func uuid_to_safe_string(uniqueID uuid.UUID) string {
	str := uniqueID.String()
	for key, value := range StringUUIDReplaceMap {
		str = strings.ReplaceAll(str, key, value)
	}
	return str
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
