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
	"encoding/binary"
	"io"
)

type AssignDebugData struct {
	Data []byte
}

func (_ *AssignDebugData) ID() uint16 {
	return 39
}

func (_ *AssignDebugData) Name() string {
	return "AssignDebugDataCommand"
}

func (cmd *AssignDebugData) Marshal(writer io.Writer) error {
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(cmd.Data)))
	_, err := writer.Write(append(lenBuf, cmd.Data...))
	return err
}

func (cmd *AssignDebugData) Unmarshal(reader io.Reader) error {
	lenBuf := make([]byte, 4)
	_, err := io.ReadAtLeast(reader, lenBuf, 4)
	if err != nil {
		return err
	}
	cmd.Data = make([]byte, int(binary.BigEndian.Uint32(lenBuf)))
	_, err = io.ReadAtLeast(reader, cmd.Data, len(cmd.Data))
	return err
}
