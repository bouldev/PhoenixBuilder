package blockNBT_Sign

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 放置一个告示牌并写入告示牌数据
func (s *Sign) WriteDatas() error {
	if s.BlockEntityDatas.Datas.FastMode {
		err := s.BlockEntityDatas.API.SetBlockFastly(s.BlockEntityDatas.Datas.Position, s.BlockEntityDatas.Block.Name, s.BlockEntityDatas.Datas.StatesString)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
	} else {
		err := s.BlockEntityDatas.API.SetBlock(s.BlockEntityDatas.Datas.Position, s.BlockEntityDatas.Block.Name, s.BlockEntityDatas.Datas.StatesString)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
	}
	// 放置告示牌
	err := s.BlockEntityDatas.API.WritePacket(&packet.BlockActorData{
		Position: s.BlockEntityDatas.Datas.Position,
		NBTData: map[string]interface{}{
			"TextOwner":                   s.SignDatas.TextOwner,
			"IgnoreLighting":              s.SignDatas.IgnoreLighting,
			"SignTextColor":               s.SignDatas.SignTextColor,
			"TextIgnoreLegacyBugResolved": s.SignDatas.TextIgnoreLegacyBugResolved,
			"Text":                        s.SignDatas.Text,
		},
	})
	if err != nil {
		return fmt.Errorf("WriteDatas: %v", err)
	}
	// 写入告示牌数据
	return nil
	// 返回值
}
