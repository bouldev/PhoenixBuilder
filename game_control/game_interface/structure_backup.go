package GameInterface

// TODO: 在某天支持 结构空位 的备份和恢复

import (
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"

	"github.com/google/uuid"
)

// 描述一个结构的起点坐标及尺寸
type MCStructure mcstructure.Area

// 描述一个单个方块的位置，这被用于恢复结构的实现
type BlockPos mcstructure.BlockPos

// 备份 structure 所指代的区域为结构。
// 返回一个 uuid.UUID 对象，其字符串形式代表被备份结构的名称
func (g *GameInterface) BackupStructure(structure MCStructure) (uuid.UUID, error) {
	uniqueId := generateUUID()
	// get new uuid
	resp, err := g.SendWSCommandWithResponse(
		fmt.Sprintf(
			`structure save "%s" %d %d %d %d %d %d`,
			uuid_to_safe_string(uniqueId),
			structure.BeginX,
			structure.BeginY,
			structure.BeginZ,
			structure.BeginX+structure.SizeX-1,
			structure.BeginY+structure.SizeY-1,
			structure.BeginZ+structure.SizeZ-1,
		),
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("BackupStructure: Failed to backup the structure; structure = %#v", structure)
	}
	// backup structure
	if resp.SuccessCount <= 0 {
		return uuid.UUID{}, fmt.Errorf("BackupStructure: Failed to backup the structure; structure = %#v; resp = %#v", structure, resp)
	}
	// check success states
	return uniqueId, nil
	// return
}

// 在 pos 处恢复名称为 unique.String() 的备份用结构并删除此结构
func (g *GameInterface) RevertStructure(uniqueID uuid.UUID, pos BlockPos) error {
	resp, err := g.SendWSCommandWithResponse(
		fmt.Sprintf(
			`structure load "%v" %d %d %d`,
			uuid_to_safe_string(uniqueID),
			pos[0],
			pos[1],
			pos[2],
		),
	)
	if err != nil {
		return fmt.Errorf(`RevertStructure: Failed to revert structure named "%v"; pos = %#v`, uniqueID.String(), pos)
	}
	if resp.SuccessCount <= 0 {
		return fmt.Errorf(`RevertStructure: Failed to revert structure named "%v"; pos = %#v`, uniqueID.String(), pos)
	}
	// revert structure
	err = g.SendSettingsCommand(
		fmt.Sprintf(
			`structure delete "%v"`,
			uuid_to_safe_string(uniqueID),
		),
		false,
	)
	if err != nil {
		return fmt.Errorf("RevertStructure: %v", err)
	}
	// delete structure
	return nil
	// return
}
