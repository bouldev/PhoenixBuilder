package ResourcesControl

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

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
