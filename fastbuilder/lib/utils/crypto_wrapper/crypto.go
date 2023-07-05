package crypto_wrapper

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
)

func AES_ECB_Encrypt(plainText []byte, key []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(plainText)%aes.BlockSize != 0 {
		return nil, ErrDataSizeAndKeySizeMismatch
	}
	cipherText := make([]byte, 0)
	text := make([]byte, 16)
	for len(plainText) > 0 {
		cipher.Encrypt(text, plainText)
		plainText = plainText[aes.BlockSize:]
		cipherText = append(cipherText, text...)
	}
	return cipherText, nil
}

func AES_ECB_Decrypt(cipherText []byte, key string) ([]byte, error) {
	cipher, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, ErrDataSizeAndKeySizeMismatch
	}
	plainText := make([]byte, 0)
	text := make([]byte, 16)
	for len(cipherText) > 0 {
		cipher.Decrypt(text, cipherText)
		cipherText = cipherText[aes.BlockSize:]
		plainText = append(plainText, text...)
	}
	return plainText, nil
}

func PKCS7Pad(data []byte) []byte {
	paddingCount := aes.BlockSize - len(data)%aes.BlockSize
	if paddingCount == 0 {
		return data
	} else {
		return append(data, bytes.Repeat([]byte{byte(paddingCount)}, paddingCount)...)
	}
}

func ZeroPad(data []byte) []byte {
	paddingCount := aes.BlockSize - len(data)%aes.BlockSize
	if paddingCount == 0 {
		return data
	} else {
		return append(data, bytes.Repeat([]byte{byte(0)}, paddingCount)...)
	}
}

func PKCS7UPad(data []byte) []byte {
	padLength := int(data[len(data)-1])
	return data[:len(data)-padLength]
}

func StrMD5Str(data string) string {
	return BytesMD5Str([]byte(data))
}

func StrSHA256Str(data string) string {
	return BytesSHA256Str([]byte(data))
}

func BytesMD5Str(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

func BytesSHA256Str(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
