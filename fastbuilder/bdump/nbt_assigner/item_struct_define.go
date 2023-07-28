package NBTAssigner

import (
	"phoenixbuilder/fastbuilder/environment/interfaces"
	"phoenixbuilder/fastbuilder/types"
)

// ------------------------- interface -------------------------

/*
GeneralItemNBT 提供了一个通用的接口，
以便于您可以方便的解析对应的 NBT 物品，
然后生成它并以最大的可能性注入 NBT 数据。

该接口实际与下方的 Item 结构体绑定
*/
type GeneralItemNBT interface {
	// 解码
	Decode() error
	// 生成物品并尽可能注入 NBT
	WriteData() error
	// 应当只被外部实现调用，
	// 用于判断当前物品是否可以仅使用命令生成
	SpecialCheck() (bool, error)
}

// ------------------------- general -------------------------

// 描述单个物品在解码前的 NBT 表达形式
type ItemOrigin map[string]interface{}

// 描述物品的单个附魔属性
type Enchantment struct {
	ID    uint8 // 该附魔属性的 ID
	Level int16 // 该附魔属性的等级
}

// 描述单个物品拥有的物品组件数据
type ItemComponents struct {
	// 控制此物品（在冒险模式下）可以使用在其上的方块类型
	CanPlaceOn []string
	// 控制此物品（在冒险模式下）可以放置在其上的方块类型
	CanDestroy []string
	// 将该物品锁定在玩家的物品栏
	ItemLock string
	// 使该物品在玩家死亡时不会掉落
	KeepOnDeath bool
}

// 描述单个物品的自定义数据。
// 这些数据实际上并不存在，
// 只是我们为了区分一些特殊的物品而设
type ItemCustomData struct {
	/*
		如果该物品本身就是一个带有 NBT 的方块，
		那么如果我们已经在 PhoenixBuilder 实现了这些方块的 NBT 注入，
		那么对于容器内的这些物品来说，
		我们仍然也可以通过 PickBlock 的方法来实现对它们的兼容。

		因此，如果该物品带有 NBT 且是一个方块，
		那么此字段不为空指针。
	*/
	SubBlockData GeneralBlockNBT
	/*
		指代该物品的 Tag 标签。

		这个物品可能是一本写了字或者签过名的书，
		也可能是一个由工作台合成的烟花。

		它与上一个字段的不同之处在于，
		这些都是带有 NBT 的物品，而非方块。

		那么如果我们已经在 PhoenixBuilder 实现了这些特殊物品的支持，
		那么对于容器内的这些物品来说，我们仍然也可以兼容。
		因此，如果该物品带有 NBT ，那么此字段不为空指针。

		检测这个物品是否需要使用特殊方法获取可能是必要的，
		因为用指令获得的烟花可能也是带有 NBT 的。 [需要验证]

		TODO: 兼容此特性
	*/
	ItemTag ItemOrigin
}

// 描述单个物品的基本数据
type ItemBasicData struct {
	// 该物品的名称
	Name string
	// 该物品的数量
	Count uint8
	// 该物品的数据值(附加值)
	MetaData uint16
	/*
		指代该物品原本所在的槽位。

		注：
		我们总是将物品生成在快捷栏中，
		因此此字段可能不代表物品的实际生成位置，
		您需要自行移动它们到正确的位置上
	*/
	Slot uint8
}

// 描述单个物品的附加数据
type ItemEnhancementData struct {
	DisplayName    string          // 该物品的显示名称
	Enchantments   *[]Enchantment  // 该物品的附魔属性
	ItemComponents *ItemComponents // 该物品的物品组件
}

// GeneralItem 结构体用于一个通用型的物品
type GeneralItem struct {
	Basic       ItemBasicData        // 该物品的基本数据
	Enhancement *ItemEnhancementData // 该物品的增强数据
	Custom      *ItemCustomData      // 由 PhoenixBuilder 定义的自定义数据
}

// AdditionalData 结构体用于描述一个物品的其他附加数据，例如该物品应该生成的位置
type ItemAdditionalData struct {
	// 如果该物品已被解码，
	// 则此段为真，否则为假
	Decoded bool
	// 指定该物品实际的生成位置。
	// 我们总是将物品生成在快捷栏
	HotBarSlot uint8
	// 指代生成 NBT 物品时可能需要预留的空位。
	// 此位置所对应的方块将可能被占用，
	// 通常情况下将被生成铁砧或工作台，
	// 例如部分情况下可能需要用铁砧修改物品的显示名称
	Position [3]int32
	// 该物品的类型，例如各式各样的告示牌可以写作 Sign
	// TODO: USE ENUM INSTEAD
	Type string
	// 此参数应当只被 PhoenixBuilder 使用，除非 Omega 也需要设置一些功能
	Settings *types.MainConfig
	// 是否是快速模式生成物品；若为真，则大多数方块物品的 NBT 数据将不会被注册
	FastMode bool
	// 部分情况下可能会携带的不定数据，通常情况下应该为空 [目前还未使用此字段]
	Others interface{}
}

// ItemPackage 是用于包装每个物品的结构体
type ItemPackage struct {
	// 储存生成该物品所需的 API ，例如发包需要用到的函数等
	// 此参数需要外部实现主动赋值，
	// 主要是为了兼容 Omega 和 PhoenixBuilder 对功能的同时使用
	Interface interfaces.GameInterface
	// 一个通用型物品的数据，例如名称和数据值(附加值)
	Item GeneralItem
	// 此物品的其他附加数据，例如物品应该被在快捷栏生成的位置
	AdditionalData ItemAdditionalData
}

// ------------------------- book -------------------------

// 描述单个成书中已解码的部分
type BookData struct {
	Pages  []string // pages(TAG_List) = []string{}
	Author string   // author(TAG_String) = ""
	Title  string   // title(TAG_String) = ""
}

// Book 结构体用于描述一个完整的成书的数据
type Book struct {
	// 该 NBT 物品的详细数据
	ItemPackage *ItemPackage
	// 存放已解码的成书数据
	BookData BookData
}
