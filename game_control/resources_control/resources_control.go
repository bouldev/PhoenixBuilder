package ResourcesControl

import (
	"context"
	"phoenixbuilder/fastbuilder/generics"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft/ai_command"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
	"time"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/google/uuid"
)

// ------------------------- Resources -------------------------

// 管理和保存 PhoenixBuilder 的各类公用资源
type Resources struct {
	// 管理命令请求队列及命令返回值
	Command command_request_with_response
	// 管理本地库存数据，如背包物品
	Inventory inventory_contents
	// 管理物品操作请求及结果
	ItemStackOperation item_stack_request_with_response
	// 管理容器资源的占用状态，同时存储容器操作的结果
	Container container
	// 管理结构资源并保存结构请求的回应
	Structure mcstructure
	// 数据包监听器
	Listener packet_listener
	// 管理和保存其他小型的资源，
	// 例如游戏刻相关
	Others others
}

// ------------------------- commandRequestWithResponce -------------------------

/*
存放命令请求及结果

!!! 重要 !!!
任何与 魔法指令 相关的操作都应该通过此结构体下的有关实现来完成，否则可能会造成严重后果。
因此，为了不出现问题，如果尝试绕过相关实现而直接发送 魔法指令 相关的数据包，则会造成程序 panic
*/
type command_request_with_response struct {
	/*
		魔法指令 共计有两种响应体，
		其一是 PyRpc 数据包(由多个数据包共同描述)，
		另外一种则是标准响应体，
		即数据类型为 packet.CommandOutput 的响应体。

		不幸的是，这两个响应体相互独立，
		我们无法通过这两个响应体中填写的 UUID 字段来进行一对一匹配，
		而只能采用各个数据包从服务器发往客户端的时间顺序来进行匹配。

		该字段的作用便是为此目的而设，
		通过提前记录标准型命令响应体的方式，
		在下一个 PyRpc 类型的响应体到来后，
		将该字段与 PyRpc 响应体相互匹配
	*/
	ai_command_resp *packet.CommandOutput
	/*
		存放命令请求。
		request_lock 用于防止可能的并发读写。

		考虑到 魔法指令 的一些特性，
		也就是在前置检查失败时，
		魔法指令 的响应体不会记录原始请求的 UUID ，
		因此我们需要依靠数据包从客户端发往服务器的实
		际时间来推断响应体对应的 魔法指令 命令请求，
		而 orderedmap 便提供了有序 map 的功能。

		除此外，还需要需要保证 魔法指令 的注册及相应
		数据包的发送所构成的整个过程保持原子性，
		因此我们提供使用 resources_occupy 来保持了这一点。

		固，如果尝试绕过相关实现而直接发送 魔法指令 相
		关的数据包，则会造成程序 panic
	*/
	resources_occupy
	request_lock sync.RWMutex
	request      *orderedmap.OrderedMap[uuid.UUID, CommandRequestOptions]
	// 存放命令请求对应的响应体
	response generics.SyncMap[uuid.UUID, *CommandRespond]
	// 用于传递信号，例如通知可以加载命令响应体，
	// 或通知 魔法指令 中的标准响应体已经到来
	signal generics.SyncMap[uuid.UUID, chan uint8]
}

// 指定单个命令请求中可以自定义的设置项
type CommandRequestOptions struct {
	// 描述当前命令请求的最长截止时间，
	// 当抵达该时间后，将返回超时错误。
	// 如果此字段为 0 ，则将永远等待，
	// 直到客户端收到对应的响应体
	TimeOut time.Duration
	/*
		指定当前命令无需追踪其对应的返回值。

		值得注意的是，响应体在实质上仍被追踪，
		它仍会被计入请求列表并经过资源控制中心，
		同时作为与其他命令请求同级的存在。
		区别于其他响应体，
		该字段为真的响应体会在对应的响应体到来时，
		由资源控制中心在内部赋值、加载和删除。

		因此，基于以上的缘由，
		此字段应当仅用于 魔法指令 的执行，
		因为 魔法指令 被要求必须经由资源控制中心发送
	*/
	WithNoResponse bool
}

// 描述 魔法指令 的响应体。
// 该响应体不是命令请求的标准响应体，
// 而是由多个 PyRpc 数据包共同描述
type AICommandDetails struct {
	// 描述 魔法指令 是否成功。
	// 若前置检查失败，则此字段为空值
	Result *ai_command.AfterExecuteCommandEvent
	// 描述 魔法指令 的输出。
	// 如果命令失败，则此切片为空
	Output []ai_command.ExecuteCommandOutputEvent
	// 命令请求在前置检查失败时的响应体，
	// 通常在 作弊未开启 或 机器人不具备操作员权限 时触发，
	// 可能不存在
	PreCheckError *ai_command.AvailableCheckFailed
}

