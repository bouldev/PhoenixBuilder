package main

import (
	"io"
	"os"
	"fmt"
	"encoding/json"
	"phoenixbuilder/fastbuilder/bdump/command"
	
	"github.com/andybalholm/brotli"
)

type bdumpWriter struct {
	writer io.Writer
}

func (w *bdumpWriter) WriteCommand(cmd command.Command) error {
	return command.WriteCommand(cmd, w.writer)
}

var bdumpCommandNameToCommandPool map[string]func()command.Command = map[string]func()command.Command {}

func init() {
	for _, f:=range command.BDumpCommandPool {
		tmpitm:=f()
		bdumpCommandNameToCommandPool[tmpitm.Name()]=f
	}
}

func construct(input map[string]interface{}, output_file io.Writer) {
	_, err:=output_file.Write([]byte("BD@"))
	if err!=nil {
		panic(err)
	}
	brw:=brotli.NewWriter(output_file)
	_, err=brw.Write(append([]byte("BDX"),[]byte{0,0}...))
	if err!=nil {
		panic(err)
	}
	writer:=&bdumpWriter{writer:brw}
	contents_arr:=input["contents"].([]interface{})
	for _, _v:=range contents_arr {
		v:=_v.(map[string]interface{})
		id_pex, has_id:=v["id"]
		name_pex, has_name:=v["name"]
		var cmd command.Command
		if has_id {
			id:=uint16(id_pex.(float64))
			cmd_f, found:=command.BDumpCommandPool[id]
			if !found {
				fmt.Printf("Fatal: Command with ID %d not found.\n", id)
				os.Exit(7)
			}
			cmd=cmd_f()
			if has_name {
				name:=name_pex.(string)
				if name!=cmd.Name() {
					fmt.Printf("Fatal: ID/Name pair mismatch: ID %d and Name %s (expected %s)\n", id, name, cmd.Name())
					os.Exit(6)
				}
			}
		}else if has_name {
			name:=name_pex.(string)
			cmd_f, found:=bdumpCommandNameToCommandPool[name]
			if !found {
				fmt.Printf("Fatal: Command with Name %s not found.\n", name)
				os.Exit(8)
			}
			cmd=cmd_f()
		}else{
			fmt.Printf("Fatal: NO COMMAND IDENTIFIER FOR COMMAND: %#v\n", v)
			os.Exit(9)
		}
		contents_if, found_cif:=v["command"]
		if found_cif {
			command_content_str, _:=json.Marshal(contents_if.(map[string]interface{}))
			json.Unmarshal(command_content_str, &cmd)
		}
		writer.WriteCommand(cmd)
	}
	brw.Write([]byte("XE"))
	brw.Close()
	os.Exit(0)
}
	