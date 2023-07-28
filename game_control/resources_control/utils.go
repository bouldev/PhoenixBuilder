package ResourcesControl

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/google/uuid"
)

// 生成一个新的 uuid 对象并返回
func GenerateUUID() uuid.UUID {
	for {
		uniqueId, err := uuid.NewUUID()
		if err != nil {
			continue
		}
		return uniqueId
	}
}

// 将 source 深拷贝到 destination 。
//
// register 是一个用于注册接口型变量的回调函数，
// 它将由 DeepCopy 自身执行。以下是一个例子。
//
//	func() {
//		gob.Register(map[string]any{})
//	}
func DeepCopy(
	source interface{},
	destination interface{},
	register func(),
) error {
	register()
	var buffer bytes.Buffer
	// init values
	err := gob.NewEncoder(&buffer).Encode(source)
	if err != nil {
		return fmt.Errorf("DeepCopy: %v", err)
	}
	// encode
	err = gob.NewDecoder(&buffer).Decode(destination)
	if err != nil {
		return fmt.Errorf("DeepCopy: %v", err)
	}
	// decode
	return nil
	// return
}
