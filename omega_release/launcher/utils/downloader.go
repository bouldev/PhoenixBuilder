package utils

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cheggaaa/pb"
)

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

func DownloadMicroContent(sourceUrl string) []byte {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(contents, resp.Body); err == nil {
		return contents.Bytes()
	} else {
		panic(err)
	}
}

func DownloadSmallContent(sourceUrl string) []byte {
	// Get the data
	resp, err := http.Get(sourceUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Size
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	downloadSize := uint64(size)

	// Progress Bar
	bar := pb.New64(int64(downloadSize))
	bar.SetRefreshRate(time.Second)
	bar.SetUnits(pb.U_BYTES)
	bar.Start()
	defer bar.Finish()

	// Create our bytes counter and pass it to be used alongside our writer
	counter := &WriteCounter{
		DownloadSize: downloadSize,
		ProgressBar:  bar,
	}
	contents := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(contents, io.TeeReader(resp.Body, counter)); err == nil {
		return contents.Bytes()
	} else {
		panic(err)
	}
}
