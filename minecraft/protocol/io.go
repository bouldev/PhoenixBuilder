package protocol

import (
	"fmt"
	"image/color"
	"phoenixbuilder/minecraft/nbt"
	"reflect"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

// IO represents a packet IO direction. Implementations of this interface are Reader and Writer. Reader reads
// data from the input stream into the pointers passed, whereas Writer writes the values the pointers point to
// to the output stream.
type IO interface {
	Uint16(x *uint16)
	Int16(x *int16)
	Uint32(x *uint32)
	Int32(x *int32)
	BEInt32(x *int32)
	Uint64(x *uint64)
	Int64(x *int64)
	Float32(x *float32)
	Uint8(x *uint8)
	Int8(x *int8)
	Bool(x *bool)
	Varint64(x *int64)
	Varuint64(x *uint64)
	Varint32(x *int32)
	Varuint32(x *uint32)

	// PhoenixBuilder specific changes.
	// Author: Happy2018new, Liliya233
	Varint16(x *int16)
	Varuint16(x *uint16)

	String(x *string)
	StringUTF(x *string)
	ByteSlice(x *[]byte)
	Vec3(x *mgl32.Vec3)
	Vec2(x *mgl32.Vec2)
	BlockPos(x *BlockPos)
	UBlockPos(x *BlockPos)
	ChunkPos(x *ChunkPos)
	SubChunkPos(x *SubChunkPos)
	SoundPos(x *mgl32.Vec3)
	ByteFloat(x *float32)
	Bytes(p *[]byte)

	// PhoenixBuilder specific changes.
	// Author: Happy2018new, Liliya233
	PistonAttachedBlocks(m *[]int32)

	NBT(m *map[string]any, encoding nbt.Encoding)

	// PhoenixBuilder specific changes.
	// Author: Happy2018new, Liliya233
	NBTWithLength(m *map[string]any)

	NBTList(m *[]any, encoding nbt.Encoding)

	// PhoenixBuilder specific changes.
	// Author: Happy2018new, Liliya233
	EnchantList(x *[]Enchant)
	NBTItem(m *Item)
	ItemList(x *[]ItemWithSlot)

	UUID(x *uuid.UUID)
	RGBA(x *color.RGBA)
	VarRGBA(x *color.RGBA)
	EntityMetadata(x *map[uint32]any)
	Item(x *ItemStack)
	ItemInstance(i *ItemInstance)
	ItemDescriptorCount(i *ItemDescriptorCount)
	StackRequestAction(x *StackRequestAction)
	MaterialReducer(x *MaterialReducer)
	Recipe(x *Recipe)
	EventType(x *Event)
	TransactionDataType(x *InventoryTransactionData)

	// PhoenixBuilder specific changes.
	// Author: Happy2018new
	MapPixelsDataType(x *MapPixelsData)

	PlayerInventoryAction(x *UseItemTransactionData)
	GameRule(x *GameRule)
	AbilityValue(x *any)
	CompressedBiomeDefinitions(x *map[string]any)

	ShieldID() int32
	UnknownEnumOption(value any, enum string)
	InvalidValue(value any, forField, reason string)

	/*
		PhoenixBuilder specific func.
		Changes Maker: Liliya233
		Committed by Happy2018new.
	*/
	USubChunkPos(x *SubChunkPos)
	USoundPos(x *mgl32.Vec3)

	/*
		PhoenixBuilder specific func.
		Author: Liliya233, CMA2041PT, Happy2018new

		Netease's Python MsgPack
	*/
	MsgPack(x *any)
}

// Marshaler is a type that can be written to or read from an IO.
type Marshaler interface {
	Marshal(r IO)
}

// PhoenixBuilder specific interface.
// Author: Happy2018new
//
// Varint 描述了 __tag NBT 在网络传输时整数的数据类型
type Int interface {
	uint16 | uint32 | uint64 | int16 | int32 | int64
}

/*
PhoenixBuilder specific interface.
Author: Happy2018new

TAGNumber 描述了标准 NBT 中允许的整数类型。
uint8 被指代 TAG_Byte(1)，
uint16 被指代 TAG_Short(3)，
int32 被指代 TAG_Int(4)，
int64 被指代 TAG_Long(5)
*/
type TAGNumber interface {
	uint8 | int16 | int32 | int64
}

// Slice reads/writes a slice of T with a varuint32 prefix.
func Slice[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	SliceOfLen[T, S, A](r, count, x)
}

// PhoenixBuilder specific changes.
// Author: Happy2018new, Liliya233
//
// SliceVarint16Length reads/writes a slice of T with a varint16 prefix.
func SliceVarint16Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := int16(len(*x))
	r.Varint16(&count)
	SliceOfLen[T, S, A](r, uint32(count), x)
}

// SliceUint8Length reads/writes a slice of T with a uint8 prefix.
func SliceUint8Length[T any, S *[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint8(len(*x))
	r.Uint8(&count)
	SliceOfLen[T, S, A](r, uint32(count), x)
}

// SliceUint16Length reads/writes a slice of T with a uint16 prefix.
func SliceUint16Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint16(len(*x))
	r.Uint16(&count)
	SliceOfLen[T, S, A](r, uint32(count), x)
}

