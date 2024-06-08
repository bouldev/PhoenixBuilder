package packet

import (
	"fmt"
	"image/color"
	"math"
	"phoenixbuilder/minecraft/protocol"
)

// 一些魔法字段，
// 看起来像是用于描述是否需要继续 编码/解码 地图数据。
// 它们的作用仍是未知的，这也包括它们的数据类型，
// 因为这一切都仅仅是通过实验观察得到的结论
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
		Pixels is a list of pixel colours for the new texture of the map. It is indexed as Pixels[y][x], with
		the length of the outer slice having to be exactly Height long and the inner slices exactly Width long.

		Pixels 中的数字被表示为相应的颜色，
		这个颜色可以通过上方的 ColorMap 来查找对应，
		即“目标颜色 = ColorMap[数字]”。

		需要说明的是，被用于表示颜色的 uint32 是不确定的，
		这个数据类型仅仅是一个未经验证的推断 [需要更多信息]
	*/
	Pixels [][]uint32
	// Origin is the center position of the map being updated.
	Origin protocol.BlockPos
}

// ID ...
func (*ClientBoundMapItemData) ID() uint32 {
	return IDClientBoundMapItemData
}

// Marshal ...
func (pk *ClientBoundMapItemData) Marshal(w *protocol.Writer) {
	var magic_mark_0 uint8 = MapDataContinue
	var magic_mark_1 uint8 = 1

	w.Varint64(&pk.MapID)
	w.Varuint32(&pk.UpdateFlags)
	w.Uint8(&pk.Dimension)
	w.Bool(&pk.LockedMap)
	w.BlockPos(&pk.Origin)

	if pk.UpdateFlags&MapUpdateFlagInitialisation != 0 {
		l := uint32(len(pk.MapsIncludedIn))
		w.Varuint32(&l)
		for _, mapID := range pk.MapsIncludedIn {
			w.Varint64(&mapID)
		}
	}

	if pk.UpdateFlags&(MapUpdateFlagInitialisation|MapUpdateFlagDecoration|MapUpdateFlagTexture) != 0 {
		w.Uint8(&pk.Scale)
	}

	if pk.UpdateFlags&MapUpdateFlagDecoration != 0 {
		l := uint32(len(pk.TrackedObjects))
		w.Varuint32(&l)
		for _, obj := range pk.TrackedObjects {
			protocol.MapTrackedObj(w, &obj)
		}
		l = uint32(len(pk.TrackedObjects))
		w.Varuint32(&l)
		for _, decoration := range pk.Decorations {
			protocol.MapDeco(w, &decoration)
		}
	}

	if pk.UpdateFlags&MapUpdateFlagTexture != 0 {
		// Some basic validation for the values passed into the packet.
		if pk.Width <= 0 || pk.Height <= 0 {
			panic("invalid map texture update: width and height must be at least 1")
		}

		w.Varint32(&pk.Width)
		w.Varint32(&pk.Height)
		w.Varint32(&pk.XOffset)
		w.Varint32(&pk.YOffset)
		w.Uint8(&magic_mark_0)
		w.Uint8(&magic_mark_1)

		l := uint32(pk.Width * pk.Height)
		w.Varuint32(&l)

		if len(pk.Pixels) != int(pk.Height) {
			panic("invalid map texture update: length of outer pixels array must be equal to height")
		}
		for y := int32(0); y < pk.Height; y++ {
			if len(pk.Pixels[y]) != int(pk.Width) {
				panic("invalid map texture update: length of inner pixels array must be equal to width")
			}
			for x := int32(0); x < pk.Width; x++ {
				w.Varuint32(&pk.Pixels[y][x])
			}
		}

		l = uint32(len(pk.ColorMap))
		w.Varuint32(&l)
		for i := uint32(0); i < l; i++ {
			w.NeteaseRGBA(&pk.ColorMap[i])
			w.Varuint32(&i)
		}
	}
}

// Unmarshal ...
func (pk *ClientBoundMapItemData) Unmarshal(r *protocol.Reader) {
	var count uint32
	var magic_mark_0 uint8
	var magic_mark_1 uint8

	r.Varint64(&pk.MapID)
	r.Varuint32(&pk.UpdateFlags)
	r.Uint8(&pk.Dimension)
	r.Bool(&pk.LockedMap)
	r.BlockPos(&pk.Origin)

	if pk.UpdateFlags&MapUpdateFlagInitialisation != 0 {
		r.Varuint32(&count)
		pk.MapsIncludedIn = make([]int64, count)
		for i := uint32(0); i < count; i++ {
			r.Varint64(&pk.MapsIncludedIn[i])
		}
	}

	if pk.UpdateFlags&(MapUpdateFlagInitialisation|MapUpdateFlagDecoration|MapUpdateFlagTexture) != 0 {
		r.Uint8(&pk.Scale)
	}

	if pk.UpdateFlags&MapUpdateFlagDecoration != 0 {
		r.Varuint32(&count)
		pk.TrackedObjects = make([]protocol.MapTrackedObject, count)
		for i := uint32(0); i < count; i++ {
			protocol.MapTrackedObj(r, &pk.TrackedObjects[i])
		}
		r.Varuint32(&count)
		pk.Decorations = make([]protocol.MapDecoration, count)
		for i := uint32(0); i < count; i++ {
			protocol.MapDeco(r, &pk.Decorations[i])
		}
	}

	if pk.UpdateFlags&MapUpdateFlagTexture != 0 {
		r.Varint32(&pk.Width)
		r.Varint32(&pk.Height)
		r.Varint32(&pk.XOffset)
		r.Varint32(&pk.YOffset)
		{
			r.Uint8(&magic_mark_0)
			if magic_mark_0 == MapDataTerminate {
				return
			}
			r.Uint8(&magic_mark_1)
			if !(magic_mark_0 == MapDataContinue && magic_mark_1 == 1) {
				panic(fmt.Sprintf(
					"(pk *ClientBoundMapItemData) Unmarshal: Magic mark not matched, expect %#v but case %#v.",
					[]byte{MapDataContinue, 1},
					[]byte{magic_mark_0, magic_mark_1},
				))
			}
		}
		r.Varuint32(&count)

		r.LimitInt32(pk.Width, 0, math.MaxInt16)
		r.LimitInt32(pk.Height, 0, math.MaxInt16)
		r.LimitInt32(pk.Width*pk.Height, int32(count), int32(count))

		pk.Pixels = make([][]uint32, pk.Height)
		for y := int32(0); y < pk.Height; y++ {
			pk.Pixels[y] = make([]uint32, pk.Width)
			for x := int32(0); x < pk.Width; x++ {
				r.Varuint32(&pk.Pixels[y][x])
			}
		}

		r.Varuint32(&count)
		pk.ColorMap = make([]color.RGBA, count)
		for i := uint32(0); i < count; i++ {
			var rgba color.RGBA
			var key uint32
			r.NeteaseRGBA(&rgba)
			r.Varuint32(&key)
			pk.ColorMap[key] = rgba
		}
	}
}
