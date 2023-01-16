package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	AccessUrl  = "https://data-?????.cos.ap-shanghai.myqcloud.com"
	ServiceUrl = "https://cos.ap-??????.myqcloud.com"
	SecretID   = "AKxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	SecretKey  = "????????????????????????????????"
)

func AcquireCosClient() *cos.Client {
	fmt.Printf("AccessUrl: %v... ServiceUrl: %v... SecretID: %v... SecretKey: %v... \n", AccessUrl[:16], ServiceUrl[:16], SecretID[:4], SecretKey[:1])
	u, _ := url.Parse(AccessUrl)
	su, _ := url.Parse(ServiceUrl)
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  SecretID,
			SecretKey: SecretKey,
		},
	})
	return client
}

var localDir = flag.String("l", ".", "local dir to upload")
var remoteDir = flag.String("r", "", "remote dir to upload to")

func main() {
	flag.Parse()
	client := AcquireCosClient()
	fmt.Println(client)

	s, _, err := client.Service.Get(context.Background())
	if err != nil {
		panic(err)
	}
	if len(s.Buckets) == 0 {
		_, err := client.Bucket.Put(context.Background(), nil)
		if err != nil {
			panic(err)
		}
	}
	for _, b := range s.Buckets {
		fmt.Printf("Bucket: %#v\n", b)
	}
	localDirName := *localDir
	remoteDirName := *remoteDir

	localDirName, err = filepath.Abs(localDirName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dir to upload: %v -> %v \n", localDirName, remoteDirName)

	if err = filepath.Walk(localDirName, func(fullPath string, info os.FileInfo, err error) error {
		i := 0
		// have to do so to avoid error remote path
		for i = 0; i < len(localDirName) && i < len(fullPath); i++ {
			if localDirName[i] != fullPath[i] {
				break
			}
		}
		fileSubPath := fullPath[i:]
		if len(fileSubPath) > 0 && fileSubPath[0] == '/' {
			fileSubPath = fileSubPath[1:]
		}
		// no need (also cannot) to upload dir
		if info.IsDir() {
			return nil
		}
		targetFilePath := fileSubPath
		if len(remoteDirName) > 0 && remoteDirName[len(remoteDirName)-1] != '/' {
			remoteDirName = strings.TrimRight(remoteDirName, "/")
			targetFilePath = remoteDirName + "/" + fileSubPath
		}
		fmt.Printf("uploading %v -> %v\n", fullPath, targetFilePath)
		_, err = client.Object.PutFromFile(context.Background(), targetFilePath, fullPath, nil)
		if err != nil {
			fmt.Printf("on uploading %v -> %v, error happen %v", fullPath, targetFilePath, err)
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}
	fmt.Println("Done")
}
