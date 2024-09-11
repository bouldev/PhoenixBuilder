/*
PhoenixBuilder specific packages.
Author: Liliya233, Happy2018new
*/
package NetEaseSkin

// ...
type SkinManifest struct {
	Header struct {
		UUID string `json:"uuid"`
	} `json:"header"`
}

// 描述皮肤信息
type Skin struct {
	// 储存皮肤数据的二进制负载。
	// 对于普通皮肤，这是一个二进制形式的 PNG；
	// 对于高级皮肤(如 4D 皮肤)，
	// 这是一个压缩包形式的二进制表示
	FullSkinData []byte
	// 皮肤的 UUID
	SkinUUID string
	// 皮肤项目的 UUID
	SkinItemID string
	// 皮肤的手臂是否纤细
	SkinIsSlim bool
	// 皮肤的一维密集像素矩阵
	SkinPixels []byte
	// 皮肤的骨架信息
	SkinGeometry []byte
	// SkinResourcePatch 是一个 JSON 编码对象，
	// 其中包含一些指向皮肤所具有的几何形状的字段。
	// 它包含的 JSON 对象指定动画的几何形状，
	// 以及播放器的默认皮肤的组合方式
	SkinResourcePatch []byte
	// 皮肤的宽度
	SkinWidth int
	// 皮肤的高度
	SkinHight int
}
