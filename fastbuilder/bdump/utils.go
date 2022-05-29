package bdump

import (
	"crypto"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/sha256"
	"net/http"
	"encoding/pem"
	"encoding/hex"
	"phoenixbuilder/fastbuilder/configuration"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"bytes"
)

const signBDXURL=`https://uc.fastbuilder.pro/signbdx.web`
const verifyBDXURL=`https://uc.fastbuilder.pro/verifybdx.web`
const userAgent = "PhoenixBuilder/General"

// SignBDX(fileContent)
// []byte - sign
// error  - err
func SignBDX(filecontent []byte, privateKeyString string, cert string) ([]byte, error) {
	if(len(privateKeyString)!=0&&len(cert)!=0) {
		return SignBDXNew(filecontent,privateKeyString,cert)
	}
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
	if(sign[0]==0&&sign[1]==0x8B) {
		return VerifyBDXNew(filecontent,sign)
	}
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

// SignBDXNew(fileContent)
// []byte - sign
// error  - err
func SignBDXNew(filecontent []byte, privateKeyString string, cert string) ([]byte, error) {
	buf:=bytes.NewBuffer([]byte{})
	derKey, _ := pem.Decode([]byte(privateKeyString))
	privateKey, err:=x509.ParsePKCS1PrivateKey(derKey.Bytes)
	if(err!=nil) {
		return nil, err
	}
	hashedFileContent:=sha256.Sum256(filecontent)
	signContent, err:=privateKey.Sign(rand.Reader,hashedFileContent[:],crypto.SHA256)
	if(err!=nil) {
		return nil, err
	}
	buf.Write([]byte{0x00,0x8B})
	{
		certLenIndicator:=make([]byte, 2)
		binary.LittleEndian.PutUint16(certLenIndicator,uint16(len(cert)))
		buf.Write(certLenIndicator)
	}
	buf.Write([]byte(cert))
	buf.Write(signContent)
	return buf.Bytes(), nil
}

const constantServerKey="-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEAzOoZfky1sYQXkTXWuYqf7HZ+tDSLyyuYOvyqt/dO4xahyNqvXcL5\n1A+eNFhsk6S5u84RuwsUk7oeNDpg/I0hbiRuJwCxFPJKNxDdj5Q5P5O0NTLR0TAT\nNBP7AjX6+XtNB/J6cV3fPcduqBbN4NjkNZxP4I1lgbupIR2lMKU9lXEn58nFSqSZ\nvG4BZfYLKUiu89IHaZOG5wgyDwwQrejxqkLUftmXibUO4s4gf8qAiLp3ukeIPYRj\nwGhGNlUfdU0foCxf2QwAoBV2xREL8/Sx1AIvmoVUg1SqCiIVMvbBkDoFfkzPZCgC\nLtmbkmqZJnpoBVHcBhBdUYsfyM6QwtWBNQIDAQAB\n-----END RSA PUBLIC KEY-----"

// bool corrupted
// string username
// error error
func VerifyBDXNew(filecontent []byte, sign []byte) (bool, string, error) {
	if(sign[0]!=0||sign[1]!=0x8B) {
		panic("Not a valid 2nd generation signature format");
		return false, "", fmt.Errorf("Not a valid 2nd generation signature format");
	}
	reader:=bytes.NewReader(sign[2:])
	certLenIndicator:=make([]byte, 2)
	reader.Read(certLenIndicator)
	certLen:=int(binary.LittleEndian.Uint16(certLenIndicator))
	certPartBuf:=make([]byte, certLen)
	reader.Read(certPartBuf)
	certPart:=string(certPartBuf)
	firstSplit:=strings.Split(certPart, "::")
	if(len(firstSplit)!=2) {
		fmt.Printf("%v\n", "111")
		return true, "", nil
	}
	serverKeyDer, _ := pem.Decode([]byte(constantServerKey))
	csk, _ := x509.ParsePKCS1PublicKey(serverKeyDer.Bytes)
	signature, _ := hex.DecodeString(firstSplit[1])
	sum1:=sha256.Sum256([]byte(firstSplit[0]))
	err:=rsa.VerifyPKCS1v15(csk, crypto.SHA256, sum1[:], signature)
	if(err!=nil) {
		return true, "", nil
	}
	firstPart:=firstSplit[0]
	fpContent:=strings.Split(firstPart,"|")
	parsedPEM, _ := pem.Decode([]byte(fpContent[0]))
	publicKey, err:=x509.ParsePKCS1PublicKey(parsedPEM.Bytes)
	if(err!=nil) {
		return true, "", nil
	}
	signatureLen:=reader.Len()
	signature=make([]byte, signatureLen)
	reader.Read(signature)
	sum1=sha256.Sum256(filecontent)
	err=rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, sum1[:], signature)
	if(err!=nil) {
		return true, "", nil
	}
	return false, fpContent[1], nil
}

