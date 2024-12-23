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

import "fmt"

// Describe the type of PyRpc/StartType packet
const (
	StartTypeRequest  = "GetStartType"
	StartTypeResponse = "SetStartType"
)

// 机器人进服后的挑战之一
type StartType struct {
	Content string
	Type    string
}

// Return the name of g
func (s *StartType) Name() string {
	return s.Type
}

// Convert g to go object which only contains go-built-in types
func (s *StartType) MakeGo() (res any) {
	return []any{s.Content}
}

// Sync data to g from obj
func (s *StartType) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	if len(object) != 1 {
		return fmt.Errorf("FromGo: The length of object is not equal to 1; object = %#v", object)
	}
	// convert and check data
	s.Content, success = object[0].(string)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert object[0] to string; object[0] = %#v", object[0])
	}
	// get and sync data
	return nil
	// return
}
