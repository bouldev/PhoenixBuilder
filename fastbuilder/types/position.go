package types

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

type Position struct {
	X, Y, Z int
}

func (p Position) FromInt(arr []int) {
	p.X = arr[0]
	p.Y = arr[1]
	p.Z = arr[2]
}

type FloatPosition struct {
	X, Y, Z float64
}

func (p *FloatPosition) TransferInt() Position {
	return Position{
		int(p.X),
		int(p.Y),
		int(p.Z),
	}
}
