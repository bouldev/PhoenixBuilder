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

type AddInt8ZValue struct {
	Value int8
}

func (_ *AddInt8ZValue) ID() uint16 {
	return 30
}

func (_ *AddInt8ZValue) Name() string {
	return "AddInt8ZValueCommand"
}

func (cmd *AddInt8ZValue) Marshal(writer io.Writer) error {
	buf := []byte{uint8(cmd.Value)}
	_, err := writer.Write(buf)
	return err
}

func (cmd *AddInt8ZValue) Unmarshal(reader io.Reader) error {
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(reader, buf, 1)
	if err != nil {
		return err
	}
	cmd.Value = int8(buf[0])
	return nil
}
