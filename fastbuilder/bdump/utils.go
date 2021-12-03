package bdump

import (
	"crypto/sha256"
	"net/http"
	"encoding/hex"
	"phoenixbuilder/fastbuilder/configuration"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
	"strings"
)

const signBDXURL=`https://uc.fastbuilder.pro/signbdx.web`
const verifyBDXURL=`https://uc.fastbuilder.pro/verifybdx.web`
const userAgent = "PhoenixBuilder/General"

// SignBDX(fileContent)
// []byte - sign
// error  - err
func SignBDX(filecontent []byte) ([]byte, error) {
	hash:=sha256.New()
	hash.Write(filecontent)
	hexOfHash:=hex.EncodeToString(hash.Sum(nil))
	body:=fmt.Sprintf(`{"hash": "%s", "token": "%s"}`,hexOfHash,configuration.UserToken)
	request, err := http.NewRequest("POST", signBDXURL, strings.NewReader(body))
	if(err != nil){
		return nil, err
	}
	request.Header.Add("User-Agent", userAgent)
	c:=&http.Client{}
	response, err:=c.Do(request)
	if(err!=nil) {
		return nil, err
	}
	if(response.StatusCode != 200){
		return nil, fmt.Errorf("Invalid status code: %d",response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	_ = response.Body.Close()
	c.CloseIdleConnections()
	if(err!=nil) {
		return nil, err
	}
	var rb map[string]interface{}
	err=json.Unmarshal(data, &rb)
	isSucc, _:=rb["success"].(bool)
	if(!isSucc) {
		errmsg:=rb["message"].(string)
		return nil, fmt.Errorf("%s", errmsg)
	}
	sign, _:=rb["sign"].(string)
	theBytes, err := base64.StdEncoding.DecodeString(sign)
	if(err!=nil) {
		return nil, fmt.Errorf("Failed to decode hex: %v", err)
	}
	return theBytes, nil
}


// bool corrupted
// string username
// error error
func VerifyBDX(filecontent []byte, sign []byte) (bool, string, error) {
	hash:=sha256.New()
	hash.Write(filecontent)
	hexOfHash:=hex.EncodeToString(hash.Sum(nil))
	body:=fmt.Sprintf(`{"hash": "%s", "sign": "%s"}`,hexOfHash,base64.StdEncoding.EncodeToString(sign))
	request, err := http.NewRequest("POST", verifyBDXURL, strings.NewReader(body))
	if(err != nil){
		return false,"", err
	}
	request.Header.Add("User-Agent", userAgent)
	c:=&http.Client{}
	response, err:=c.Do(request)
	if(err!=nil) {
		return false,"", err
	}
	if(response.StatusCode != 200){
		return false,"", fmt.Errorf("Invalid status code: %d",response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	_ = response.Body.Close()
	c.CloseIdleConnections()
	if(err!=nil) {
		return false,"", err
	}
	var rb map[string]interface{}
	err=json.Unmarshal(data, &rb)
	isCorrupted, _ := rb["corrupted"].(bool)
	if(isCorrupted) {
		return true,"",nil
	}
	isSucc, _:=rb["success"].(bool)
	if(!isSucc) {
		errmsg:=rb["message"].(string)
		return false,"", fmt.Errorf("%s", errmsg)
	}
	un, _:=rb["username"].(string)
	return false,un,nil
}