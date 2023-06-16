package file_wrapper

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func WriteFile(fname string, data []byte, perm os.FileMode) error {
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

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

func CopyDirectory(scrDir, dstDir string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dstDir, entry.Name())

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
			if err := CreateDirIfNotExists(destPath, 0755); err != nil {
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
			containedDir := path.Dir(destPath)
			CreateDirIfNotExists(containedDir, 0755)
			if err := CopyFile(sourcePath, destPath); err != nil {
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

func CopyFile(srcFile, dstFile string) error {
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

func DirExist(dirPath string) bool {
	if stat, err := os.Stat(dirPath); err != nil || !stat.IsDir() {
		return false
	}
	return true
}

func CreateDirIfNotExists(dir string, perm os.FileMode) error {
	if DirExist(dir) {
		return nil
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}
	return nil
}

func CreateDirFromCandidatesPaths(paths []string, perm os.FileMode) (string, error) {
	for _, path := range paths {
		if err := CreateDirIfNotExists(path, perm); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no suitable path found for creating directory")
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
			errStr = append(errStr, fmt.Sprintf("cannot find %v", path))
			path = filepath.Dir(path)
		} else {
			errStr = append(errStr, fmt.Sprintf("can find %v ", path))
			return false, false, errStr
		}
	}
}

func GetFileMD5Str(filePath string) (string, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
