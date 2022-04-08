package bridge

import (
	"os"
	"fmt"
	"time"
	"regexp"
	"io/ioutil"
	"encoding/hex"
	"path/filepath"
	"crypto/sha256"
	"encoding/json"
	"github.com/blakesmith/ar"
)

type manifest_struct struct {
	Identifier string `json:"identifier"`
	Name string `json:"name"`
	Description string `json:"description"`
	Author string `json:"author"`
	Version string `json:"version"`
	RelatedInformation string `json:"related_information"`
	EntryPoint string `json:"entrypoint"`
}

func MakePackage(path string, out string) int {
	manifestFile, err:=ioutil.ReadFile(path)
	if(err!=nil) {
		fmt.Printf("Error reading manifest JSON file: %v\n", err)
		return 26
	}
	var manifest manifest_struct
	err=json.Unmarshal(manifestFile, &manifest)
	if(err!=nil) {
		fmt.Printf("Invalid manifest JSON file: %v\n",err)
		return 27
	}
	if(len(manifest.Identifier)>32||len(manifest.Identifier)<4) {
		fmt.Printf("Package identifier is too long or too short\n")
		return 28
	}
	identifierREGEXP:=regexp.MustCompile("^([A-Za-z0-9_-]|\\.)*$")
	if(!identifierREGEXP.MatchString(manifest.Identifier)) {
		fmt.Printf("Illegal package identifier\n")
		return 28
	}
	if(len(out)==0) {
		out=fmt.Sprintf("%s/%s.scb",filepath.Dir(path),manifest.Identifier)
	}
	os.Remove(out)
	outfile, err:=os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if(err!=nil) {
		fmt.Printf("Failed to write to output file: %v\n",err)
		return 29
	}
	all_done:=false
	defer func() {
		if(!all_done) {
			outfile.Close()
			os.Remove(out)
		}
	} ()
	writer:=ar.NewWriter(outfile)
	err=writer.WriteGlobalHeader()
	if(err!=nil) {
		fmt.Printf("Failed to write to output file: %v\n",err)
		return 29
	}
	writer.WriteHeader(&ar.Header {
		Name: "manifest.json",
		ModTime: time.Now(),
		Uid: 1000,
		Gid: 1000,
		Mode: 0600,
		Size: int64(len(manifestFile)),
	})
	_,err=writer.Write([]byte(manifestFile))
	if(err!=nil) {
		fmt.Printf("Failed to write to output file: %v\n",err)
		return 29
	}
	scriptPath:=filepath.Dir(path)
	var wdir func(string) int
	wdir=func(dirpath string) int {
		contents, err:=os.ReadDir(dirpath)
		if(err!=nil) {
			fmt.Printf("Failed to read script directory: %v\n",err)
			return 30
		}
		for _, entry:=range contents {
			fp:=filepath.Join(dirpath, entry.Name())
			if(filepath.Clean(fp)==filepath.Clean(path)||filepath.Clean(fp)==filepath.Clean(out)) {
				continue
			}
			if(len(entry.Name())>16) {
				fmt.Printf("Filename too long for file %s.\n",fp)
				return 32
			}
			if(entry.IsDir()) {
				nr:=wdir(fp)
				if(nr!=0) {
					return nr
				}
				continue
			}
			content, err:=ioutil.ReadFile(fp)
			if(err!=nil) {
				fmt.Printf("Failed to read file %s: %v\n",fp, err)
				return 30
			}
			hash:=sha256.Sum256(content)
			hashhex:=hex.EncodeToString(hash[:])
			
			writer.WriteHeader(&ar.Header {
				Name: fmt.Sprintf(".sign:%s_%s",entry.Name(),hashhex)[0:16],
				ModTime: time.Now(),
				Uid: 1000,
				Gid: 1000,
				Mode: 0600,
				Size: int64(len(hashhex)),
			})
			writer.Write([]byte(hashhex))
			writer.WriteHeader(&ar.Header {
				Name: fp[len(scriptPath)+1:],
				ModTime: time.Now(),
				Uid: 1000,
				Gid: 1000,
				Mode: 0600,
				Size: int64(len(content)),
			})
			_, err=writer.Write(content)
			if(err!=nil) {
				fmt.Printf("Failed to write to archive: %s: %v\n",fp, err)
				return 31
			}
		}
		return 0
	}
	nc:=wdir(scriptPath)
	if(nc!=0) {
		return nc
	}
	all_done=true
	outfile.Close()
	return 0
}
	