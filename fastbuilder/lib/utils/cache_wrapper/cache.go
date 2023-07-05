package cache_wrapper

import (
	"os"
	"path"
	"phoenixbuilder/fastbuilder/lib/utils/crypto_wrapper"
	"phoenixbuilder/fastbuilder/lib/utils/file_wrapper"
	"phoenixbuilder/fastbuilder/lib/utils/lang"
	"phoenixbuilder/fastbuilder/lib/utils/triple_wrapper"
	"strings"
)

var cacheDir = ""
var candidateDirs []string

func AddCandicatesDir(candidates []string, candidateIfHomeExist string) {
	candidateDirs = append(candidateDirs, candidates...)
	homedir, err := os.UserHomeDir()
	if err == nil && candidateIfHomeExist != "" {
		candidateDirs = append(candidateDirs, path.Join(homedir, candidateIfHomeExist))
	}
}

func GetCacheDir() (string, error) {
	if cacheDir != "" {
		return cacheDir, nil
	}
	if selectedDir, err := file_wrapper.CreateDirFromCandidatesPaths(candidateDirs, 0755); err != nil {
		return "", lang.Errorf("fail to create cache dir")
	} else {
		cacheDir = selectedDir
		return selectedDir, nil
	}
}

func UpdateFileCache(key string, data []byte, exec bool) (err error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}
	caches := map[string]string{}
	if err := file_wrapper.GetJsonData(path.Join(cacheDir, "entries.json"), &caches); err != nil {
		// ignore error
	}
	if oldFile, found := caches[key]; found {
		os.Remove(oldFile)
	}
	suffix := ""
	if exec && triple_wrapper.IsWindows() {
		suffix = ".exe"
	}
	caches[key] = crypto_wrapper.BytesMD5Str(data) + suffix
	if err := file_wrapper.WriteFile(path.Join(cacheDir, crypto_wrapper.BytesMD5Str(data)+suffix), data, 0755); err != nil {
		return err
	}
	return file_wrapper.WriteJsonData(path.Join(cacheDir, "entries.json"), caches)
}

func GetFilePathFromCache(key string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	caches := map[string]string{}
	if err = file_wrapper.GetJsonData(path.Join(cacheDir, "entries.json"), &caches); err == nil {
		if md5str, found := caches[key]; found && file_wrapper.Exists(path.Join(cacheDir, md5str)) {
			fileName := path.Join(cacheDir, md5str)
			if realMD5, err := file_wrapper.GetFileMD5Str(fileName); err == nil && strings.TrimRight(realMD5, ".exe") == strings.TrimRight(md5str, ".exe") {
				return fileName, nil
			}
		}
	}
	return "", lang.Errorf("file not found")
}

func GetFileHashFromCache(key string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	caches := map[string]string{}
	if err := file_wrapper.GetJsonData(path.Join(cacheDir, "entries.json"), &caches); err == nil {
		if md5str, found := caches[key]; found && file_wrapper.Exists(path.Join(cacheDir, md5str)) {
			fileName := path.Join(cacheDir, md5str)
			if realMD5, err := file_wrapper.GetFileMD5Str(fileName); err == nil && realMD5 == strings.TrimRight(md5str, ".exe") {
				return realMD5, nil
			}
		}
	}
	return "", lang.Errorf("file not found")
}

func UpdateCacheValue(key string, value string, writeHash bool) (err error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}
	caches := map[string]string{}
	if err := file_wrapper.GetJsonData(path.Join(cacheDir, "entries.json"), &caches); err != nil {
		// ignore error
	}
	caches[key] = value
	if err = file_wrapper.WriteJsonData(path.Join(cacheDir, "entries.json"), caches); err != nil {
		return err
	} else {
		if writeHash {
			return UpdateCacheValue(crypto_wrapper.StrSHA256Str(key+".hash.guarantee"), crypto_wrapper.BytesSHA256Str([]byte(value)), false)
		} else {
			return nil
		}
	}
}

func GetValueFromCache(key string, checkHash bool) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	caches := map[string]string{}
	if err := file_wrapper.GetJsonData(path.Join(cacheDir, "entries.json"), &caches); err == nil {
		if value, found := caches[key]; found {
			if checkHash {
				hash, err := GetValueFromCache(crypto_wrapper.StrSHA256Str(key+".hash.guarantee"), false)
				if err != nil || hash != crypto_wrapper.BytesSHA256Str([]byte(value)) {
					return "", lang.Errorf("hash mismatch")
				} else {
					return value, nil
				}
			} else {
				return value, nil
			}
		}
	}
	return "", lang.Errorf("value not found")
}
