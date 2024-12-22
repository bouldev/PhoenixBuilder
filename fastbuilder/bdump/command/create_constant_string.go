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

type CreateConstantString struct {
	ConstantString string
}

func (_ *CreateConstantString) ID() uint16 {
	return 1
}

func (_ *CreateConstantString) Name() string {
	return "CreateConstantStringCommand"
}

func (cmd *CreateConstantString) Marshal(writer io.Writer) error {
	str_c := append([]byte(cmd.ConstantString), 0)
	_, err := writer.Write(str_c)
	return err
}

func (cmd *CreateConstantString) Unmarshal(reader io.Reader) error {
	singlebuf := make([]byte, 1)
	cmd.ConstantString = ""
	for {
		_, err := io.ReadAtLeast(reader, singlebuf, 1)
		if err != nil {
			return err
		}
		if singlebuf[0] == 0 {
			break
		}
		cmd.ConstantString += string(singlebuf[0])
	}
	return nil
}
