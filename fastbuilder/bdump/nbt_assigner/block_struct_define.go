package NBTAssigner

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment/interfaces"
	"phoenixbuilder/fastbuilder/types"
)

// ------------------------- interface -------------------------

/*
GeneralBlockNBT 提供了一个通用的接口，
以便于您可以方便的解析对应的方块实体，
然后放置它并以最大的可能性注入 NBT 数据。

该接口实际与下方的 BlockEntity 结构体绑定
*/
type GeneralBlockNBT interface {
	// 解码
	Decode() error
	// 放置方块并尽可能注入 NBT
	WriteData() error
}

// ------------------------- general -------------------------

// GeneralBlock 结构体用于一个通用型的方块
type GeneralBlock struct {
	// 方块名称(不含命名空间且应该全部小写)
	Name string
	// 方块状态
	States map[string]interface{}
	// 当前方块所携带的 NBT 数据
	NBT map[string]interface{}
}

// AdditionalData 结构体用于描述一个方块实体的其他附加数据，例如方块的绝对坐标
type BlockAdditionalData struct {
	// 字符串形式的方块状态，用于在放置方块时使用
	BlockStates string
	// 方块坐标(绝对坐标)
	Position [3]int32
	// 该方块的类型，例如各式各样的告示牌可以写作 Sign
	// TODO: USE ENUM INSTEAD
	Type string
	// 此参数应当只被 PhoenixBuilder 使用，除非 Omega 也需要设置一些功能
	Settings *types.MainConfig
	// 是否是快速模式放置；若为真，则大多数方块实体的 NBT 数据将不会被注册
	FastMode bool
	// 部分情况下可能会携带的不定数据，通常情况下应该为空 [目前还未使用此字段]
	Others interface{}
}

// BlockEntity 是用于包装每个方块实体的结构体
type BlockEntity struct {
	// 储存放置该方块实体所需的 API ，例如发包需要用到的函数等
	// 此参数需要外部实现主动赋值，
	// 主要是为了兼容 Omega 和 PhoenixBuilder 对功能的同时使用
	Interface interfaces.GameInterface
	// 一个通用型方块的数据，例如名称、方块状态和所携带的 NBT 数据
	Block GeneralBlock
	// 此方块的其他附加数据，例如方块的绝对坐标
	AdditionalData BlockAdditionalData
}

// ------------------------- command_block -------------------------

// 描述单个命令方块中已解码的部分
type CommandBlockData struct {
	Command            string // Command(TAG_String) = ""
	CustomName         string // CustomName(TAG_String) = ""
	LastOutput         string // LastOutput(TAG_String) = ""
	TickDelay          int32  // TickDelay(TAG_Int) = 0
	ExecuteOnFirstTick bool   // ExecuteOnFirstTick(TAG_Byte) = 1
	TrackOutput        bool   // TrackOutput(TAG_Byte) = 1
	ConditionalMode    bool   // conditionalMode(TAG_Byte) = 0
	Auto               bool   // auto(TAG_Byte) = 1
	Version            int32  // Version(TAG_Int) = 35
}

// CommandBlock 结构体用于描述一个完整的命令方块数据
type CommandBlock struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 存放已解码的命令方块数据
	CommandBlockData CommandBlockData
	// 为向下兼容而设，因为旧方法下不需要放置命令方块
	ShouldPlaceBlock bool
}

// ------------------------- container -------------------------

// 描述单个容器的固定数据，
// 这被用于 SupportContainerPool
type ContainerConstantData struct {
	// 描述此容器放置物品所使用的复合标签或列表
	StorageItemValue string
	// 描述此容器是否可以打开
	CouldOpen bool
	/*
		描述此容器的通用容器 ID ，
		因为大多数容器只存在唯一的容器 ID。

		当然，如果该容器不同槽位对应的容器 ID 不同，
		则该字段为 255 且下方 ContainerIDMapping 字段为非空。

		另，如果该字段为 255 且下方 ContainerIDMapping 为空，
		则代表该容器未被支持
	*/
	UniversalContainerID uint8
	/*
		描述此容器对应槽位的容器 ID 。

		目前应只被熔炉、烟熏炉、
		高炉 和 酿造台 所使用。

		键代表槽位编号，
		值代表该槽位对应的容器 ID 。

		另，如果该字段为空且上方 UniversalContainerID 为 255，
		则代表该容器未被支持
	*/
	ContainerIDMapping map[uint8]uint8
}

