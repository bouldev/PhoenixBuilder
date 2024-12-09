package mod_event_interface

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

// Default Module, which used to
// describe the unsupported parts
type DefaultModule struct {
	MODULE_NAME string // The name of this module
	Event              // The event contained in this module
}

// Return the module name of d
func (d *DefaultModule) ModuleName() string {
	return d.MODULE_NAME
}

// Return a pool/map that contains all the event of d
func (d *DefaultModule) EventPool() map[string]Event {
	return map[string]Event{}
}

// Init the event data for d from pool
func (d *DefaultModule) InitEventFromPool(
	event_name string,
	pool map[string]Event,
) {
	event, ok := pool[event_name]
	if !ok {
		d.Event = &DefaultEvent{EVENT_NAME: event_name}
		return
	}
	d.Event = event
}
