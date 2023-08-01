package ResourcesControl

import (
	"context"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
	"time"
)

// ------------------------- Resources -------------------------

// 管理和保存 PhoenixBuilder 的各类公用资源
type Resources struct {
	// 管理命令请求队列及命令返回值
	Command commandRequestWithResponse
	// 管理本地库存数据，如背包物品
	Inventory inventoryContents
	// 管理物品操作请求及结果
	ItemStackOperation itemStackRequestWithResponse
	// 管理容器资源的占用状态，同时存储容器操作的结果
	Container container
	// 管理结构资源并保存结构请求的回应
	Structure mcstructure
	// 数据包监听器
	Listener packetListener
	// 管理和保存其他小型的资源，
	// 例如游戏刻相关
	Others others
}

// ------------------------- commandRequestWithResponce -------------------------

// 指定单个命令请求中可以自定义的设置项
type CommandRequestOptions struct {
	// 描述当前命令请求的最长截止时间，
	// 当抵达该时间后，将返回超时错误。
	// 如果此字段为 0 ，则将永远等待，
	// 直到客户端收到对应的响应体
	TimeOut time.Duration
}

// 存放命令请求及结果
type commandRequestWithResponse struct {
	// 存放命令请求。
	// 数据类型为 map[uuid.UUID]CommandRequestOptions
	request sync.Map
	// 存放命令请求的响应体。
	// 数据类型为 map[uuid.UUID](chan packet.CommandOutput)
	response sync.Map
}

// 描述命令请求的响应体
type CommandRespond struct {
	// 来自租赁服的响应体
	Respond packet.CommandOutput
	// 获取响应体时发生错误信息，
	// 可能不存在
	Error error
	// 如果获取响应体时发生了错误，
	// 那么此字段非 0 ，否则为 0
	ErrorType uint8
}

// ------------------------- inventoryContents -------------------------

// 存放所有有效库存中的物品数据，例如背包和盔甲栏
type inventoryContents struct {
	// 防止并发读写而设置的读写锁
	lockDown sync.RWMutex
	// uint32 代表打开的库存的窗口 ID ，即 WindowID ；
	// uint8 代表物品所在的槽位；
	// 最内层的 protocol.ItemInstance 存放物品数据
	datas map[uint32]map[uint8]protocol.ItemInstance
}

// ------------------------- itemStackReuqestWithResponse -------------------------

/*
存放物品操作请求及结果。

!!! 重要 !!!
任何物品操作都应该通过此结构体下的有关实现来完成，否则可能会造成严重后果。
因此，为了绝对的安全，如果尝试绕过相关实现而直接发送物品操作数据包，则会造成程序 panic
*/
type itemStackRequestWithResponse struct {
	// 存放物品操作的请求队列。
	// 数据类型为 map[int32]singleItemStackRequestWithResponse
	requestWithResponse sync.Map
	/*
		记录已累计的 RequestID 。

		客户端在发送 ItemStackRequest 时需要发送一个 RequestID 。
		经过观察，这个值会随着请求发送的次数递减，且呈现为公差为 -2，
		首项为 -1 的递减型等差数列。

		特别地，如果你尝试在 RequestID 字段填写非负数或者偶数，
		那么客户端会被租赁服强制断开连接。

		尽管始终为 ItemStackRequest 的 RequestID 字段填写 -1 并不会造成造成断开连接的发生，
		但这样并不能保证物品操作的唯一性。

		因此，绝对地，请使用已提供的 API 发送物品操作请求，否则将导致程序 panic
	*/
	currentRequestID int32
}

// 每个物品操作请求都会使用这样一个结构体，它用于描述单个的物品操作请求
type singleItemStackRequestWithResponse struct {
	// 描述物品操作请求的返回值
	resp chan protocol.ItemStackResponse
	// 描述多个库存(容器)中物品的变动结果。
	// 租赁服不会在返回 ItemStackResponce 时返回完整的物品数据，因此需要您提供对应
	// 槽位的更改结果以便于我们依此更新本地存储的库存数据
	howToChange map[ContainerID]StackRequestContainerInfo
}

