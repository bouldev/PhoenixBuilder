package bdump

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
	"hash"
	"io"
	"phoenixbuilder/fastbuilder/bdump/command"
)

type BDumpWriter struct {
	writer io.Writer
}

func (w *BDumpWriter) WriteCommand(cmd command.Command) error {
	return command.WriteCommand(cmd, w.writer)
}

type HashedWriter struct {
	writer io.Writer
	hash   hash.Hash
}

func (w *HashedWriter) Write(p []byte) (n int, err error) {
	w.hash.Write(p)
	n, err = w.writer.Write(p)
	return
}
