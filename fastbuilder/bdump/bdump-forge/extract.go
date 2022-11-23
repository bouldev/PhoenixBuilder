package main

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"encoding/binary"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/bdump/command"
	"github.com/andybalholm/brotli"
)

func readBrString(br *bytes.Buffer) (string, error) {
	str := ""
	c := make([]byte, 1)
	for {
		_, err := br.Read(c)
		if err != nil {
			return "", err
		}
		if c[0] == 0 {
			break
		}
		str += string(c)
	}
	return str, nil
}

func extract(file io.Reader, output_file io.Writer) {
	output:=map[string]interface{} {}
	btl:=brotli.NewReader(file)
	br:=&bytes.Buffer{}
	filelen, err := br.ReadFrom(btl)
	if err!=nil {
		panic(err)
	}
	{
		bts := br.Bytes()
		if bts[filelen-1] == 90 {
			lent := int64(bts[filelen-2])
			var sign []byte
			var fileBody []byte
			if lent == int64(255) {
				lenBuf := bts[filelen-4 : filelen-2]
				lent = int64(binary.BigEndian.Uint16(lenBuf))
				sign = bts[filelen-lent-4 : filelen-4]
				fileBody = bts[:filelen-lent-5]
			} else {
				sign = bts[filelen-lent-2 : filelen-2]
				fileBody = bts[:filelen-lent-3]
			}
			cor, un, err := bdump.VerifyBDX(fileBody, sign)
			if err!=nil {
				output["signed"]=true
				output["signature"]=map[string]interface{} {
					"signature": hex.EncodeToString(sign),
					"signature_verification_error": fmt.Sprintf("%#v", err),
					"verified": false,
				}
			}else{
				signature_status:=map[string]interface{} {
					"signature": hex.EncodeToString(sign),
					"corrupted": cor,
					"signature_verification_error": "NULL",
					"verified": true,
					"signer": un,
				}
				output["signed"]=true
				output["signature"]=signature_status
			}
		}else{
			output["signed"]=false
		}
	}
	{
		tempbuf := make([]byte, 4)
		_, err := io.ReadAtLeast(br, tempbuf, 4)
		if err != nil {
			panic(err)
		}
		if string(tempbuf) != "BDX\x00" {
			fmt.Printf("Inner content is not under a valid BDX format.\n")
			os.Exit(3)
		}
	}
	readBrString(br)
	brushPosition := []int{0, 0, 0}
	bigJsonItem:=[]interface{}{}
	for {
		cmd, err:=command.ReadCommand(br)
		if err!=nil {
			panic(err)
		}
		cmdJsonItem:=map[string]interface{} {}
		cmdJsonItem["brush_position_before_execution"]=[]int{brushPosition[0],brushPosition[1],brushPosition[2]}
		cmdJsonItem["command_name"]=cmd.Name()
		cmdJsonItem["id"]=cmd.ID()
		cmdJsonItem["command"]=cmd
		bigJsonItem=append(bigJsonItem, cmdJsonItem)
		_, isTerminate:=cmd.(*command.Terminate)
		if isTerminate {
			break
		}
		switch dcmd:=cmd.(type) {
		case *command.AddInt16ZValue0:
			brushPosition[2] += int(dcmd.Value)
		case *command.AddZValue0:
			brushPosition[2]++
		case *command.AddInt32ZValue0:
			brushPosition[2] += int(dcmd.Value)
		case *command.AddXValue:
			brushPosition[0]++
		case *command.SubtractXValue:
			brushPosition[0]--
		case *command.AddYValue:
			brushPosition[1]++
		case *command.SubtractYValue:
			brushPosition[1]--
		case *command.AddZValue:
			brushPosition[2]++
		case *command.SubtractZValue:
			brushPosition[2]--
		case *command.AddInt16XValue:
			brushPosition[0] += int(dcmd.Value)
		case *command.AddInt32XValue:
			brushPosition[0] += int(dcmd.Value)
		case *command.AddInt16YValue:
			brushPosition[1] += int(dcmd.Value)
		case *command.AddInt32YValue:
			brushPosition[1] += int(dcmd.Value)
		case *command.AddInt16ZValue:
			brushPosition[2] += int(dcmd.Value)
		case *command.AddInt32ZValue:
			brushPosition[2] += int(dcmd.Value)
		case *command.AddInt8XValue:
			brushPosition[0] += int(dcmd.Value)
		case *command.AddInt8YValue:
			brushPosition[1] += int(dcmd.Value)
		case *command.AddInt8ZValue:
			brushPosition[2] += int(dcmd.Value)
		}
	}
	output["contents"]=bigJsonItem
	json_str, err:=json.MarshalIndent(output, "", "\t")
	if err!=nil {
		panic(err)
	}
	_, err=output_file.Write(json_str)
	if err!=nil {
		panic(err)
	}
	os.Exit(0)
}