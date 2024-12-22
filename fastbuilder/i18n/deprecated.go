package I18n

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

// 此表标记了已弃用的语言
var DeprecatedLanguages []string = []string{}

// 确定代号为 langeuage_name 的语言是否被弃用。
// 若弃用，返回真，否则返回假
func IsDeprecated(language_name string) (has bool) {
	for _, value := range DeprecatedLanguages {
		if value == language_name {
			return true
		}
	}
	return false
}
