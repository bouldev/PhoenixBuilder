package py_rpc

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
	"encoding/json"
	"fmt"
)

// Convert py_rpc to go object which
// only contains go-built-in types
func Marshal(py_rpc PyRpc) (res any) {
	return []any{
		py_rpc.Name(),
		py_rpc.MakeGo(),
		nil,
	}
}

// Convert obj to PyRpc
func Unmarshal(obj any) (py_rpc PyRpc, err error) {
	object, success := obj.([]any)
	if !success {
		return nil, fmt.Errorf("Unmarshal: Unsupported PyRpc packet; obj = %#v", obj)
	}
	if len(object) < 2 {
		return nil, fmt.Errorf("Unmarshal: The length of object is less than 2; object = %#v", object)
	}
	// check data
	name, success := object[0].(string)
	if !success {
		return nil, fmt.Errorf("Unmarshal: Failed to convert object[0] to string; object[0] = %#v", object[0])
	}
	// get name
	var ok bool
	py_rpc, ok = Pool()[name]
	if !ok {
		py_rpc = &Default{NAME: name}
	}
	// get content of initial state
	err = py_rpc.FromGo(object[1])
	if err != nil {
		err = fmt.Errorf("Unmarshal: %v", err)
	}
	return
	// put data and return
}

// Output the object as json
func Debug(object any) (res string, err error) {
	str, err := json.MarshalIndent(object, "", "\t")
	if err != nil {
		err = fmt.Errorf("Debug: %v", err)
	}
	return string(str), err
}
