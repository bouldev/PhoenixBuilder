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

// Default Package, which used to
// describe the unsupported parts
type Default struct {
	PACKAGE_NAME string // The name of this package
	Module              // The module contained in this package
}

// Return the package name of d
func (d *Default) PackageName() string {
	return d.PACKAGE_NAME
}

// Return a pool/map that contains all the module of m
func (d *Default) ModulePool() map[string]Module {
	return map[string]Module{}
}

// Init the module data from pool
func (d *Default) InitModuleFromPool(
	module_name string,
	pool map[string]Module,
) {
	module, ok := pool[module_name]
	if !ok {
		d.Module = &DefaultModule{
			MODULE_NAME: module_name,
			Event:       &DefaultEvent{},
		}
		return
	}
	d.Module = module
}
