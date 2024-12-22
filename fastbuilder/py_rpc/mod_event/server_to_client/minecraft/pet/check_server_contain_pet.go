package pet

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

type CheckServerContainPet struct{}

// Return the event name of c
func (c *CheckServerContainPet) EventName() string {
	return "check_server_contain_pet"
}

// Convert c to go object which only contains go-built-in types
func (c *CheckServerContainPet) MakeGo() (res any) {
	return map[string]any{}
}

// Sync data to c from obj
func (c *CheckServerContainPet) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	if len(object) != 0 {
		return fmt.Errorf("FromGo: The length of object is not equal to 0; object = %#v", object)
	}
	// check data
	*c = CheckServerContainPet{}
	// sync data
	return nil
	// return
}
