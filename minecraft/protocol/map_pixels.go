/*
PhoenixBuilder specific.
Author: Happy2018new, Liliya233
*/
package protocol

import "image/color"

const (
	MapPixelsTypeNeteaseUint8 = iota + 1
	MapPixelsTypeNeteaseUint16
	MapPixelsTypeStandard
)

// PixelsData represents an object that holds data specific to a map pixels type.
// The data it holds depends on the type.
type MapPixelsData interface {
	// Marshal encodes/decodes a map pixels object.
	Marshal(r IO)
}

// lookupMapPixels looks up map pixels data for the ID passed.
func lookupMapPixels(id uint8, x *MapPixelsData) bool {
	switch id {
	case MapPixelsTypeNeteaseUint8:
		*x = &Uint8Pixels{}
	case MapPixelsTypeNeteaseUint16:
		*x = &Uint16Pixels{}
	case MapPixelsTypeStandard:
		*x = &StandardPixels{}
	default:
		return false
	}
	return true
}

// lookupMapPixelsType looks up an ID for a specific map pixels data.
func lookupMapPixelsType(x MapPixelsData, id *uint8) bool {
	switch x.(type) {
	case *Uint8Pixels:
		*id = MapPixelsTypeNeteaseUint8
	case *Uint16Pixels:
		*id = MapPixelsTypeNeteaseUint16
	case *StandardPixels:
		*id = MapPixelsTypeStandard
	default:
		return false
	}
	return true
}

// 描述一个颜色，
// 但其索引的数据类型为 uint8 。
// 被下方的 Uint8Pixels 所使用
type Uint8Color struct {
	Colour color.RGBA // 该颜色的 RGBA 值
	Index  uint8      // 该颜色对应的编号(索引值)
}

// Marshal ...
func (x *Uint8Color) Marshal(r IO) {
	r.RGBA(&x.Colour)
	r.Uint8(&x.Index)
}

// 描述颜色索引数据类型为 uint8 的地图画上的像素集
type Uint8Pixels struct {
	// 一维的地图画像素集合。
	//
	// 该集合中的数字被表示为相应的颜色，
	// 这个颜色可以通过下方的 ColorMap 来查找对应，
	// 即“目标颜色 = ColorMap[数字]”
	Pixels []uint8
	/*
		一个切片，但实际作用是映射，
		用于表示数字和颜色的对应关系，它并非是恒定不变的。

		例如，用切片的第 0 项和第 1 项分别代表 黄色 和 黑色，
		或在下次使用切片的 0 项和第 1 项分别代表 绿色 和 橙色。

		这个映射表被用于上方的 Pixels 二维数字矩阵，
		这个矩阵中的每个数字就代表一个颜色。因此，
		我们通过使用 ColorMap 来描述 Pixels 中数字和颜色的对应关系。
	*/
	ColorMap []Uint8Color
}

// ...
func (x *Uint8Pixels) Marshal(r IO) {
	FuncSlice(r, &x.Pixels, r.Uint8)
	Slice(r, &x.ColorMap)
}

// 描述一个颜色，
// 但其索引的数据类型为 uint16 。
// 被下方的 Uint16Pixels 所使用
type Uint16ColorMap struct {
	Colour color.RGBA // 该颜色的 RGBA 值
	Index  uint16     // 该颜色对应的编号(索引值)
}

// ...
func (x *Uint16ColorMap) Marshal(r IO) {
	r.RGBA(&x.Colour)
	r.Uint16(&x.Index)
}

// 描述颜色索引数据类型为 uint16 的地图画上的像素集
type Uint16Pixels struct {
	// 一维的地图画像素集合。
	//
	// 该集合中的数字被表示为相应的颜色，
	// 这个颜色可以通过下方的 ColorMap 来查找对应，
	// 即“目标颜色 = ColorMap[数字]”
	Pixels []uint16
	/*
		一个切片，但实际作用是映射，
		用于表示数字和颜色的对应关系，它并非是恒定不变的。

		例如，用切片的第 0 项和第 1 项分别代表 黄色 和 黑色，
		或在下次使用切片的 0 项和第 1 项分别代表 绿色 和 橙色。

		这个映射表被用于上方的 Pixels 二维数字矩阵，
		这个矩阵中的每个数字就代表一个颜色。因此，
		我们通过使用 ColorMap 来描述 Pixels 中数字和颜色的对应关系。
	*/
	ColorMap []Uint16ColorMap
}

// ...
func (x *Uint16Pixels) Marshal(r IO) {
	FuncSlice(r, &x.Pixels, r.Uint16)
	Slice(r, &x.ColorMap)
}

// 描述国际版数据包传输协议下，
// 地图画上的像素集
type StandardPixels struct {
	Pixels []color.RGBA // 一维的地图画像素集合
}

// ...
func (x *StandardPixels) Marshal(r IO) {
	FuncSlice(r, &x.Pixels, r.RGBA)
}
