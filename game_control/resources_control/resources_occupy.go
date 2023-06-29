package ResourcesControl

import (
	"sync"

	"github.com/google/uuid"
)

// 描述一个通用的客户端资源独占结构
type resourcesOccupy struct {
	// 用于阻塞其他请求该资源的互斥锁
	lockDown sync.Mutex
	// 标识资源的占用者，为 UUID 的字符串形式
	holder string
}

/*
占用客户端的某个资源。

返回的字符串指代资源的占用者，为 UUID 的字符串形式，这用于资源释放函数
func (r *resourcesOccupy) Release(holder string) bool 中的 holder 参数
*/
func (r *resourcesOccupy) Occupy() string {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return r.Occupy()
	}
	uniqueId := newUUID.String()
	// get new unique id
	r.lockDown.Lock()
	// lock down resources
	r.holder = uniqueId
	// set the holder of this resources
	return uniqueId
	// return
}

// 释放客户端的某个资源，返回值代表执行结果。
// holder 指代该资源的占用者，当且仅当填写的占用者
// 可以与内部记录的占用者对应时才可以成功释放该资源
func (r *resourcesOccupy) Release(holder string) bool {
	if r.holder != holder || r.holder == "" {
		return false
	}
	// verify the holder
	r.holder = ""
	// clear the holder of this resources
	r.lockDown.Unlock()
	// unlock resources
	return true
	// return
}

// 返回资源的占用状态，为真时代表已被占用，否则反之
func (r *resourcesOccupy) GetOccupyStates() bool {
	notOccupied := r.lockDown.TryLock()
	if notOccupied {
		r.lockDown.Unlock()
	}
	return !notOccupied
}