// 描述一个容器
type Container struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 容器的内容物
	Contents []ItemPackage
}

// 未被支持的容器会被应用此错误信息。
// 用于 Container.go 中的 getContainerContents 等函数
var ErrNotASupportedContainer error = fmt.Errorf("getContainerContents: Not a supported container")

// ------------------------- sign -------------------------

// 描述单个告示牌(旧版)中已解码的部分
type LegacySignData SignText

// 描述单个告示牌中已解码的部分
type SignData struct {
	FrontText SignText // FrontText(TAG_Compound)
	BackText  SignText // FrontText(TAG_Compound)
	IsWaxed   bool     // IsWaxed(TAG_Byte) = 0
}

// 描述单个告示牌 FrontText 或 BackText 字段中已解码的部分
type SignText struct {
	IgnoreLighting bool  // IgnoreLighting(TAG_Byte) = 0
	SignTextColor  int32 // SignTextColor(TAG_Int) = -16777216
}

// 描述一个告示牌
type Sign struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 存放已解码的告示牌(旧版)数据。
	// 如果当前告示牌不是旧版，
	// 则该字段不存在
	LegacySignData *LegacySignData
	// 存放已解码的告示牌数据。
	// 如果当前告示牌是旧版，
	// 则该字段不存在
	SignData *SignData
	// 指定当前告示牌是否是悬挂式
	IsHangingSignBlock bool
	// 指定当前告示牌是否是非旧版
	IsNotLegacySignBlock bool
}

// ------------------------- frame -------------------------

// 描述单个物品展示框中已解码的部分
type FrameData struct {
	ItemRotation float32      // ItemRotation(TAG_Float) = 0
	Item         *ItemPackage // Item(TAG_Compound) = ...
}

// 描述一个物品展示框
type Frame struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 存放已解码的物品展示框数据
	FrameData FrameData
}

// ------------------------- structure_block -------------------------

// 描述单个结构方块中已解码的部分
type StructureBlockData struct {
	AnimationMode    byte    // animationMode(TAG_Byte) = 0
	AnimationSeconds float32 // animationSeconds(TAG_Float) = 0
	Data             int32   // data(TAG_Int) = 1
	DataField        string  // dataField(TAG_String) = ""
	IgnoreEntities   bool    // ignoreEntities(TAG_Byte) = 0
	IncludePlayers   bool    // includePlayers(TAG_Byte) = 0
	Integrity        float32 // integrity(TAG_Float32) = 100
	Mirror           byte    // mirror(TAG_Byte) = 0
	RedstoneSaveMode int32   // redstoneSaveMode(TAG_Int) = 0
	RemoveBlocks     bool    // removeBlocks(TAG_Byte) = 0
	Rotation         byte    // rotation(TAG_Byte) = 0
	Seed             int64   // seed(TAG_Long) = 0
	ShowBoundingBox  bool    // showBoundingBox(TAG_Byte) = 0
	StructureName    string  // structureName(TAG_String) = ""
	XStructureOffset int32   // xStructureOffset(TAG_int) = 0
	XStructureSize   int32   // xStructureSize(TAG_int) = 5
	YStructureOffset int32   // yStructureOffset(TAG_int) = -1
	YStructureSize   int32   // xStructureSize(TAG_int) = 5
	ZStructureOffset int32   // zStructureOffset(TAG_int) = 0
	ZStructureSize   int32   // xStructureSize(TAG_int) = 5
}

// 描述一个结构方块
type StructureBlock struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 存放已解码的结构方块数据
	StructureBlockData StructureBlockData
}

// ------------------------- banner -------------------------

// 描述单个旗帜方块中已解码的部分
type BannerBlockData struct {
	Base     int32           // Base(TAG_Int) = 0
	Patterns []BannerPattern // Patterns(TAG_List) = []BannerPattern{}
	Type     int32           // Type(TAG_List) = 0
}

// 描述一个旗帜方块
type BannerBlock struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 存放已解码的旗帜方块数据
	BannerBlockData BannerBlockData
}
