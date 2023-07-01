package main

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
