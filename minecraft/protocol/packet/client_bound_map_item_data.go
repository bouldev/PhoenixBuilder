package packet

import (
	"fmt"
	"image/color"
	"phoenixbuilder/minecraft/protocol"
)

/*
PhoenixBuilder specific constants.
Author: Happy2018new

一些魔法字段，
看起来像是用于描述是否需要继续 编码/解码 地图数据。
它们的作用仍是未知的，这也包括它们的数据类型，
因为这一切都仅仅是通过实验观察得到的结论
*/
const (
	MapDataContinue = uint8(iota)
	MapDataTerminate
)

const (
	MapUpdateFlagTexture = 1 << (iota + 1)
	MapUpdateFlagDecoration
	MapUpdateFlagInitialisation
)

// ClientBoundMapItemData is sent by the server to the client to update the data of a map shown to the client.
// It is sent with a combination of flags that specify what data is updated.
// The ClientBoundMapItemData packet may be used to update specific parts of the map only. It is not required
// to send the entire map each time when updating one part.
type ClientBoundMapItemData struct {
	// MapID is the unique identifier that represents the map that is updated over network. It remains
	// consistent across sessions.
	MapID int64
	// UpdateFlags is a combination of flags found above that indicate what parts of the map should be updated
	// client-side.
	UpdateFlags uint32
	// Dimension is the dimension of the map that should be updated, for example the overworld (0), the nether
	// (1) or the end (2).
	Dimension byte
	// LockedMap specifies if the map that was updated was a locked map, which may be done using a cartography
	// table.
	LockedMap bool
	// Origin is the center position of the map being updated.
	Origin protocol.BlockPos
	// Scale is the scale of the map as it is shown in-game. It is written when any of the MapUpdateFlags are
	// set to the UpdateFlags field.
	Scale byte

	// The following fields apply only for the MapUpdateFlagInitialisation.

	// MapsIncludedIn holds an array of map IDs that the map updated is included in. This has to do with the
	// scale of the map: Each map holds its own map ID and all map IDs of maps that include this map and have
	// a bigger scale. This means that a scale 0 map will have 5 map IDs in this slice, whereas a scale 4 map
	// will have only 1 (its own).
	// The actual use of this field remains unknown.
	MapsIncludedIn []int64

	// The following fields apply only for the MapUpdateFlagDecoration.

	// TrackedObjects is a list of tracked objects on the map, which may either be entities or blocks. The
	// client makes sure these tracked objects are actually tracked. (position updated etc.)
	TrackedObjects []protocol.MapTrackedObject
	// Decorations is a list of fixed decorations located on the map. The decorations will not change
	// client-side, unless the server updates them.
	Decorations []protocol.MapDecoration

	// The following fields apply only for the MapUpdateFlagTexture update flag.

	// Height is the height of the texture area that was updated. The height may be a subset of the total
	// height of the map.
	Height int32
	// Width is the width of the texture area that was updated. The width may be a subset of the total width
	// of the map.
	Width int32
	// XOffset is the X offset in pixels at which the updated texture area starts. From this X, the updated
	// texture will extend exactly Width pixels to the right.
	XOffset int32
	// YOffset is the Y offset in pixels at which the updated texture area starts. From this Y, the updated
	// texture will extend exactly Height pixels up.
	YOffset int32
	/*
		PhoenixBuilder specific fields.
		Author: Happy2018new



		一个切片，但实际作用是映射，
		用于表示数字和颜色的对应关系，它并非是恒定不变的。

		例如，用切片的第 0 项和第 1 项分别代表 黄色 和 黑色，
		或在下次使用切片的 0 项和第 1 项分别代表 绿色 和 橙色。

		这个映射表被用于下方的 Pixels 二维数字矩阵，
		这个矩阵中的每个数字就代表一个颜色。因此，
		我们通过使用 ColorMap 来描述 Pixels 中数字和颜色的对应关系。
	*/
	ColorMap []color.RGBA
	/*
		PhoenixBuilder specific fields, which modified from orgin version.
		Author: Happy2018new



		Pixels is a list of pixel colours for the new texture of the map. It is indexed as Pixels[y*height + x].

		Pixels 中的数字被表示为相应的颜色，
		这个颜色可以通过上方的 ColorMap 来查找对应，
		即“目标颜色 = ColorMap[数字]”。

		需要说明的是，被用于表示颜色的 uint32 是不确定的，
		这个数据类型仅仅是一个未经验证的推断 [需要更多信息]
	*/
	Pixels []uint32
}

// ID ...
func (*ClientBoundMapItemData) ID() uint32 {
	return IDClientBoundMapItemData
}

func (pk *ClientBoundMapItemData) Marshal(io protocol.IO) {
	// PhoenixBuilder specific changes.
	// Author: Happy2018new
	var magic_mark_0 uint8 = MapDataContinue
	var magic_mark_1 uint8 = 1
	var length uint32 = uint32(len(pk.ColorMap))

	io.Varint64(&pk.MapID)
	io.Varuint32(&pk.UpdateFlags)
	io.Uint8(&pk.Dimension)
	io.Bool(&pk.LockedMap)
	io.BlockPos(&pk.Origin)

	if pk.UpdateFlags&MapUpdateFlagInitialisation != 0 {
		protocol.FuncSlice(io, &pk.MapsIncludedIn, io.Varint64)
	}

	if pk.UpdateFlags&(MapUpdateFlagInitialisation|MapUpdateFlagDecoration|MapUpdateFlagTexture) != 0 {
		io.Uint8(&pk.Scale)
	}

	if pk.UpdateFlags&MapUpdateFlagDecoration != 0 {
		protocol.Slice(io, &pk.TrackedObjects)
		protocol.Slice(io, &pk.Decorations)
	}

	if pk.UpdateFlags&MapUpdateFlagTexture != 0 {
		io.Varint32(&pk.Width)
		io.Varint32(&pk.Height)
		io.Varint32(&pk.XOffset)
		io.Varint32(&pk.YOffset)

		// PhoenixBuilder specific changes.
		// Author: Happy2018new
		{
			// protocol.FuncSlice(io, &pk.Pixels, io.VarRGBA)

			io.Uint8(&magic_mark_0)
			if magic_mark_0 != MapDataContinue {
				return
			}

			io.Uint8(&magic_mark_1)
			if magic_mark_1 != 1 {
				panic(fmt.Sprintf(
					"(pk *ClientBoundMapItemData) Marshal: Magic mark not matched, expect %#v but case %#v.",
					[]byte{1}, []byte{magic_mark_0},
				))
			}

			protocol.FuncSlice(io, &pk.Pixels, io.Varuint32)

			io.Varuint32(&length)
			if pk.ColorMap == nil {
				pk.ColorMap = make([]color.RGBA, length)
			}
			for i := uint32(0); i < length; i++ {
				var rgba color.RGBA = pk.ColorMap[i]
				var key uint32 = i
				io.NeteaseRGBA(&rgba)
				io.Varuint32(&key)
				pk.ColorMap[key] = rgba
			}
		}
	}
}
