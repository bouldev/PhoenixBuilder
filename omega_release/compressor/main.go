package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
)

func GetFileData(fname string) ([]byte, error) {
	fp, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	return buf, err
}

func WriteFileData(fname string, data []byte) error {
	fp, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer fp.Close()
	if _, err := fp.Write(data); err != nil {
		return err
	}
	return nil
}

func CompressSingleFile(in, out string) {
	var origData []byte
	var err error
	origData, err = GetFileData(in)
	if err != nil || len(origData) == 0 {
		panic(fmt.Sprintf("read %v fail", in))
	}

	buf := bytes.NewBuffer([]byte{})
	compressor := brotli.NewWriterLevel(buf, brotli.DefaultCompression)
	compressor.Write(origData)
	compressor.Close()
	newData := buf.Bytes()

	if err := WriteFileData(out, newData); err != nil {
		panic(err)
	}
	fmt.Printf("comprerss: %v -> %v  compress %.3f\n", in, out, float32(len(newData))/float32(len(origData)))
}

func main() {
	inFile := flag.String("in", "", "input")
	outFile := flag.String("out", "", "outfile")
	flag.Parse()
	if strings.Contains(*inFile, ",") {
		ins := strings.Split(*inFile, ",")
		outs := strings.Split(*outFile, ",")
		if len(ins) != len(outs) {
			panic(fmt.Errorf("%v!->%v :input/outputs mismatch", ins, outs))
		}
		var wg sync.WaitGroup
		for i := range ins {
			wg.Add(1)
			in, out := strings.TrimSpace(ins[i]), strings.TrimSpace(outs[i])
			go func() {
				CompressSingleFile(in, out)
				wg.Done()
			}()
		}
		wg.Wait()
	} else {
		CompressSingleFile(*inFile, *outFile)
	}

	// var compressedData []byte
	// if compressedData, err = GetFileData(*outFile); err != nil {
	// 	panic(err)
	// }

	// if recoveredData, err := ioutil.ReadAll(brotli.NewReader(bytes.NewReader(compressedData))); err != nil {
	// 	panic(err)
	// } else {
	// 	if bytes.Compare(recoveredData, origData) != 0 {
	// 		panic("not same")
	// 	}
	// 	fmt.Println("Success")
	// }
}
