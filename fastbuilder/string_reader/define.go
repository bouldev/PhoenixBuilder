package string_reader

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

// 描述一个字符串阅读器
type StringReader struct {
	ptr int     // 指代当前的阅读进度
	ctx *string // 指代该阅读器所包含的底层字符串
}

// 返回以 content 为底层的字符串阅读器
func NewStringReader(content *string) *StringReader {
	reader := StringReader{}
	reader.Reset(content)
	return &reader
}
