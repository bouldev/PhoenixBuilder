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
func (p *packetListener) CreateNewListen(
	packetsID []uint32,
	upperStorageLimit int16,
) (uuid.UUID, <-chan packet.Packet) {
	uniqueId := GenerateUUID()
	ctx, stop := context.WithCancel(context.Background())
	newListen := singleListen{
		packetsID:      packetsID,
		packetReceived: make(chan packet.Packet, upperStorageLimit),
		ctx:            ctx,
		stop:           stop,
	}
	p.listenerWithData.Store(uniqueId, newListen)
	return uniqueId, newListen.packetReceived
}

// 将数据包 pk 发送到管道 s.packetReceived 。
// 此函数可能会被阻塞，因此需要以协程执行。
// 如果 s 所对应的监听已被它的监听者中止，
// 那么此函数将会返回值，无论其是否已被阻塞。
// 属于私有实现
func (s *singleListen) simplePacketDistributor(
	pk packet.Packet,
) {
	if atomic.LoadInt32(&s.runningCounts) >= MaximumCoroutinesRunningCount {
		return
	}
	// 如果该监听器下已运行的协程数超过了最大允许数量，
	// 则丢当前数据包，直接返回值
	atomic.AddInt32(&s.runningCounts, 1)
	defer atomic.AddInt32(&s.runningCounts, -1)
	// 更新该监听器下已运行的协程数
	select {
	case <-s.ctx.Done():
		// 如果监听器已被它的监听者终止并关闭，
		// 那么本协程需要立即销毁
	case s.packetReceived <- pk:
		// 将数据包发送到管道，
		// 将在管道缓冲区已满时遭遇阻塞
	}
	// 分发数据包
}

// 将数据包 pk 分发到每个监听器上。
// 属于私有实现
func (p *packetListener) distributePacket(pk packet.Packet) error {
	var err error
	// 初始化
	p.listenerWithData.Range(
		func(key, value any) bool {
			singleListen, success := value.(singleListen)
			if !success {
				err = fmt.Errorf("distributePacket: Failed to convert value into singleListen; value = %#v", value)
				return false
			}
			// 转换数据类型
			if len(singleListen.packetsID) == 0 {
				go singleListen.simplePacketDistributor(pk)
				return true
			}
			// 如果要监听所有的数据包
			for _, val := range singleListen.packetsID {
				if val == pk.ID() {
					go singleListen.simplePacketDistributor(pk)
				}
			}
			return true
			// 如果只监听特定的数据包
		},
	)
	// 分发数据包到每个监听器上
	if err != nil {
		return err
	}
	return nil
	// 返回值
}

// 终止并关闭 listener 所指代的监听器
func (p *packetListener) StopAndDestroy(listener uuid.UUID) error {
	single_listen_origin, ok := p.listenerWithData.Load(listener)
	if !ok {
		return fmt.Errorf("StopAndDestroy: %v is not recorded", listener.String())
	}
	singleListen, success := single_listen_origin.(singleListen)
	if !success {
		return fmt.Errorf("StopAndDestroy: Failed to convert single_listen_origin into singleListen; single_listen_origin = %#v", single_listen_origin)
	}
	// convert data into known data type
	singleListen.stop()
	p.listenerWithData.Delete(listener)
	// send stop command and delete listener
	return nil
	// return
}