// SliceUint32Length reads/writes a slice of T with a uint32 prefix.
func SliceUint32Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint32(len(*x))
	r.Uint32(&count)
	SliceOfLen[T, S, A](r, count, x)
}

// SliceVarint32Length reads/writes a slice of T with a varint32 prefix.
func SliceVarint32Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := int32(len(*x))
	r.Varint32(&count)
	SliceOfLen[T, S, A](r, uint32(count), x)
}

// PhoenixBuilder specific func.
// Author: Liliya233
//
// SliceVaruint32Length reads/writes a slice of T with a varuint32 prefix.
func SliceVaruint32Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	SliceOfLen[T, S, A](r, count, x)
}

// FuncSliceUint16Length reads/writes a slice of T using function f with a uint16 length prefix.
func FuncSliceUint16Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint16(len(*x))
	r.Uint16(&count)
	FuncSliceOfLen(r, uint32(count), x, f)
}

// FuncSliceUint32Length reads/writes a slice of T using function f with a uint32 length prefix.
func FuncSliceUint32Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint32(len(*x))
	r.Uint32(&count)
	FuncSliceOfLen(r, count, x, f)
}

// PhoenixBuilder specific func.
// Author: Happy2018new, Liliya233
//
// FuncSliceVarint16Length reads/writes a slice of T using function f with a varint16 length prefix.
func FuncSliceVarint16Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := int16(len(*x))
	r.Varint16(&count)
	FuncSliceOfLen(r, uint32(count), x, f)
}

// FuncSlice reads/writes a slice of T using function f with a varuint32 length prefix.
func FuncSlice[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	FuncSliceOfLen(r, count, x, f)
}

// PhoenixBuilder specific func.
// Author: Happy2018new, Liliya233
//
// FuncSliceVarint32Length reads/writes a slice of T using function f with a varint32 length prefix.
func FuncSliceVarint32Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := int32(len(*x))
	r.Varint32(&count)
	FuncSliceOfLen(r, uint32(count), x, f)
}

// FuncIOSlice reads/writes a slice of T using a function f with a varuint32 length prefix.
func FuncIOSlice[T any, S ~*[]T](r IO, x S, f func(IO, *T)) {
	FuncSlice(r, x, func(v *T) {
		f(r, v)
	})
}

// FuncIOSliceUint32Length reads/writes a slice of T using a function with a uint32 length prefix.
func FuncIOSliceUint32Length[T any, S ~*[]T](r IO, x S, f func(IO, *T)) {
	count := uint32(len(*x))
	r.Uint32(&count)
	FuncIOSliceOfLen(r, count, x, f)
}

const maxSliceLength = 1024

// SliceOfLen reads/writes the elements of a slice of type T with length l.
func SliceOfLen[T any, S ~*[]T, A PtrMarshaler[T]](r IO, l uint32, x S) {
	rd, reader := r.(*Reader)
	if reader {
		if rd.limitsEnabled && l > maxSliceLength {
			rd.panicf("slice length was too long: length of %v", l)
		}
		*x = make([]T, l)
	}

	for i := uint32(0); i < l; i++ {
		A(&(*x)[i]).Marshal(r)
	}
}

// FuncSliceOfLen reads/writes the elements of a slice of type T with length l using func f.
func FuncSliceOfLen[T any, S ~*[]T](r IO, l uint32, x S, f func(*T)) {
	rd, reader := r.(*Reader)
	if reader {
		if rd.limitsEnabled && l > maxSliceLength {
			rd.panicf("slice length was too long: length of %v", l)
		}
		*x = make([]T, l)
	}

	for i := uint32(0); i < l; i++ {
		f(&(*x)[i])
	}
}

// FuncIOSliceOfLen reads/writes the elements of a slice of type T with length l using func f.
func FuncIOSliceOfLen[T any, S ~*[]T](r IO, l uint32, x S, f func(IO, *T)) {
	FuncSliceOfLen(r, l, x, func(v *T) {
		f(r, v)
	})
}

// PtrMarshaler represents a type that implements Marshaler for its pointer.
type PtrMarshaler[T any] interface {
	Marshaler
	*T
}

// Single reads/writes a single Marshaler x.
func Single[T any, S PtrMarshaler[T]](r IO, x S) {
	x.Marshal(r)
}

// Optional is an optional type in the protocol. If not set, only a false bool is written. If set, a true bool is
// written and the Marshaler.
type Optional[T any] struct {
	set bool
	val T
}

// Option creates an Optional[T] with the value passed.
func Option[T any](val T) Optional[T] {
	return Optional[T]{set: true, val: val}
}

// Value returns the value set in the Optional. If no value was set, false is returned.
func (o Optional[T]) Value() (T, bool) {
	return o.val, o.set
}

