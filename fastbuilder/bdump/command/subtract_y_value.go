package command

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
	"io"
)

type SubtractYValue struct{}

func (_ *SubtractYValue) ID() uint16 {
	return 17
}

func (_ *SubtractYValue) Name() string {
	return "SubtractYValueCommand"
}

func (_ *SubtractYValue) Marshal(_ io.Writer) error {
	return nil
}

func (_ *SubtractYValue) Unmarshal(_ io.Reader) error {
	return nil
}
