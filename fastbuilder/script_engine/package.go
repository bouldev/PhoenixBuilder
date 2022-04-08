package script_engine

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"regexp"
	"encoding/json"
	"path/filepath"
	"github.com/blakesmith/ar"
	"rogchap.com/v8go"
	"io/ioutil"
	"crypto/sha256"
	"encoding/hex"
)

type ScriptPackage struct {
	Scripts map[string]*v8go.UnboundScript
	Datas map[string]string
	Identifier string
	Name string
	Description string
	Author string
	Version string
	RelatedInformation string
	Manifest string
	EntryPoint string
}

type manifest_struct struct {
	Identifier string `json:"identifier"`
	Name string `json:"name"`
	Description string `json:"description"`
	Author string `json:"author"`
	Version string `json:"version"`
	RelatedInformation string `json:"related_information"`
	EntryPoint string `json:"entrypoint"`
}

func LoadDebugPackage(isolate *v8go.Isolate, path string) (*ScriptPackage, error) {
	manifestFile, err:=ioutil.ReadFile(path)
	if(err!=nil) {
		return nil, err
	}
	var manifest manifest_struct
	err=json.Unmarshal(manifestFile, &manifest)
	if(err!=nil) {
		return nil, fmt.Errorf("Error reading manifest.json: %v",err)
	}
	if(len(manifest.Identifier)>32||len(manifest.Identifier)<4) {
		return nil, fmt.Errorf("Package identifier is too long or too short")
	}
	identifierREGEXP:=regexp.MustCompile("^([A-Za-z0-9_-]|\\.)*$")
	if(!identifierREGEXP.MatchString(manifest.Identifier)) {
		return nil, fmt.Errorf("Illegal package identifier")
	}
	pkg:=&ScriptPackage {
		Identifier: manifest.Identifier,
		Name: manifest.Name,
		Description: manifest.Description,
		Author: manifest.Author,
		Version: manifest.Version,
		RelatedInformation: manifest.RelatedInformation,
		Manifest: string(manifestFile),
		EntryPoint: manifest.EntryPoint,
		Scripts: map[string]*v8go.UnboundScript {},
		Datas: map[string]string {},
	}
	scriptPath:=filepath.Dir(path)
	var wdir func(string)error
	wdir=func(dirpath string) error {
		contents, err:=os.ReadDir(dirpath)
		if(err!=nil) {
			return err
		}
		for _, entry:=range contents {
			fp:=filepath.Join(dirpath, entry.Name())
			if(entry.IsDir()) {
				err:=wdir(fp)
				if(err!=nil) {
					return err
				}
				continue
			}
			contentb, err:=ioutil.ReadFile(fp)
			if(err!=nil) {
				return fmt.Errorf("Failed to read file %s: %v",fp, err)
			}
			content:=string(contentb)
			if(filepath.Ext(fp)==".js") {
				script, err:=isolate.CompileUnboundScript(content, fp, v8go.CompileOptions{})
				if(err!=nil) {
					return fmt.Errorf("Failed to compile script %s: %v",fp,err)
				}
				pkg.Scripts[fp[len(scriptPath)+1:]]=script
			}else{
				pkg.Datas[fp[len(scriptPath)+1:]]=content
			}
		}
		return nil
	}
	err=wdir(scriptPath)
	if(err!=nil) {
		return nil, err
	}
	return pkg, nil
}

func LoadPackage(isolate *v8go.Isolate,path string) (*ScriptPackage, error) {
	file, err := os.Open(path)
	if(err!=nil) {
		return nil, err
	}
	defer file.Close()
	reader:=ar.NewReader(file)
	manifestHeader, err := reader.Next()
	if(err!=nil) {
		return nil, err
	}
	if(manifestHeader.Name!="manifest.json") {
		return nil, fmt.Errorf("The first file should be manifest.json")
	}
	manifest_content:=make([]byte,manifestHeader.Size)
	_, err=reader.Read(manifest_content)
	if(err!=nil) {
		return nil, err
	}
	var manifest manifest_struct
	err=json.Unmarshal(manifest_content,&manifest)
	if(err!=nil) {
		return nil, fmt.Errorf("Error reading manifest.json: %v",err)
	}
	if(len(manifest.Identifier)>32||len(manifest.Identifier)<4) {
		return nil, fmt.Errorf("Package identifier is too long or too short")
	}
	identifierREGEXP:=regexp.MustCompile("^([A-Za-z0-9_-]|\\.)*$")
	if(!identifierREGEXP.MatchString(manifest.Identifier)) {
		return nil, fmt.Errorf("Illegal package identifier")
	}
	pkg:=&ScriptPackage {
		Identifier: manifest.Identifier,
		Name: manifest.Name,
		Description: manifest.Description,
		Author: manifest.Author,
		Version: manifest.Version,
		RelatedInformation: manifest.RelatedInformation,
		Manifest: string(manifest_content),
		EntryPoint: manifest.EntryPoint,
		Scripts: map[string]*v8go.UnboundScript {},
		Datas: map[string]string {},
	}
	nextSign:=""
	for {
		header, err := reader.Next()
		if(err==io.EOF) {
			break
		}
		if(err!=nil) {
			return nil, err
		}
		var buf bytes.Buffer
		io.Copy(&buf, reader)
		content:=string(buf.Bytes())
		if(header.Name[0:5]==".sign:") {
			nextSign=content
			continue
		}
		if(len(nextSign)!=0) {
			hash:=sha256.Sum256([]byte(content))
			if(hex.EncodeToString(hash[:])!=nextSign) {
				return nil, fmt.Errorf("Signature invalid for file %s, file may be corrupted.",header.Name)
			}
			nextSign=""
		}
		if(filepath.Ext(header.Name)==".js") {
			script, err:=isolate.CompileUnboundScript(content, fmt.Sprintf("%s/%s",path,header.Name), v8go.CompileOptions{})
			if(err!=nil) {
				return nil, fmt.Errorf("Failed to compile script %s: %v",header.Name,err)
			}
			pkg.Scripts[header.Name]=script
		}else{
			pkg.Datas[header.Name]=content
		}
	}
	return pkg, nil
}