package blockNBT_Sign

import "fmt"

// 从 s.BlockEntityDatas.Block.NBT 提取告示牌数据并保存在 s.SignDatas 中
func (s *Sign) Decode() error {
	var ok bool = false
	var normal bool = false
	var textOwner string = ""
	var ignoreLighting byte = byte(0)
	var signTextColor int32 = 0
	var textIgnoreLegacyBugResolved byte = byte(0)
	var text string = ""
	// 初始化
	_, ok = s.BlockEntityDatas.Block.NBT["TextOwner"]
	if ok {
		textOwner, normal = s.BlockEntityDatas.Block.NBT["TextOwner"].(string)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.BlockEntityDatas.Block.NBT[\"TextOwner\"]; s.BlockEntityDatas.Block.NBT = %#v", s.BlockEntityDatas.Block.NBT)
		}
	}
	// TextOwner
	_, ok = s.BlockEntityDatas.Block.NBT["IgnoreLighting"]
	if ok {
		ignoreLighting, normal = s.BlockEntityDatas.Block.NBT["IgnoreLighting"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.BlockEntityDatas.Block.NBT[\"IgnoreLighting\"]; s.BlockEntityDatas.Block.NBT = %#v", s.BlockEntityDatas.Block.NBT)
		}
	}
	// IgnoreLighting
	_, ok = s.BlockEntityDatas.Block.NBT["SignTextColor"]
	if ok {
		signTextColor, normal = s.BlockEntityDatas.Block.NBT["SignTextColor"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.BlockEntityDatas.Block.NBT[\"SignTextColor\"]; s.BlockEntityDatas.Block.NBT = %#v", s.BlockEntityDatas.Block.NBT)
		}
	}
	// SignTextColor
	_, ok = s.BlockEntityDatas.Block.NBT["TextIgnoreLegacyBugResolved"]
	if ok {
		textIgnoreLegacyBugResolved, normal = s.BlockEntityDatas.Block.NBT["TextIgnoreLegacyBugResolved"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.BlockEntityDatas.Block.NBT[\"TextIgnoreLegacyBugResolved\"]; s.BlockEntityDatas.Block.NBT = %#v", s.BlockEntityDatas.Block.NBT)
		}
	}
	// TextIgnoreLegacyBugResolved
	_, ok = s.BlockEntityDatas.Block.NBT["Text"]
	if ok {
		text, normal = s.BlockEntityDatas.Block.NBT["Text"].(string)
		if !normal {
			return fmt.Errorf("Decode: Could not parse s.BlockEntityDatas.Block.NBT[\"Text\"]; s.BlockEntityDatas.Block.NBT = %#v", s.BlockEntityDatas.Block.NBT)
		}
	}
	// Text
	s.SignDatas = SignDatas{
		TextOwner:                   textOwner,
		IgnoreLighting:              ignoreLighting,
		SignTextColor:               signTextColor,
		TextIgnoreLegacyBugResolved: textIgnoreLegacyBugResolved,
		Text:                        text,
	}
	// 写入数据
	return nil
	// 返回值
}
