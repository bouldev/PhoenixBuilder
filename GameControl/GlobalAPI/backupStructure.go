package GlobalAPI

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
func (g *GlobalAPI) BackupStructure(structure MCStructure) (uuid.UUID, error) {
	uniqueId := generateUUID()
	// get new uuid
	resp, err := g.SendWSCommandWithResponce(
		fmt.Sprintf(
			`structure save "%v" %d %d %d %d %d %d`,
			uniqueId.String(),
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
func (g *GlobalAPI) RevertStructure(uniqueID uuid.UUID, pos BlockPos) error {
	{
		resp, err := g.SendWSCommandWithResponce(
			fmt.Sprintf(
				`structure load "%v" %d %d %d`,
				uniqueID.String(),
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
	}
	// revert structure
	{
		err := g.SendSettingsCommand(
			fmt.Sprintf(
				`structure delete "%v"`,
				uniqueID.String(),
			),
			false,
		)
		if err != nil {
			return fmt.Errorf("RevertStructure: %v", err)
		}
	}
	// delete structure
	return nil
	// return
}
