package ResourcesControl

import (
	"sync"

	"github.com/google/uuid"
)

// 描述一个通用的客户端资源独占结构
type resources_occupy struct {
	// 用于阻塞其他请求该资源的互斥锁
	lock_down sync.Mutex
	// 标识资源的占用者，为 UUID 的字符串形式
	holder string
}

/*
占用客户端的某个资源。

返回的字符串指代资源的占用者，
为 UUID 的字符串形式，
这用于资源释放函数中的 holder 参数
*/
func (r *resources_occupy) Occupy() string {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return r.Occupy()
	}
	uniqueId := newUUID.String()
	// get new unique id
	r.lock_down.Lock()
	// lock down resources
	r.holder = uniqueId
	// set the holder of this resources
	return uniqueId
	// return
}

// 释放客户端的某个资源，返回值代表执行结果。
// holder 指代该资源的占用者，当且仅当填写的占用者
// 可以与内部记录的占用者对应时才可以成功释放该资源
func (r *resources_occupy) Release(holder string) bool {
	if r.holder != holder || r.holder == "" {
		return false
	}
	// verify the holder
	r.holder = ""
	// clear the holder of this resources
	r.lock_down.Unlock()
	// unlock resources
	return true
	// return
}

// 返回资源的占用状态，为真时代表已被占用，否则反之
func (r *resources_occupy) GetOccupyStates() bool {
	not_occupied := r.lock_down.TryLock()
	if not_occupied {
		r.lock_down.Unlock()
	}
	return !not_occupied
}
