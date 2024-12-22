package bdump

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const verifyBDXURL = `https://uc.fastbuilder.pro/verifybdx.web`
const userAgent = "PhoenixBuilder/General"

// bool corrupted
// string username
// error error
func VerifyBDX(hashsum []byte, sign []byte) (bool, string, error) {
	if sign[0] == 0 && sign[1] == 0x8B {
		return VerifyBDXNew(hashsum, sign)
	} else {
		return false, "Unknown User(This file is using a deprecated signing method)", nil
	}
	hexOfHash := hex.EncodeToString(hashsum)
	body := fmt.Sprintf(`{"hash": "%s", "sign": "%s"}`, hexOfHash, base64.StdEncoding.EncodeToString(sign))
	request, err := http.NewRequest("POST", verifyBDXURL, strings.NewReader(body))
	if err != nil {
		return false, "", err
	}
	request.Header.Add("User-Agent", userAgent)
	c := &http.Client{}
	response, err := c.Do(request)
	if err != nil {
		return false, "", err
	}
	if response.StatusCode != 200 {
		return false, "", fmt.Errorf("Invalid status code: %d", response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	_ = response.Body.Close()
	c.CloseIdleConnections()
	if err != nil {
		return false, "", err
	}
	var rb map[string]interface{}
	err = json.Unmarshal(data, &rb)
	isCorrupted, _ := rb["corrupted"].(bool)
	if isCorrupted {
		return true, "", nil
	}
	isSucc, _ := rb["success"].(bool)
	if !isSucc {
		errmsg := rb["message"].(string)
		return false, "", fmt.Errorf("%s", errmsg)
	}
	un, _ := rb["username"].(string)
	return false, un, nil
}

const constantServerKey = "-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEAzOoZfky1sYQXkTXWuYqf7HZ+tDSLyyuYOvyqt/dO4xahyNqvXcL5\n1A+eNFhsk6S5u84RuwsUk7oeNDpg/I0hbiRuJwCxFPJKNxDdj5Q5P5O0NTLR0TAT\nNBP7AjX6+XtNB/J6cV3fPcduqBbN4NjkNZxP4I1lgbupIR2lMKU9lXEn58nFSqSZ\nvG4BZfYLKUiu89IHaZOG5wgyDwwQrejxqkLUftmXibUO4s4gf8qAiLp3ukeIPYRj\nwGhGNlUfdU0foCxf2QwAoBV2xREL8/Sx1AIvmoVUg1SqCiIVMvbBkDoFfkzPZCgC\nLtmbkmqZJnpoBVHcBhBdUYsfyM6QwtWBNQIDAQAB\n-----END RSA PUBLIC KEY-----"

// bool corrupted
// string username
// error error
func VerifyBDXNew(hashsum []byte, sign []byte) (bool, string, error) {
	if sign[0] != 0 || sign[1] != 0x8B {
		panic("Not a valid 2nd generation signature format")
		return false, "", fmt.Errorf("Not a valid 2nd generation signature format")
	}
	reader := bytes.NewReader(sign[2:])
	certLenIndicator := make([]byte, 2)
	reader.Read(certLenIndicator)
	certLen := int(binary.LittleEndian.Uint16(certLenIndicator))
	certPartBuf := make([]byte, certLen)
	reader.Read(certPartBuf)
	certPart := string(certPartBuf)
	firstSplit := strings.Split(certPart, "::")
	if len(firstSplit) != 2 {
		return true, "", nil
	}
	serverKeyDer, _ := pem.Decode([]byte(constantServerKey))
	csk, _ := x509.ParsePKCS1PublicKey(serverKeyDer.Bytes)
	signature, _ := hex.DecodeString(firstSplit[1])
	sum1 := sha256.Sum256([]byte(firstSplit[0]))
	err := rsa.VerifyPKCS1v15(csk, crypto.SHA256, sum1[:], signature)
	if err != nil {
		return true, "", nil
	}
	firstPart := firstSplit[0]
	fpContent := strings.Split(firstPart, "|")
	parsedPEM, _ := pem.Decode([]byte(fpContent[0]))
	publicKey, err := x509.ParsePKCS1PublicKey(parsedPEM.Bytes)
	if err != nil {
		return true, "", nil
	}
	signatureLen := reader.Len()
	signature = make([]byte, signatureLen)
	reader.Read(signature)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashsum, signature)
	if err != nil {
		return true, "", nil
	}
	return false, fpContent[1], nil
}

// bool signed
// bool corrupted
// string username
// error
func VerifyStreamBDX(stream io.Reader) (bool, bool, string, error) {
	last_block := make([]byte, 2048)
	cur_block := make([]byte, 2048)
	afterFirstRun := false
	hash := sha256.New()
	for {
		if afterFirstRun {
			copy(last_block, cur_block)
		}
		n, _ := io.ReadAtLeast(stream, cur_block, 2048)
		if n != 2048 {
			var buffered_block []byte
			if n == 0 {
				buffered_block = last_block
			} else {
				if !afterFirstRun {
					buffered_block = cur_block[:n]
				} else {
					buffered_block = append(last_block, cur_block[:n]...)
				}
			}
			bbl := len(buffered_block)
			if buffered_block[bbl-1] != 90 {
				// Not signed
				return false, false, "", nil
			}
			signlen := int(buffered_block[bbl-2])
			var sign []byte
			var bodyLeft []byte
			if signlen == int(255) {
				signlenBuf := buffered_block[bbl-4 : bbl-2]
				signlen = int(binary.BigEndian.Uint16(signlenBuf))
				if signlen >= bbl-4 {
					return false, false, "", fmt.Errorf("Too long signature")
				}
				sign = buffered_block[bbl-signlen-4 : bbl-4]
				bodyLeft = buffered_block[:bbl-signlen-5]
			} else {
				sign = buffered_block[bbl-signlen-2 : bbl-2]
				bodyLeft = buffered_block[:bbl-signlen-3]
			}
			hash.Write(bodyLeft)
			hash_val := hash.Sum(nil)
			cor, usn, err := VerifyBDX(hash_val, sign)
			return true, cor, usn, err
		}
		if afterFirstRun {
			hash.Write(last_block)
		} else {
			afterFirstRun = true
		}
	}
}
