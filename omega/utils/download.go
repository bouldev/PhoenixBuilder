package utils

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
)

func GetBinaryHash(fileData []byte) string {
	cvt := func(in [16]byte) []byte {
		return in[:16]
	}
	hashedBytes := cvt(md5.Sum(fileData))
	return fmt.Sprintf("%x", hashedBytes)
}
func GetFileHash(filePath string) (string, error) {
	if IsFile(filePath) {
		fileData, err := GetFileData(filePath)
		if err != nil {
			return "", err
		}
		return GetBinaryHash(fileData), nil
	}
	return "", nil
}

type WriteCounter struct {
	Total        uint64
	DownloadSize uint64
	ProgressBar  *pb.ProgressBar
}

func (wc WriteCounter) PrintProgress() {
	wc.ProgressBar.Add64(int64(wc.Total))
	wc.ProgressBar.Increment()
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total = uint64(n)
	wc.PrintProgress()
	return n, nil
}

func DownloadMicroContent(sourceUrl string) ([]byte, error) {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(contents, resp.Body); err == nil {
		return contents.Bytes(), nil
	} else {
		return nil, err
	}
}

func DownloadSmallContent(sourceUrl string) ([]byte, error) {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	downloadSize := uint64(size)
	bar := pb.New64(int64(downloadSize))
	bar.SetRefreshRate(time.Second)
	bar.SetUnits(pb.U_BYTES)
	bar.Start()
	defer bar.Finish()

	counter := &WriteCounter{
		DownloadSize: downloadSize,
		ProgressBar:  bar,
	}
	contents := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(contents, io.TeeReader(resp.Body, counter)); err == nil {
		return contents.Bytes(), nil
	} else {
		return nil, err
	}
}

type SimpleDeployer struct {
	CacheFilePath      string
	TargetDeployDir    string
	SourceFileURL      string
	SourceFileMD5      string
	SourceFileMD5ByURL string
}

func (o *SimpleDeployer) needDownload() (needDownload bool, err error) {
	cacheHash, _ := GetFileHash(o.CacheFilePath)
	// if err != nil {
	// 	return true, nil
	// }
	if o.SourceFileMD5 == "" {
		remoteMD5HashStr, err := DownloadMicroContent(o.SourceFileMD5ByURL)
		if err != nil {
			return true, err
		}
		o.SourceFileMD5 = string(remoteMD5HashStr)
	}
	return !(cacheHash == o.SourceFileMD5), nil
}

func (o *SimpleDeployer) Deploy() (err error) {
	var sourceFileData []byte
	if needDownload, err := o.needDownload(); err != nil {

		return err
	} else if needDownload {
		downloadFileData, err := DownloadSmallContent(o.SourceFileURL)
		if err != nil {
			return err
		}
		sourceFileData = downloadFileData
		os.MkdirAll(path.Dir(o.CacheFilePath), 0755)
		file, err := os.OpenFile(o.CacheFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		file.Write(downloadFileData)
		defer file.Close()
	} else {
		cacheFileData, err := GetFileData(o.CacheFilePath)
		if err != nil {
			return err
		}
		sourceFileData = cacheFileData
	}
	reader := bytes.NewReader(sourceFileData)
	if strings.HasSuffix(o.SourceFileURL, ".zip") {
		return UnZip(reader, reader.Size(), o.TargetDeployDir)
	} else {
		return GZIPDecompress(bytes.NewReader(sourceFileData), o.TargetDeployDir)
	}
}
