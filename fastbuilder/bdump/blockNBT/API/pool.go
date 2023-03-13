package blockNBT_API

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"

	"github.com/google/uuid"
)

// 描述了一个通用型的 API ，它可以被用在任何地方，但主要被用于 Bdump/blockNBT
type GlobalAPI struct {
	WritePacket        func(packet.Packet) error // 用于向租赁服发送数据包的函数
	BotName            string                    // 客户端的游戏昵称
	BotIdentity        string                    // 客户端的唯一标识符 [当前还未使用]
	BotUniqueID        int64                     // 客户端的唯一 ID [当前还未使用]
	BotRunTimeID       uint64                    // 客户端的运行时 ID
	PacketHandleResult *PacketHandleResult       // 保存包处理结果；由外部实现实时更新
}

// 保存包处理结果。理论上，这些结果应该由此结构体下的 HandlePacket 实现实时更新
type PacketHandleResult struct {
	commandDatas       commandRequestWithResponce   // 命令请求队列及命令返回值；此实现及内部所有实现不予公开
	Inventory          inventoryContents            // 存放所有有效库存中的物品数据，例如背包和盔甲栏；此实现不公开但可以通过公开的函数实现对数据的访问
	ItemStackOperation itemStackReuqestWithResponce // 存放物品更改请求及结果；此实现不公开但可以通过公开的函数实现对数据的访问和修改
	ContainerResources container                    // 描述容器公用资源的占用状态并同时存放容器的打开(关闭)状态及打开(关闭)数据；此实现不公开但可以通过公开的函数实现对资源的访问、占用和释放
}

// 存放命令请求及结果
type commandRequestWithResponce struct {
	// 命令请求队列
	commandRequest struct {
		lockDown sync.RWMutex              // 防止并发读写而设置的读写锁；此实现不公开
		datas    map[uuid.UUID]*sync.Mutex // 存放命令请求的等待队列。每次写入请求后将会自动为此请求上锁以便于阻塞。此实现不公开
	}
	// 命令请求的返回值
	commandResponce struct {
		lockDown sync.RWMutex                       // 防止并发读写而设置的读写锁；此实现不公开
		datas    map[uuid.UUID]packet.CommandOutput // 存放命令返回值。每次写入返回值后将会自动为对应等待队列中的读写锁解锁。此实现不公开
	}
}

// 存放所有有效库存中的物品数据，例如背包和盔甲栏；此实现不公开但可以通过公开的函数实现对数据的访问
type inventoryContents struct {
	lockDown sync.RWMutex                               // 防止并发读写而设置的读写锁；此实现不公开
	datas    map[uint32]map[uint8]protocol.ItemInstance // int32 代表打开的库存的窗口 ID ，即 WindowID ；uint8 代表物品所在的槽位；最内层的 protocol.ItemInstance 存放物品数据；此实现不公开
}

/*
存放物品更改请求及结果；此实现不公开但可以通过公开的函数实现对数据的访问和修改。
重要：
任何物品操作都应该通过此结构体下的有关实现来完成，否则可能会造成严重后果。
因此，为了绝对的安全，如果尝试绕过相关实现而直接发送物品操作数据包，则会造成程序惊慌
*/
type itemStackReuqestWithResponce struct {
	// 物品操作请求队列
	itemStackRequest struct {
		lockDown sync.RWMutex          // 防止并发读写而设置的读写锁；此实现不公开
		datas    map[int32]*sync.Mutex // 存放物品操作请求的等待队列。每次写入请求后将会自动为此请求上锁以便于阻塞。此实现不公开
	}
	// 物品操作的结果
	itemStackResponce struct {
		lockDown sync.RWMutex                         // 防止并发读写而设置的读写锁；此实现不公开
		datas    map[int32]protocol.ItemStackResponse // 存放物品操作的结果。每每次写入返回值后将会自动为对应等待队列中的读写锁解锁。此实现不公开
	}
	/*
		客户端在发送 ItemStackRequest 时需要发送一个 RequestID 。
		经过观察，这个值会随着请求发送的次数递减，且呈现为公差为 -2，首项为 -1 的递减型等差数列。
		特别地，如果你尝试在 RequestID 字段填写非负数或者偶数，那么客户端会被租赁服强制断开连接。
		尽管始终为 ItemStackRequest 的 RequestID 字段填写 -1 并不会造成造成断开连接的发生，
		但这样并不能保证物品操作的唯一性。
		因此，为了安全性，请使用本结构体下的相关实现构造物品操作的发起函数，否则可能会造成潜在的问题。
		其他：本结构体的相关实现在修改这个值时，将会以原子操作修改
	*/
	requestID int32
}

/*
描述容器公用资源的占用状态并同时存放容器的打开(关闭)状态及打开(关闭)数据；
此实现不公开但可以通过公开的函数实现对资源的访问、占用和释放。
重要：
容器由于是客户端的其中一个公用资源，因此为了公平性，现在由 blockNBT_API 负责完成对该公用资源的占用和释放之实现。
因此，为了绝对的安全，如果尝试绕过相关实现而直接打开容器/关闭容器，则会造成程序惊慌。
任何时刻，如果你需要打开或关闭容器，或者在某一段时间内使用某容器，则请提前占用此资源，然后再发送相应数据包，
完成后再释放此公用资源
*/
type container struct {
	// 容器被打开时的数据
	containerOpen struct {
		lockDown sync.RWMutex // 防止并发读写而设置的读写锁；此实现不公开
		/*
			当客户端打开容器后，租赁服会以此数据包回应，届时此变量将被赋值。
			当容器被关闭时，此变量将被重置
		*/
		datas packet.ContainerOpen
	}
	// 容器被关闭时的数据
	containerClose struct {
		lockDown sync.RWMutex // 防止并发读写而设置的读写锁；此实现不公开
		/*
			客户端可以使用该数据包关闭已经打开的容器，而后，租赁服会以相同的数据包回应容器的关闭。
			当侦测到来自租赁服的响应，此变量将被赋值。
			当容器被打开时，此变量将被重置
		*/
		datas packet.ContainerClose
	}
	awaitChanges sync.Mutex // 客户端在打开或关闭容器后可能需要等待回应，此互斥锁便是为了完成这一实现
	isUsing      sync.Mutex // 客户端同一时刻至多打开一个容器。此互斥锁是为了解决资源纠纷问题而设
}
