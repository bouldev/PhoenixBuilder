package step2_add_standard_mc_converts

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

func ReadSnbtFile(filePath string) (snbtInOut [][4]string) {
	fp, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	dataBytes, err := io.ReadAll(fp)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(bytes.NewBuffer(dataBytes))
	snbtInOut = [][4]string{}
	inBlockSNBT := ""
	for {
		cmd, err := reader.ReadString(':')
		if err != nil {
			break
		}
		cmd = strings.ReplaceAll(cmd, ":", " ")
		cmd = strings.TrimSpace(cmd)
		reader.ReadString(' ')
		snbt, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		snbt = strings.TrimSpace(snbt)
		if cmd == "in" {
			inBlockSNBT = snbt
		}
		if cmd == "out" {
			_1, _2 := snbtToNameAndState(inBlockSNBT)
			_3, _4 := snbtToNameAndState(snbt)
			snbtInOut = append(snbtInOut, [4]string{_1, _2, _3, _4})
			inBlockSNBT = ""
		}
	}
	return snbtInOut
}

func snbtToNameAndState(snbt string) (name, stateSnbt string) {
	outSS := strings.Split(snbt, "[")
	outBlockName, outBlockState := outSS[0], ""
	if len(outSS) > 1 {
		outBlockState = outSS[1]
	}
	outBlockState = strings.TrimSuffix(outBlockState, "]")
	outBlockState = strings.TrimSpace(outBlockState)
	outBlockName = strings.TrimPrefix(outBlockName, "minecraft:")
	outBlockName = strings.TrimSpace(outBlockName)
	return outBlockName, outBlockState
}
