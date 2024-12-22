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

type PlaceRuntimeBlockWithUint32RuntimeID struct {
	BlockRuntimeID uint32
}

func (_ *PlaceRuntimeBlockWithUint32RuntimeID) ID() uint16 {
	return 33
}

func (_ *PlaceRuntimeBlockWithUint32RuntimeID) Name() string {
	return "PlaceRuntimeBlockUint32RuntimeIDCommand"
}

func (cmd *PlaceRuntimeBlockWithUint32RuntimeID) Marshal(writer io.Writer) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, cmd.BlockRuntimeID)
	_, err := writer.Write(buf)
	return err
}

func (cmd *PlaceRuntimeBlockWithUint32RuntimeID) Unmarshal(reader io.Reader) error {
	buf := make([]byte, 4)
	_, err := io.ReadAtLeast(reader, buf, 4)
	if err != nil {
		return err
	}
	cmd.BlockRuntimeID = binary.BigEndian.Uint32(buf)
	return nil
}
