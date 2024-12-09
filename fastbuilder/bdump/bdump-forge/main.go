package main

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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"phoenixbuilder/fastbuilder/args"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("BDumpForge built with PhoenixBuilder %s\n", args.FBVersion)
		fmt.Printf("%s <input (bdx/json)> <output>\n", os.Args[0])
		os.Exit(1)
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to open input file: %v\n", err)
		os.Exit(2)
	}
	output_file, err := os.OpenFile(os.Args[2], os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer output_file.Close()
	defer file.Close()
	header_byte := []byte{0}
	_, err = io.ReadAtLeast(file, header_byte, 1)
	if err != nil {
		panic(err)
	}
	if header_byte[0] == '{' {
		bf := &bytes.Buffer{}
		_, err = bf.ReadFrom(file)
		if err != nil {
			panic(err)
		}
		json_str := string(append([]byte{'{'}, bf.Bytes()...))
		gvmap := map[string]interface{}{}
		err = json.Unmarshal([]byte(json_str), &gvmap)
		if err != nil {
			panic(err)
		}
		construct(gvmap, output_file)
	} else if header_byte[0] == 'B' {
		header_byte = make([]byte, 2)
		_, err = io.ReadAtLeast(file, header_byte, 2)
		if err != nil {
			panic(err)
		}
		if string(header_byte) != "D@" {
			fmt.Printf("Not a Brotli-Compressed BDump file.\n")
			os.Exit(3)
		}
		extract(file, output_file)
	}
	fmt.Printf("Invalid input file\n")
	os.Exit(3)
}
