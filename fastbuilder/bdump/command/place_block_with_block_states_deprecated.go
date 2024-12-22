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

type PlaceBlockWithBlockStatesDeprecated struct {
	BlockConstantStringID uint16
	BlockStatesString     string
}

func (_ *PlaceBlockWithBlockStatesDeprecated) ID() uint16 {
	return 13
}

func (_ *PlaceBlockWithBlockStatesDeprecated) Name() string {
	return "PlaceBlockWithBlockStatesDeprecatedCommand"
}

func (cmd *PlaceBlockWithBlockStatesDeprecated) Marshal(writer io.Writer) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.BlockConstantStringID)
	_, err := writer.Write(buf)
	if err != nil {
		return err
	}
	_, err = writer.Write(append([]byte(cmd.BlockStatesString), 0))
	return err
}

func (cmd *PlaceBlockWithBlockStatesDeprecated) Unmarshal(reader io.Reader) error {
	buf := make([]byte, 2)
	_, err := io.ReadAtLeast(reader, buf, 2)
	if err != nil {
		return err
	}
	cmd.BlockConstantStringID = binary.BigEndian.Uint16(buf)
	blockStates, err := readString(reader)
	if err != nil {
		return err
	}
	cmd.BlockStatesString = blockStates
	return nil
}
