package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tencentyun/cos-go-sdk-v5"
)


func AcquireCosClient() *cos.Client {
	u, _ := url.Parse("https://data-?????.cos.ap-shanghai.myqcloud.com")
	su, _ := url.Parse("https://cos.ap-??????.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  "?????????",
			SecretKey: "?????????",
		},
	})
	return client
}

var dir = flag.String("d", ".", "需要上传的目录")

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
	dirName := *dir
	fmt.Printf("dir to upload: %v\n", dirName)

	if err = filepath.Walk(dirName, func(fullPath string, info os.FileInfo, err error) error {
		i := 0
		for i = 0; i < len(dirName) && i < len(fullPath); i++ {
			if dirName[i] != fullPath[i] {
				break
			}
		}
		fileSubPath := fullPath[i:]
		if len(fileSubPath) > 0 && fileSubPath[0] == '/' {
			fileSubPath = fileSubPath[1:]
		}
		if info.IsDir() {
			return nil
		}
		fmt.Printf("uploading %v -> %v\n", fullPath, fileSubPath)
		_, err = client.Object.PutFromFile(context.Background(), fileSubPath, fullPath, nil)
		if err != nil {
			fmt.Printf("on uploading %v -> %v, error happen %v", fullPath, fileSubPath, err)
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}
	fmt.Println("Done")
}
