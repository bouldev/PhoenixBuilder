package ResourcesControl

import (
	"context"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync/atomic"

	"github.com/google/uuid"
)

/*
创建一个新的数据包监听器。

packetID 指代本次欲监听的数据包，
为空则代表监听所有数据包。

upperStorageLimit 代表缓冲区可保存的最大数据包数。

返回的 uuid.UUID 用于标识当前监听器，
而返回的管道则代表用于储存数据包的缓冲区，
它将被实时更新，直到被它的监听者关闭
*/
func (p *packet_listener) CreateNewListen(
	packets_id []uint32,
	upperStorageLimit int16,
) (uuid.UUID, <-chan packet.Packet) {
	uniqueId := GenerateUUID()
	ctx, stop := context.WithCancel(context.Background())
	newListen := single_listen{
		packets_id:      packets_id,
		packet_received: make(chan packet.Packet, upperStorageLimit),
		ctx:             ctx,
		stop:            stop,
	}
	p.listener_with_data.Store(uniqueId, newListen)
	return uniqueId, newListen.packet_received
}

// 将数据包 pk 发送到管道 s.packet_received 。
// 此函数可能会被阻塞，因此需要以协程执行。
// 如果 s 所对应的监听已被它的监听者中止，
// 那么此函数将会返回值，无论其是否已被阻塞。
// 属于私有实现
func (s *single_listen) simple_packet_distributor(
	pk packet.Packet,
) {
	if atomic.LoadInt32(&s.running_counts) >= MaximumCoroutinesRunningCount {
		return
	}
	// 如果该监听器下已运行的协程数超过了最大允许数量，
	// 则丢当前数据包，直接返回值
	atomic.AddInt32(&s.running_counts, 1)
	defer atomic.AddInt32(&s.running_counts, -1)
	// 更新该监听器下已运行的协程数
	select {
	case <-s.ctx.Done():
		// 如果监听器已被它的监听者终止并关闭，
		// 那么本协程需要立即销毁
	case s.packet_received <- pk:
		// 将数据包发送到管道，
		// 将在管道缓冲区已满时遭遇阻塞
	}
	// 分发数据包
}

// 将数据包 pk 分发到每个监听器上。
// 属于私有实现
func (p *packet_listener) distribute_packet(pk packet.Packet) {
	p.listener_with_data.Range(
		func(key uuid.UUID, value single_listen) bool {
			if len(value.packets_id) == 0 {
				go value.simple_packet_distributor(pk)
				return true
			}
			// 如果要监听所有的数据包
			for _, val := range value.packets_id {
				if val == pk.ID() {
					go value.simple_packet_distributor(pk)
				}
			}
			return true
			// 如果只监听特定的数据包
		},
	)
	// 分发数据包到每个监听器上
}

// 终止并关闭 listener 所指代的监听器
func (p *packet_listener) StopAndDestroy(listener uuid.UUID) error {
	single_listen, ok := p.listener_with_data.Load(listener)
	if !ok {
		return fmt.Errorf("StopAndDestroy: %v is not recorded", listener.String())
	}
	single_listen.stop()
	p.listener_with_data.Delete(listener)
	// send stop command and delete listener
	return nil
	// return
}
