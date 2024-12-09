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

type AddXValue struct{}

func (_ *AddXValue) ID() uint16 {
	return 14
}

func (_ *AddXValue) Name() string {
	return "AddXValueCommand"
}

func (_ *AddXValue) Marshal(_ io.Writer) error {
	return nil
}

func (_ *AddXValue) Unmarshal(_ io.Reader) error {
	return nil
}
