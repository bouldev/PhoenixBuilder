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

import "io"

type NoOperation struct{}

func (_ *NoOperation) ID() uint16 {
	return 9
}

func (_ *NoOperation) Name() string {
	return "NoOperationCommand"
}

func (_ *NoOperation) Marshal(_ io.Writer) error {
	return nil
}

func (_ *NoOperation) Unmarshal(_ io.Reader) error {
	return nil
}
