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

// Default content, which used to
// describe the unsupported parts
type Default struct {
	NAME string // The name of this content
	Data any    // The data contained in this content
}

// Return the name of d
func (d *Default) Name() string {
	return d.NAME
}

// Convert d to go object which only contains go-built-in types
func (d *Default) MakeGo() (res any) {
	return d.Data
}

// Sync data to d from obj
func (d *Default) FromGo(res any) error {
	d.Data = res
	return nil
}
