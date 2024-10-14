/*
PhoenixBuilder specific packages.
Author: Liliya233, Happy2018new
*/
package NetEaseSkin

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	fbauth "phoenixbuilder/fastbuilder/mv4"
	"strings"

	_ "embed"
)

//go:embed default_wide_skin_resource_patch.json
var DefaultWideSkinResourcePatch []byte

//go:embed default_slim_skin_resource_patch.json
var DefaultSlimSkinResourcePatch []byte

//go:embed skin_geometry.json
var DefaultSkinGeometry []byte

// ...
func IsZIPFile(fileData []byte) bool {
	return len(fileData) >= 4 && bytes.Equal(fileData[0:4], []byte("PK\x03\x04"))
}

// 从 url 指定的网址下载文件，
// 并返回该文件的二进制形式
func DownloadFile(url string) (result []byte, err error) {
	// 获取 HTTP 响应
	httpResponse, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile: %v", err)
	}
	defer httpResponse.Body.Close()
	// 读取文件数据
	result, err = io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile: %v", err)
	}
	// 返回值
	return
}

// 从 authResponse 指定的网址下载文件，
// 并处理为有效的皮肤数据，
// 然后保存在 skin 中
func GetSkinFromAuthResponse(authResponse fbauth.AuthResponse, skin *Skin) error {
	// 初始化
	var skinImageData []byte
	// 从远程服务器下载皮肤文件
	res, err := DownloadFile(authResponse.BotSkin.SkinDownloadURL)
	if err != nil {
		return fmt.Errorf("GetSkinFromAuthResponse: %v", err)
	}
	// 获取并设置皮肤数据
	{
		// 如果这是一个普通的皮肤，
		// 那么 res 就是该皮肤的 PNG 二进制形式，
		// 并且该皮肤使用的骨架格式为默认格式
		skin.SkinItemID = authResponse.BotSkin.ItemID
		skin.FullSkinData, skin.SkinGeometry = res, DefaultSkinGeometry
		skin.SkinIsSlim = authResponse.BotSkin.SkinIsSlim
		skinImageData = res
		// 设置皮肤默认资源路径
		if skin.SkinIsSlim {
			skin.SkinResourcePatch = DefaultSlimSkinResourcePatch
		} else {
			skin.SkinResourcePatch = DefaultWideSkinResourcePatch
		}
		// 如果这是一个高级的皮肤(比如 4D 皮肤)，
		// 那么 res 是一个压缩包，
		// 我们需要处理这个压缩包以得到皮肤文件
		if IsZIPFile(res) {
			skinImageData, err = ConvertZIPToSkin(skin)
			if err != nil {
				return fmt.Errorf("GetSkinFromAuthResponse: %v", err)
			}
		}
	}
	// 将皮肤 PNG 二进制形式解码为图片
	img, err := ConvertToPNG(skinImageData)
	if err != nil {
		return fmt.Errorf("GetSkinFromAuthResponse: %v", err)
	}
	// 设置皮肤像素、高度、宽度等数据
	skin.SkinPixels = img.(*image.NRGBA).Pix
	skin.SkinWidth, skin.SkinHight = img.Bounds().Dx(), img.Bounds().Dy()
	// 返回值
	return nil
}

// 从 skin.FullSkinData 指代的 ZIP
// 二进制负载提取皮肤数据，
// 并把处理好的皮肤数据保存到 skin 自身，
// 同时返回皮肤图片(PNG)的二进制表示
func ConvertZIPToSkin(skin *Skin) (skinImageData []byte, err error) {
	// 创建 ZIP 读取器
	reader, err := zip.NewReader(
		bytes.NewReader(skin.FullSkinData),
		int64(len(skin.FullSkinData)),
	)
	if err != nil {
		return nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
	}
	// 查找皮肤内容
	for _, file := range reader.File {
		// 皮肤数据
		if strings.HasSuffix(file.Name, ".png") && !strings.HasSuffix(file.Name, "_bloom.png") {
			r, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			defer r.Close()
			skinImageData, err = io.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
		}
		// 皮肤骨架信息
		if strings.HasSuffix(file.Name, "geometry.json") {
			r, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			defer r.Close()
			geometryData, err := io.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			ProcessGeometry(skin, geometryData)
		}
		// 处理 `manifest.json` or `pack_manifest.json`
		if strings.HasSuffix(file.Name, "manifest.json") {
			var manifest SkinManifest
			r, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			defer r.Close()
			manifestData, err := io.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			if err = json.Unmarshal(manifestData, &manifest); err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			skin.SkinUUID = manifest.Header.UUID
		}
	}
	// 返回值
	return
}

// 将 imageData 解析为 PNG 图片
func ConvertToPNG(imageData []byte) (image.Image, error) {
	buffer := bytes.NewBuffer(imageData)
	img, err := png.Decode(buffer)
	if err != nil {
		return nil, fmt.Errorf("ConvertToPNG: %v", err)
	}
	return img, nil
}