// 描述单个库存(容器)中物品的变动结果
type StackRequestContainerInfo struct {
	// 其容器对应库存的窗口 ID
	WindowID uint32
	// 描述此容器中每个槽位的变动结果，键代表槽位编号，而值代表物品的新值。
	// 特别地，您无需设置物品数量以及物品的 StackNetworkID 信息，
	// 因为这些数据会在租赁服发回 ItemStackResponce 后被重新设置
	ChangeResult map[uint8]protocol.ItemInstance
}

// ------------------------- container -------------------------

// 描述一个容器 ID
type ContainerID uint8

/*
存储容器的 打开/关闭 状态，同时存储容器资源的占用状态。

!!! 重要 !!!
容器由于是 PhoenixBuilder 的其中一个公用资源，因此为了公平性，
现在由我们(资源管理中心)负责完成对该公用资源的占用和释放之实现。

因此，为了绝对的安全，如果尝试绕过相关实现而直接 打开/关闭 容器，则会造成程序 panic 。

任何时刻，如果你需要打开或关闭容器，或者在某一段时间内使用某容器，则请提前占用此资源，
然后再发送相应数据包，完成后再释放此公用资源
*/
type container struct {
	// 防止并发读写而安排的读写锁
	lockDown sync.RWMutex
	// 存放容器被打开时的数据。
	// 当客户端打开容器后，租赁服会以此数据包回应，届时此变量将被赋值。
	// 当容器被关闭或从未被打开，则此变量将会为 nil
	containerOpeningData *packet.ContainerOpen
	/*
		存放容器被关闭时的数据。

		客户端可以使用该数据包关闭已经打开的容器，
		而后，租赁服会以相同的数据包回应容器的关闭。

		当侦测到来自租赁服的响应，此变量将被赋值。
		当容器被打开或从未被关闭，则此变量将会为 nil
	*/
	containerClosingData *packet.ContainerClose
	// 其他实现在打开或关闭容器后可能需要等待回应，此管道便是为了完成这一实现
	responded chan struct{}
	// 描述容器资源的占用状态及占用者
	resourcesOccupy
}

// ------------------------- mcstructure -------------------------

// 记录结构资源并保存结构请求的回应
type mcstructure struct {
	// 描述结构资源的占用状态及占用者
	resourcesOccupy
	// 保存结构请求的响应体
	resp chan packet.StructureTemplateDataResponse
}

// ------------------------- packetListener -------------------------

// 储存单次监听请求下所需要保存的数据
type singleListen struct {
	// 指代本次请求中欲监听的数据包 ID
	packetsID []uint32
	// 用于存放本次请求中已经监听的数据包
	packetReceived chan (packet.Packet)
	// 标记该监听器下有多少个协程正在尝试分发数据包。
	// 我们最多允许同时存在 MaximumCoroutinesRunningCount 个这样的协程，
	// 对于超出的部分，对应的数据包将被丢弃
	runningCounts int32
	// 如果监听者终止并关闭了当次监听，
	// 则相应的上层实现会取消该上下文，
	// 以表明相关联的所有监听协程均应当关闭
	ctx context.Context
	// 当调用此函数时，
	// 监听器将终止并关闭
	stop context.CancelFunc
}

// 数据包监听器
type packetListener struct {
	// 数据类型为 map[uuid.UUID]singleListen 。
	// 键代表监听器，而值代表此监听器下已保存的数据
	listenerWithData sync.Map
}

// ------------------------- others -------------------------

// 记录其他小型的资源，例如游戏刻相关
type others struct {
	// 存放 TickSync 请求并保存其对应的返回值，
	// 它用于获取当前的游戏刻。
	// 数据类型为 map[uuid.UUID]chan int64
	currentTickRequestWithResp sync.Map
}
