package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"phoenixbuilder/minecraft/nbt"
	"reflect"
)

var ErrImportFormateNotSupport = errors.New("formate not support")

type SchemFileStructrue struct {
	Palette       map[string]int32
	Metadata      map[string]interface{}
	DataVersion   int32
	BlockDataIn   interface{} `nbt:"BlockData"`
	OffsetIn      interface{} `nbt:"Offset"`
	blockData     []uint32
	offset        []int
	PaletteMax    int32
	Version       int32
	Length        int16
	Height        int16
	Width         int16
	BlockEntities []map[string]interface{}
}

func DecodeSchem(data []byte) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unknown error %v", r)
		}
	}()
	err = ErrImportFormateNotSupport
	hasGzipHeader := false
	if len(data) < 2 {
		return err
	}
	if data[0] == 0x1f && data[1] == 0x8b {
		hasGzipHeader = true
	}
	var dataFeeder io.Reader
	if !hasGzipHeader {
		dataFeeder = bytes.NewBuffer(data)
	} else {
		dataFeeder, err = gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return err
		}
	}
	nbtDecoder := nbt.NewDecoderWithEncoding(dataFeeder, nbt.BigEndian)
	var schemData SchemFileStructrue
	err = nbtDecoder.Decode(&schemData)
	if err != nil {
		return ErrImportFormateNotSupport
	}

	convertBlockData := reflect.ValueOf(schemData.BlockDataIn)
	dataLen := convertBlockData.Len()
	// convertBlockData := reflect.ValueOf(schemData.BlockDataIn).Interface().(*[dataLen]interface{}) //.Slice(0, 1)
	// tc := reflect.ChanOf(reflect.SendDir, ta)
	// convertBlockData := reflect.SliceOf(schemData.BlockDataIn)
	// fmt.Println(convertBlockData)
	// if !ok {
	// 	return fmt.Errorf("cannot convert block data, please contact 2401PT")
	// }
	schemData.blockData = make([]uint32, dataLen)
	for i := 0; i < dataLen; i++ {
		rv := convertBlockData.Index(i)
		if !rv.CanUint() {
			return fmt.Errorf("cannot convert block data %v to uint, please contact 2401PT", rv)
		}
		schemData.blockData[i] = uint32(rv.Uint())
	}
	fmt.Println(schemData.blockData)
	// for k, _ := range schemData {
	// 	fmt.Println(k)
	// }
	// m, err := json.Marshal(schemData)

	// if err != nil {
	// 	return err
	// }
	// print(string(m))
	return nil
}

func main() {
	fp, err := os.OpenFile("/home/dai/projects/PhoenixBuilder/cli/schem_support_test/zsjtest.schem", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		panic(err)
	}
	err = DecodeSchem(data)
	fmt.Println(err)
	fp.Close()
}
