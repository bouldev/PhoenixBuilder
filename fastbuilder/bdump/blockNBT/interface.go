package blockNBT

// 用于放置方块实体的通用接口
type GeneralBlockNBT interface {
	// 解析 NBT(map[string]interface{}) 为 golang struct
	Decode() error
	// 放置方块并以合法手段写入方块实体数据
	WriteDatas() error
}
