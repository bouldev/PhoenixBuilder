package download_wrapper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"phoenixbuilder/fastbuilder/lib/utils/compress_wrapper"
	"phoenixbuilder/fastbuilder/lib/utils/lang"
	"phoenixbuilder/fastbuilder/lib/utils/string_wrapper"
	"strconv"
	"strings"
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

type readerCtx struct {
	ctx context.Context
	r   io.Reader
}

func (r *readerCtx) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func NewReaderCtx(ctx context.Context, r io.Reader) io.Reader {
	return &readerCtx{ctx: ctx, r: r}
}

func DownloadMicroContentWithCtx(ctx context.Context, sourceUrl string) ([]byte, error) {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(contents, NewReaderCtx(ctx, resp.Body)); err == nil {
		return contents.Bytes(), nil
	} else {
		return nil, err
	}
}

func DownloadSmallContent(sourceUrl string) ([]byte, error) {
	// Get the data
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return nil, err
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
		return contents.Bytes(), nil
	} else {
		return nil, err
	}
}

type Provider struct {
	// check version or hash
	VersionUrl string `json:"version_url"`
	HashUrl    string `json:"hash_url"`
	Url        string `json:"url"`
	Compress   string `json:"compress"`
}

func DownloadContent(url string, compress string) ([]byte, error) {
	if downloadFile, err := DownloadSmallContent(url); err == nil {
		if compress == "brotli" {
			if data, err := compress_wrapper.DecompressBrotli(downloadFile); err == nil {
				downloadFile = data
			} else {
				return nil, fmt.Errorf("compress file broken")
			}
		}
		return downloadFile, nil
	}
	return nil, fmt.Errorf("fail to download file")
}

func DownloadContentFromCandidatesIfNecessaryKey(providers []Provider, replacements map[string]string, existFileHash string, existFileVersion uint64) (content []byte, version string, err error) {
	for _, provider := range providers {
		url := provider.Url
		for k, v := range replacements {
			url = strings.ReplaceAll(url, fmt.Sprintf("${%v}", k), v)
		}
		// check version
		if provider.VersionUrl != "" {
			VersionUrl := provider.VersionUrl
			for k, v := range replacements {
				VersionUrl = strings.ReplaceAll(VersionUrl, fmt.Sprintf("${%v}", k), v)
			}
			if data, err := DownloadMicroContent(VersionUrl); err == nil {
				version := strings.TrimSpace(string(data))
				versionNumber, err := string_wrapper.TranslateVersionStringToNumber(version)
				if err != nil {
					continue
				}
				if versionNumber > existFileVersion {
					url = strings.ReplaceAll(url, "${version}", version)
					if content, err := DownloadContent(url, provider.Compress); err == nil {
						return content, version, nil
					} else {
						continue
					}
				} else {
					return nil, version, nil
				}
			}
		}
		// check hash
		if provider.HashUrl != "" {
			HashUrl := provider.HashUrl
			for k, v := range replacements {
				HashUrl = strings.ReplaceAll(HashUrl, fmt.Sprintf("${%v}", k), v)
			}
			if data, err := DownloadMicroContent(HashUrl); err == nil {
				hash := strings.TrimSpace(string(data))
				if existFileHash == hash {
					return nil, "", nil
				} else {
					url = strings.ReplaceAll(url, "${hash}", hash)
					if content, err := DownloadContent(url, provider.Compress); err == nil {
						return content, "", nil
					} else {
						continue
					}
				}
			}
		}
	}
	return nil, "", lang.Errorf("cannot connect to update server")
}