// OptionalFunc reads/writes an Optional[T].
func OptionalFunc[T any](r IO, x *Optional[T], f func(*T)) any {
	r.Bool(&x.set)
	if x.set {
		f(&x.val)
	}
	return x
}

// OptionalMarshaler reads/writes an Optional assuming *T implements Marshaler.
func OptionalMarshaler[T any, A PtrMarshaler[T]](r IO, x *Optional[T]) {
	r.Bool(&x.set)
	if x.set {
		A(&x.val).Marshal(r)
	}
}

/*
PhoenixBuilder specific func.
Author: Happy2018new

NBTOptionalFunc 读写网易一个可选的字段 x 。

readPrefix 指代该字段是否在网易 __tag NBT 传输协议中可选，
此时若 x 为空，则仅写入 false 布尔值，
否则写入 true 布尔值和该字段的二进制表达形式。

f1 用于返回非空的 x 字段，
f2 则是用于 读取/写入 该字段的函数
*/
func NBTOptionalFunc[T any](r IO, x *T, f1 func() *T, readPrefix bool, f2 func(*T)) {
	var has bool
	if readPrefix {
		if x != nil {
			has = true
		}
		r.Bool(&has)
		if !has {
			return
		}
	}
	f2(f1())
}

/*
PhoenixBuilder specific func.
Author: Happy2018new

NBTOptionalMarshaler 读写网易一个可选的且已实现 Marshal 的字段 x 。

readPrefix 指代该字段是否在网易 __tag NBT 传输协议中可选，
此时若 x 为空，则仅写入 false 布尔值，
否则写入 true 布尔值和该字段的二进制表达形式。

x 用于返回非空的 x 字段
*/
func NBTOptionalMarshaler[T any, A PtrMarshaler[T]](r IO, x *T, f func() *T, readPrefix bool) {
	var has bool
	if readPrefix {
		if x != nil {
			has = true
		}
		r.Bool(&has)
		if !has {
			return
		}
	}
	A(f()).Marshal(r)
}

// PhoenixBuilder specific func.
// Author: Happy2018new
//
// 从 __tag NBT 的传输流以 T1 的数据类型 读取/写入 数据到 x 上。
// f 指代用于被用于传输流 解码/编码 网端 __tag NBT 的函数
func NBTInt[T1 Int, T2 TAGNumber](x *T2, f func(*T1)) {
	t2 := T1(*x)
	f(&t2)
	*x = T2(t2)
}

/*
PhoenixBuilder specific func.
Author: Happy2018new

在读取时，NBTSlice 使用 f 将底层输出流解码，
然后并转换为 []any 并输出到 x 上。
在写入时，NBTSlice 将 x 转换为 []T，
然后使用 f 向底层输出流编码
*/
func NBTSlice[T any](r IO, x *[]any, f func(*[]T)) {
	if _, isReader := r.(*Reader); isReader {
		new := make([]T, 0)
		f(&new)
		*x = make([]any, len(new))
		// read
		for key, value := range new {
			var mapping map[string]any
			// prepare
			val := reflect.ValueOf(value)
			valType := val.Kind()
			matchA := valType == reflect.Struct
			matchB := valType == reflect.Ptr && val.Elem().Kind() == reflect.Struct
			if !matchA && !matchB {
				(*x)[key] = value
				continue
			}
			// for normal data
			err := mapstructure.Decode(value, &mapping)
			if err != nil {
				panic(fmt.Sprintf("NBTSlice: %v", err))
			}
			(*x)[key] = mapping
			// for struct
		}
	} else {
		new := make([]T, len(*x))
		err := mapstructure.Decode(*x, &new)
		if err != nil {
			panic(fmt.Sprintf("NBTSlice: %v", err))
		}
		f(&new)
	}
}

// PhoenixBuilder specific func.
// Author: Happy2018new
//
// NBTSliceVarint16Length reads/writes a []any by using func SliceVarint16Length.
// s refer to the true data type of this slice.
func NBTSliceVarint16Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x *[]any, s S) {
	NBTSlice(r, x, func(t *[]T) {
		SliceVarint16Length[T, S, A](r, t)
	})
}

// PhoenixBuilder specific func.
// Author: Happy2018new
//
// NBTFuncSliceVarint32Length reads/writes a []any by using func FuncSliceVarint32Length.
// f refer to the function which FuncSliceVarint32Length request to.
func NBTFuncSliceVarint32Length[T any, S ~*[]T](r IO, x *[]any, f func(*T)) {
	NBTSlice(r, x, func(t *[]T) {
		FuncSliceVarint32Length[T, S](r, t, f)
	})
}

// PhoenixBuilder specific func.
// Author: Happy2018new
//
// NBTOptionalSliceVarint16Length reads/writes an optional []any with a varint16 prefix.
// s refer to the true data type of this slice.
func NBTOptionalSliceVarint16Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x *[]any, s S) {
	var has bool
	if x != nil {
		has = true
	}
	r.Bool(&has)
	if has {
		NBTSlice(r, x, func(t *[]T) {
			SliceVarint16Length[T, S, A](r, t)
		})
	}
}
