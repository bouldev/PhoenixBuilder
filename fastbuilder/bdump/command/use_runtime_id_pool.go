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

type UseRuntimeIDPool struct {
	PoolID uint8
}

func (_ *UseRuntimeIDPool) ID() uint16 {
	return 31
}

func (_ *UseRuntimeIDPool) Name() string {
	return "UseRuntimeIDPoolCommand"
}

func (cmd *UseRuntimeIDPool) Marshal(writer io.Writer) error {
	_, err := writer.Write([]byte{cmd.PoolID})
	return err
}

func (cmd *UseRuntimeIDPool) Unmarshal(reader io.Reader) error {
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(reader, buf, 1)
	if err != nil {
		return err
	}
	cmd.PoolID = buf[0]
	return nil
}
