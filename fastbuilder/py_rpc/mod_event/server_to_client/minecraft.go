package mod_event_server_to_client

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
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client/minecraft"
)

// Minecraft Package
type Minecraft struct{ mei.Default }

// Return the package name of m
func (m *Minecraft) PackageName() string {
	return "Minecraft"
}

// Return a pool/map that contains all the module of m
func (m *Minecraft) ModulePool() map[string]mei.Module {
	return map[string]mei.Module{
		"aiCommand":     &minecraft.AICommand{Module: &mei.DefaultModule{}},
		"pet":           &minecraft.Pet{Module: &mei.DefaultModule{}},
		"chatPhrases":   &minecraft.ChatPhrases{Module: &mei.DefaultModule{}},
		"achievement":   &minecraft.Achievement{Module: &mei.DefaultModule{}},
		"chatExtension": &minecraft.ChatExtension{Module: &mei.DefaultModule{}},
	}
}
