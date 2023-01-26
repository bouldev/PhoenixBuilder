package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func WriteJsonData(fname string, data interface{}) error {
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	enc.SetEscapeHTML(false)
	err = enc.Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func GetFileData(fname string) ([]byte, error) {
	fp, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, 0644)
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

func GetJsonData(fname string, ptr interface{}) error {
	data, err := GetFileData(fname)

	if err != nil {
		return err
	}
	if data == nil || len(data) == 0 {
		return nil
	}
	data = bytes.Trim(data, "\xef\xbb\xbf")
	err = json.Unmarshal(data, ptr)
	if err != nil {
		return err
	}
	return nil
}

func CopyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		// stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		// if !ok {
		// 	return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		// }

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		// if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
		// 	return err
		// }

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func GetFileNotFindStack(path string) (find bool, isDir bool, errStr []string) {
	if stat, err := os.Stat(path); err == nil {
		isDir := stat.Mode().IsDir()
		return true, isDir, nil
	}
	errStr = []string{}
	for {
		if path == "." {
			return false, false, errStr
		}
		if _, err := os.Stat(path); err != nil {
			errStr = append(errStr, fmt.Sprintf("无法找到 %v", path))
			path = filepath.Dir(path)
		} else {
			errStr = append(errStr, fmt.Sprintf("可以找到 %v ", path))
			return false, false, errStr
		}
	}
}

func WriteJsonDataWithAttachment(fname string, data interface{}, remapping map[string]map[string]string) error {
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		return err
	}
	fmt.Println(remapping)
	if currentJsonBytes, err := json.Marshal(data); err == nil {
		if remapping["orig"]["json"] == string(currentJsonBytes) {
			_, err := file.Write([]byte(remapping["orig"]["file"]))
			return err
		}
	}

	buf := bytes.NewBuffer([]byte{})

	enc := json.NewEncoder(buf)
	enc.SetIndent("", "\t")
	enc.SetEscapeHTML(false)
	err = enc.Encode(data)
	if err != nil {
		return err
	}
	PostWriteJsonFile(buf, remapping)
	file.Write(buf.Bytes())

	return nil
}

func PostWriteJsonFile(w io.Writer, mapping map[string]map[string]string) {
	if comments, hasK := mapping["comments"]; hasK {
		for ls, s := range comments {
			w.Write([]byte(fmt.Sprintf("// line: %v %v\n", ls, s)))
		}
	}
	w.Write([]byte(fmt.Sprintf("\n")))
	if fields, hasK := mapping["fields"]; hasK {
		for ls, s := range fields {
			w.Write([]byte(fmt.Sprintf("// OMEGA_START: %v\n%v\n//OMEGA_END\n\n", ls, s)))
		}
	}
}

func PreScanJsonFile(data []byte, mapping map[string]map[string]string) (reducedJson []byte, err error) {
	comments := map[string]string{}
	fields := map[string]string{}
	reducedJsonLines := make([]string, 0)
	lines := strings.Split(string(data), "\n")
	inUnLeveledField := false
	UnLeveledFiledName := ""
	for lineI, line := range lines {
		if inUnLeveledField {
			if strings.HasPrefix(line, "//OMEGA_END") || strings.HasPrefix(line, "// OMEGA_END") {
				UnLeveledFiledName = ""
				inUnLeveledField = false
				continue
			} else {
				if _, hasK := fields[UnLeveledFiledName]; hasK {
					fields[UnLeveledFiledName] += "\n" + line
				} else {
					fields[UnLeveledFiledName] = line
				}

				continue
			}
		}
		if strings.HasPrefix(line, "//OMEGA_START:") {
			inUnLeveledField = true
			UnLeveledFiledName = strings.TrimSpace(strings.TrimLeft(line, "//OMEGA_START:"))
			if _, hasK := fields[UnLeveledFiledName]; hasK {
				return nil, fmt.Errorf("域冲突: 具有两个 OMEGA_START: %v 标记，请删除其中一个")
			}
			continue
		}
		if strings.HasPrefix(line, "// OMEGA_START:") {
			inUnLeveledField = true
			UnLeveledFiledName = strings.TrimSpace(strings.TrimLeft(line, "// OMEGA_START:"))
			continue
		}
		flag := false
		for _, c := range []byte(line) {
			if c == '\t' || c == ' ' {
				// h += "c"
			} else if c == '/' {
				flag = true
				break
			} else {
				break
			}
		}
		if flag {
			comments[fmt.Sprintf("%v", lineI)] = line
		} else {
			reducedJsonLines = append(reducedJsonLines, line)
		}
	}
	mapping["comments"] = comments
	mapping["fields"] = fields
	return []byte(strings.Join(reducedJsonLines, "\n")), nil
}

func GetJsonDataWithAttachment(fname string, ptr interface{}) (remapping map[string]map[string]string, err error) {
	data, err := GetFileData(fname)
	if err != nil {
		return nil, err
	}
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	data = bytes.Trim(data, "\xef\xbb\xbf")
	remapping = map[string]map[string]string{"orig": {"file": string(data)}}
	data, err = PreScanJsonFile(data, remapping)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, ptr)
	if err != nil {
		return nil, err
	}
	origBytes, _ := json.Marshal(ptr)
	remapping["orig"]["json"] = string(origBytes)
	return remapping, nil
}

func MoveDir(oldPath, newPath string) (err error) {
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func IsDirEmpty(dir string) bool {
	contents, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(contents) == 0
}
