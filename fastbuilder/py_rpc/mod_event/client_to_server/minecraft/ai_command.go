package minecraft

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
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/client_to_server/minecraft/ai_command"
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
)

// 魔法指令
type AICommand struct{ mei.Module }

// Return the module name of a
func (a *AICommand) ModuleName() string {
	return "aiCommand"
}

// Return a pool/map that contains all the event of a
func (a *AICommand) EventPool() map[string]mei.Event {
	return map[string]mei.Event{
		"ExecuteCommandEvent": &ai_command.ExecuteCommandEvent{},
	}
}
