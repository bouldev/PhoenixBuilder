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
	"fmt"
	"io"
)

type Command interface {
	ID() uint16 // Extra ID spaces (uint16) may be allocated in the future.
	Name() string
	Marshal(writer io.Writer) error
	Unmarshal(reader io.Reader) error
}

// Some deprecated commands may not be placed in this directory
// as I think we do not have to make them work

func readString(reader io.Reader) (string, error) {
	full_buf := []byte{}
	buf := make([]byte, 1)
	for {
		_, err := io.ReadAtLeast(reader, buf, 1)
		if err != nil {
			return "", err
		}
		if buf[0] == 0 {
			return string(full_buf), nil
		}
		full_buf = append(full_buf, buf...)
	}
	// This should not happen
	return string(full_buf), nil
}

func ReadCommand(reader io.Reader) (Command, error) {
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(reader, buf, 1)
	if err != nil {
		return nil, err
	}
	command_func, found_command := BDumpCommandPool[uint16(buf[0])]
	if !found_command {
		return nil, fmt.Errorf("Command::ReadCommand: Unknown Command ID: %d", int(buf[0]))
	}
	command := command_func()
	err = command.Unmarshal(reader)
	if err != nil {
		return nil, err
	}
	return command, nil
}

func WriteCommand(command Command, writer io.Writer) error {
	_, err := writer.Write([]byte{uint8(command.ID())})
	if err != nil {
		return err
	}
	return command.Marshal(writer)
}
