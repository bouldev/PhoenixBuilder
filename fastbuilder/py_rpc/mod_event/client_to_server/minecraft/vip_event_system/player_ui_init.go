package vip_event_system

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
type PlayerUiInit struct {
	RuntimeID string
}

// Return the event name of p
func (p *PlayerUiInit) EventName() string {
	return "PlayerUiInit"
}

// Convert p to go object which only contains go-built-in types
func (p *PlayerUiInit) MakeGo() (res any) {
	return p.RuntimeID
}

// Sync data to p from obj
func (p *PlayerUiInit) FromGo(obj any) error {
	object, success := obj.(string)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to string; obj = %#v", obj)
	}
	p.RuntimeID = object
	// convert and sync data
	return nil
	// return
}
