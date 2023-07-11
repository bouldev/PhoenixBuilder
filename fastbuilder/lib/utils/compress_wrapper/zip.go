package compress_wrapper

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Zip(src_dir string, zipfile io.Writer, ignores []string) error {
	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	src_dir = path.Clean(src_dir)
	return filepath.Walk(src_dir, func(filePath string, info os.FileInfo, _ error) error {
		if filePath == src_dir {
			return nil
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filePath[len(src_dir)+1:]
		header.Name = strings.ReplaceAll(header.Name, "\\", "/")
		for _, ignore := range ignores {
			if strings.HasPrefix(header.Name, ignore) {
				return nil
			}
		}
		if info.IsDir() {
			return nil
		} else {
			header.Method = zip.Deflate
		}

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

func GetUnZipSize(zipFile string) (int64, error) {
	if s, err := os.Stat(zipFile); err != nil {
		return 0, err
	} else {
		return s.Size(), nil
	}
}

func UnZip(zipfile io.ReaderAt, size int64, dst_dir string) error {
	os.Mkdir(dst_dir, 0755)
	fr, err := zip.NewReader(zipfile, size)
	if err != nil {
		return err
	}
	for _, file := range fr.File {
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(file.Name, 0644)
			if err != nil {
				return err
			}
			continue
		}
		r, err := file.Open()
		if err != nil {
			return err
		}
		fullPath := path.Join(dst_dir, file.Name)
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
