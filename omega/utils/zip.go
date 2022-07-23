package utils

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func GetUnZipSize(zipFile string) (int64, error) {
	if s, err := os.Stat(zipFile); err != nil {
		return 0, err
	} else {
		return s.Size(), nil
	}
}

func UnZip(zipFile io.ReaderAt, size int64, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0644); err != nil {
		return err
	}
	fr, err := zip.NewReader(zipFile, size)
	if err != nil {
		return err
	}
	for _, file := range fr.File {
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(file.Name, 0644)
			if err != nil {
				return err
			}
		}
		r, err := file.Open()
		if err != nil {
			return err
		}
		fullPath := path.Join(dstDir, file.Name)
		dir := path.Dir(fullPath)
		os.MkdirAll(dir, 0755)
		NewFile, err := os.Create(fullPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(NewFile, r); err != nil {
			return err
		}
		if err := NewFile.Close(); err != nil {
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}
	}
	return nil
}

func Zip(srcDir string, zipFile *os.File, ignores []string) error {
	archive := zip.NewWriter(zipFile)
	defer archive.Close()
	return filepath.Walk(srcDir, func(filePath string, info os.FileInfo, _ error) error {
		if filePath == srcDir {
			return nil
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filePath[len(srcDir)+1:]
		header.Name = strings.ReplaceAll(header.Name, "\\", "/")
		for _, ignore := range ignores {
			if strings.HasPrefix(header.Name, ignore) {
				return nil
			}
		}
		if info.IsDir() {
			return nil
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}
		return nil
	})
}
