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

// Used on initialize the
// Minecraft connection
type ClientLoadAddonsFinishedFromGac []any

// Return the name of s
func (s *ClientLoadAddonsFinishedFromGac) Name() string {
	return "ClientLoadAddonsFinishedFromGac"
}

// Convert s to go object which only contains go-built-in types
func (s *ClientLoadAddonsFinishedFromGac) MakeGo() (res any) {
	if s == nil {
		return nil
	}
	return []any(*s)
}

// Sync data to s from obj
func (s *ClientLoadAddonsFinishedFromGac) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	*s = object
	// convert and sync data
	return nil
	// return
}
