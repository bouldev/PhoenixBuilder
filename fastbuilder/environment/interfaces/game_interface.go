package interfaces

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
	resources_control "phoenixbuilder/game_control/resources_control"
)

type GameInterface interface {
	SendSettingsCommand(string, bool) error

	SendCommand(string) error
	SendWSCommand(string) error
	SendAICommand(string) error

	SendCommandWithResponse(string, resources_control.CommandRequestOptions) resources_control.CommandRespond
	SendWSCommandWithResponse(string, resources_control.CommandRequestOptions) resources_control.CommandRespond
	SendAICommandWithResponse(string, resources_control.CommandRequestOptions) resources_control.CommandRespond

	SetBlock([3]int32, string, string) error
	SetBlockAsync([3]int32, string, string) error

	SendChat(string) error
	Output(string) error
	Title(string) error
}
