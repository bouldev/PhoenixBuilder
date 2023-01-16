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
	in = strings.TrimSpace(in)
	out = strings.TrimSpace(out)
	origData, err = GetFileData(in)
	if err != nil {
		panic(fmt.Sprintf("read %v fail: %v", in, err))
	} else if len(origData) == 0 {
		panic(fmt.Sprintf("read %v fail: data length = 0", in))
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
	_inFile := flag.String("in", "", "input")
	_outFile := flag.String("out", "", "outfile")
	flag.Parse()
	inFile := strings.TrimSpace(*_inFile)
	outFile := strings.TrimSpace(*_outFile)
	fmt.Println(inFile, outFile)
	if strings.Contains(inFile, ",") {
		ins := strings.Split(inFile, ",")
		outs := strings.Split(outFile, ",")
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
		CompressSingleFile(inFile, outFile)
	}
}