// 描述命令请求的响应体
type CommandRespond struct {
	// 命令请求的标准响应体。
	// 如果原始请求是 魔法指令 ，
	// 则可能不存在
	Respond *packet.CommandOutput
	// 描述 魔法指令 的独有响应体，
	// 该部分由多个 PyRpc 数据包决定。
	// 如果原始请求不是 魔法指令 ，
	// 则该字段不存在
	AICommand *AICommandDetails
	// 描述原始请求的类型，
	// 例如这是一个 标准命令 ，
	// 亦或是一个 魔法指令
	Type string
	// 获取响应体时发生错误信息，
	// 可能不存在
	Error error
	// 如果获取响应体时发生了错误，
	// 那么此字段非 0 ，否则为 0
	ErrorType uint8
}

// ------------------------- inventory_contents -------------------------

// 存放所有有效库存中的物品数据，例如背包和盔甲栏
type inventory_contents struct {
	// 防止并发读写而设置的读写锁
	lock_down sync.RWMutex
	// uint32 代表打开的库存的窗口 ID ，即 WindowID ；
	// uint8 代表物品所在的槽位；
	// 最内层的 protocol.ItemInstance 存放物品数据
	data generics.SyncMap[uint32, *generics.SyncMap[uint8, protocol.ItemInstance]]
}

// ------------------------- itemStackReuqestWithResponse -------------------------

/*
存放物品操作请求及结果。

!!! 重要 !!!
任何物品操作都应该通过此结构体下的有关实现来完成，否则可能会造成严重后果。
因此，为了绝对的安全，如果尝试绕过相关实现而直接发送物品操作数据包，则会造成程序 panic
*/
type item_stack_request_with_response struct {
	// 存放物品操作的请求队列
	request_with_response generics.SyncMap[int32, singleitem_stack_request_with_response]
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
	current_request_id int32
}

// 每个物品操作请求都会使用这样一个结构体，它用于描述单个的物品操作请求
type singleitem_stack_request_with_response struct {
	// 描述物品操作请求的返回值
	resp chan protocol.ItemStackResponse
	// 描述多个库存(容器)中物品的变动结果。
	// 租赁服不会在返回 ItemStackResponce 时返回完整的物品数据，因此需要您提供对应
	// 槽位的更改结果以便于我们依此更新本地存储的库存数据
	how_to_change map[ContainerID]StackRequestContainerInfo
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
	// 描述容器资源的占用状态及占用者
	resources_occupy
	// 防止并发读写而安排的读写锁
	lock_down sync.RWMutex
	// 存放容器被打开时的数据。
	// 当客户端打开容器后，租赁服会以此数据包回应，届时此变量将被赋值。
	// 当容器被关闭或从未被打开，则此变量将会为 nil
	container_opening_data *packet.ContainerOpen
	/*
		存放容器被关闭时的数据。

		客户端可以使用该数据包关闭已经打开的容器，
		而后，租赁服会以相同的数据包回应容器的关闭。

		当侦测到来自租赁服的响应，此变量将被赋值。
		当容器被打开或从未被关闭，则此变量将会为 nil
	*/
	container_closing_data *packet.ContainerClose
	// 其他实现在打开或关闭容器后可能需要等待回应，此管道便是为了完成这一实现
	responded chan struct{}
}

// ------------------------- mcstructure -------------------------

// 记录结构资源并保存结构请求的回应
type mcstructure struct {
	// 描述结构资源的占用状态及占用者
	resources_occupy
	// 保存结构请求的响应体
	resp chan packet.StructureTemplateDataResponse
}

// ------------------------- packet_listener -------------------------

// 储存单次监听请求下所需要保存的数据
type single_listen struct {
	// 指代本次请求中欲监听的数据包 ID
	packets_id []uint32
	// 用于存放本次请求中已经监听的数据包
	packet_received chan (packet.Packet)
	// 标记该监听器下有多少个协程正在尝试分发数据包。
	// 我们最多允许同时存在 MaximumCoroutinesRunningCount 个这样的协程，
	// 对于超出的部分，对应的数据包将被丢弃
	running_counts int32
	// 如果监听者终止并关闭了当次监听，
	// 则相应的上层实现会取消该上下文，
	// 以表明相关联的所有监听协程均应当关闭
	ctx context.Context
	// 当调用此函数时，
	// 监听器将终止并关闭
	stop context.CancelFunc
}

// 数据包监听器
type packet_listener struct {
	// 键代表监听器，而值代表此监听器下已保存的数据
	listener_with_data generics.SyncMap[uuid.UUID, single_listen]
}

// ------------------------- others -------------------------

// 记录其他小型的资源，例如游戏刻相关
type others struct {
	// 存放 TickSync 请求并保存其对应的返回值，
	// 它用于获取当前的游戏刻。
	current_tick_request_with_resp generics.SyncMap[uuid.UUID, chan int64]
}
