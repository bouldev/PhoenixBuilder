package GameInterface

// TODO: 在某天支持 结构空位 的备份和恢复

import (
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"
	ResourcesControl "phoenixbuilder/game_control/resources_control"

	"github.com/google/uuid"
)

// 描述一个结构的起点坐标及尺寸
type MCStructure mcstructure.Area

// 描述一个单个方块的位置，这被用于恢复结构的实现
type BlockPos mcstructure.BlockPos

// 备份 structure 所指代的区域为结构。
// 返回一个 uuid.UUID 对象，
// 其 uuid_to_safe_string(uuid.UUID) 形式代表被备份结构的名称
func (g *GameInterface) BackupStructure(structure MCStructure) (uuid.UUID, error) {
	uniqueId := ResourcesControl.GenerateUUID()
	// get new uuid
	request := fmt.Sprintf(
		`structure save "%s" %d %d %d %d %d %d`,
		uuid_to_safe_string(uniqueId),
		structure.BeginX,
		structure.BeginY,
		structure.BeginZ,
		structure.BeginX+structure.SizeX-1,
		structure.BeginY+structure.SizeY-1,
		structure.BeginZ+structure.SizeZ-1,
	)
	// get command to backup structure
	resp := g.SendWSCommandWithResponse(
		request,
		ResourcesControl.CommandRequestOptions{
			TimeOut: ResourcesControl.CommandRequestDefaultDeadLine,
		},
	)
	if resp.Error != nil && resp.ErrorType != ResourcesControl.ErrCommandRequestTimeOut {
		return uuid.UUID{}, fmt.Errorf("BackupStructure: Failed to backup the structure; structure = %#v", structure)
	}
	if resp.Error == nil && resp.Respond.SuccessCount <= 0 {
		return uuid.UUID{}, fmt.Errorf("BackupStructure: Failed to backup the structure; structure = %#v; resp = %#v", structure, resp)
	}
	// send command and check success states
	if resp.Error != nil && resp.ErrorType == ResourcesControl.ErrCommandRequestTimeOut {
		err := g.SendSettingsCommand(request, true)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("BackupStructure: Failed to backup the structure; structure = %#v, err = %v", structure, err)
		}
		err = g.AwaitChangesGeneral()
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("BackupStructure: Failed to backup the structure; structure = %#v; err = %#v", structure, err)
		}
	}
	// some special solutions for when we facing Netease Mask Words System
	return uniqueId, nil
	// return
}

// 删除名称为 uuid_to_safe_string(uuid.UUID) 的结构
func (g *GameInterface) DeleteStructure(uniqueID uuid.UUID) error {
	err := g.SendSettingsCommand(
		fmt.Sprintf(
			`structure delete "%v"`,
			uuid_to_safe_string(uniqueID),
		),
		false,
	)
	if err != nil {
		return fmt.Errorf("DeleteStructure: %v", err)
	}
	return nil
}

// 在 pos 处恢复名称为 uuid_to_safe_string(uuid.UUID) 的备份用结构，
// 然后删除此结构
func (g *GameInterface) RevertStructure(uniqueID uuid.UUID, pos BlockPos) error {
	defer func() {
		g.DeleteStructure(uniqueID)
	}()
	// delete structure
	request := fmt.Sprintf(
		`structure load "%v" %d %d %d`,
		uuid_to_safe_string(uniqueID),
		pos[0],
		pos[1],
		pos[2],
	)
	// get command to revert the structure
	resp := g.SendWSCommandWithResponse(
		request,
		ResourcesControl.CommandRequestOptions{
			TimeOut: ResourcesControl.CommandRequestDefaultDeadLine,
		},
	)
	if resp.Error != nil && resp.ErrorType != ResourcesControl.ErrCommandRequestTimeOut {
		return fmt.Errorf(`RevertStructure: Failed to revert structure named "%v"; pos = %#v`, uniqueID.String(), pos)
	}
	if resp.Error == nil && resp.Respond.SuccessCount <= 0 {
		return fmt.Errorf(`RevertStructure: Failed to revert structure named "%v"; pos = %#v`, uniqueID.String(), pos)
	}
	// send command and check sucess states
	if resp.Error != nil && resp.ErrorType == ResourcesControl.ErrCommandRequestTimeOut {
		err := g.SendSettingsCommand(request, true)
		if err != nil {
			return fmt.Errorf(`RevertStructure: Failed to revert structure named "%v"; pos = %#v, err = %v`, uniqueID.String(), pos, err)
		}
		err = g.AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf(`RevertStructure: Failed to revert structure named "%v"; pos = %#v, err = %#v`, uniqueID.String(), pos, err)
		}
	}
	// some special solutions for when we facing Netease Mask Words System
	return nil
	// return
}
