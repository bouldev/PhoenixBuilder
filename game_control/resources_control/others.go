package ResourcesControl

import "sync/atomic"

// ------------------------- currentTick -------------------------

// 以原子操作获取当前的游戏刻
func (o *others) GetCurrentTick() int64 {
	return atomic.LoadInt64(&o.currentTick)
}

// 以原子操作写入当前的游戏刻 currentTick 。
// 属于私有实现
func (o *others) writeCurrentTick(currentTick int64) {
	atomic.StoreInt64(&o.currentTick, currentTick)
}

// ------------------------- END -------------------------
