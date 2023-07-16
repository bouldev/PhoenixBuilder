package mux_pumper

import (
	"context"
	"phoenixbuilder/fastbuilder/lib/utils/sync_wrapper"
	"phoenixbuilder/minecraft/protocol/packet"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// MCPacketNameIDMapping 表示 Minecraft 游戏协议包的名字与 ID 的映射关系
type MCPacketNameIDMapping map[string]uint32

var mcPacketNameIDMapping MCPacketNameIDMapping

// 初始化游戏包id与name的对应表
func initMCPacketNameIDMapping() {
	pool := packet.NewPool()
	mcPacketNameIDMapping = MCPacketNameIDMapping{}
	for id, pkMaker := range pool {
		pk := pkMaker()
		pkName := reflect.TypeOf(pk).Elem().Name()
		mcPacketNameIDMapping[pkName] = id
		// mcPacketNameIDMapping["ID"+pkName] = id
		// mcPacketNameIDMapping[fmt.Sprint(id)] = id
	}
}

// 在包被导入时，初始化协议包名字与 ID 的映射关系
func init() {
	initMCPacketNameIDMapping()
}

// stringWantsToIDSet 将字符串形式的协议包名字列表转换为对应的协议包 ID 集合
func stringWantsToIDSet(want []string) map[uint32]bool {
	s := map[uint32]bool{}
	for _, w := range want {
		// 如果字符串为 "any" 或 "all"，则将 s 中添加所有协议包的 ID
		if w == "any" || w == "all" {
			for _, id := range mcPacketNameIDMapping {
				s[id] = true
			}
			continue
		}
		add := true
		if strings.HasPrefix(w, "!") {
			add = false
			w = w[1:]
		}
		// 如果字符串以 "ID" 开头，则去掉 "ID" 前缀
		if strings.HasPrefix(w, "ID") {
			w = w[2:]
		}
		if id, found := mcPacketNameIDMapping[w]; found {
			if add {
				s[id] = true
			} else {
				delete(s, id)
			}
		}
	}
	return s
}

// should be no block
// PumperNoBlock 表示没有阻塞的数据处理函数类型
type PumperNoBlock func(pk packet.Packet) error

// GamePacketPumperMux 是一个 Minecraft 游戏协议包的多路复用器，用于将协议包分发给对应的数据处理函数
type GamePacketPumperMux struct {
	// 存储协议包 ID 到对应的数据处理函数集合的映射关系
	subPumpers map[uint32]*sync_wrapper.SyncMap[PumperNoBlock]
}

// NewGamePacketPumperMux 创建一个新的 GamePacketPumperMux 实例
func NewGamePacketPumperMux() *GamePacketPumperMux {
	if len(mcPacketNameIDMapping) == 0 {
		initMCPacketNameIDMapping()
	}
	pm := &GamePacketPumperMux{
		subPumpers: map[uint32]*sync_wrapper.SyncMap[PumperNoBlock]{},
	}
	// 将所有协议包 ID 与对应的数据处理函数集合的映射关系添加到 subPumpers 中
	for _, id := range mcPacketNameIDMapping {
		pm.subPumpers[id] = sync_wrapper.NewInstanceMap[PumperNoBlock]()
	}
	return pm
}

// translateStringWantsToIDSet 将字符串形式的协议包名字列表转换为对应的协议包 ID 集合
func (p *GamePacketPumperMux) translateStringWantsToIDSet(want []string) map[uint32]bool {
	return stringWantsToIDSet(want)
}

// GetMCPacketNameIDMapping 返回协议包名字与 ID 的映射关系
func (p *GamePacketPumperMux) GetMCPacketNameIDMapping() MCPacketNameIDMapping {
	return mcPacketNameIDMapping
}

// PumpGamePacket 将协议包分发给对应的数据处理函数进行处理
func (p *GamePacketPumperMux) PumpGamePacket(pk packet.Packet) {
	id := pk.ID()
	// 获取与协议包 ID 对应的数据处理函数集合
	if subPumper, found := p.subPumpers[id]; found {
		toRemove := []string{}
		// 遍历数据处理函数集合，依次调用其中的函数进行处理
		subPumper.Iter(func(k string, pumper PumperNoBlock) (continueIter bool) {
			err := pumper(pk)
			// 如果处理函数返回错误，则将其从集合中删除
			if err != nil {
				toRemove = append(toRemove, k)
			}
			return true
		})
		// 将需要删除的处理函数从集合中删除
		for _, k := range toRemove {
			subPumper.Delete(k)
		}
	}
}

// AddNewPumper 向 GamePacketPumperMux 中添加一个新的数据处理函数
func (p *GamePacketPumperMux) AddNewPumper(want []string, pumper PumperNoBlock) {
	// 将字符串形式的协议包名字列表转换为对应的协议包 ID 集合
	idSet := p.translateStringWantsToIDSet(want)
	for id := range idSet {
		// 将数据处理函数添加到与协议包 ID 对应的集合中
		if subPumpers, found := p.subPumpers[id]; found {
			subPumpers.Set(uuid.New().String(), pumper)
		}
	}
}

// MakeMCPacketNoBlockFeeder 创建一个 PumperNoBlock 类型的数据处理函数，用于将协议包发送到指定的 channel 中
func MakeMCPacketNoBlockFeeder(ctx context.Context, pkChan chan packet.Packet) PumperNoBlock {
	pumper := func(pk packet.Packet) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// 将协议包发送到指定的 channel 中
		select {
		case pkChan <- pk:
		default:
		}
		return nil
	}
	return pumper
}

type PacketDispatcher struct {
	packetQueueSize int
	pumperMux       *GamePacketPumperMux
}

func NewPacketDispatcher(packetQueueSize int, mux *GamePacketPumperMux) *PacketDispatcher {
	return &PacketDispatcher{
		packetQueueSize: packetQueueSize,
		pumperMux:       mux,
	}
}

func (m *PacketDispatcher) MakeMCPacketFeeder(ctx context.Context, wants []string) <-chan packet.Packet {
	feeder := make(chan packet.Packet, m.packetQueueSize)
	pumper := MakeMCPacketNoBlockFeeder(ctx, feeder)
	m.pumperMux.AddNewPumper(wants, pumper)
	return feeder
}

// GetMCPacketNameIDMapping 返回游戏包的名称和 ID 的映射
func (m *PacketDispatcher) GetMCPacketNameIDMapping() MCPacketNameIDMapping {
	return m.pumperMux.GetMCPacketNameIDMapping()
}
