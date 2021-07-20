package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	fbauth "cv4-auth-client/auth"
	"encoding/base64"
	"strings"
)

// RequestMinecraftChain requests a fully processed Minecraft JWT chain using the XSTS token passed, and the
// ECDSA private key of the client. This key will later be used to initialise encryption, and must be saved
// for when packets need to be decrypted/encrypted.
func RequestMinecraftChain(serverCode, password, token ,version string, key *ecdsa.PrivateKey) (string,string, error) {
	client := fbauth.CreateClient()
	data, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubKeyData := base64.StdEncoding.EncodeToString(data)
	chainAddr, err := client.Auth(serverCode, password, pubKeyData, token, version)
	if err != nil {
		return "", "", err
	}
	// The body of the requests holds a JSON object with one key in it, the 'identityPublicKey', which holds
	// the public key data of the private key passed.
	//body := fmt.Sprintf(`{"key":"%v","serverCode":"%v"}`, pubKeyData,serverCode)
	//request, _ := http.NewRequest("POST", minecraftAuthURL, strings.NewReader(body))
	//request.Header.Set("Content-Type", "application/json")
	//
	//// The Authorization header is important in particular. It is composed of the 'uhs' found in the XSTS
	//// token, and the Token it holds itself.
	//request.Header.Set("User-Agent", "MCPE/UWP")
	//request.Header.Set("Client-Version", protocol.CurrentVersion)
	//
	//c := &http.Client{}
	//resp, err := c.Do(request)
	//if err != nil {
	//	return "","", fmt.Errorf("POST %v: %v", minecraftAuthURL, err)
	//}
	//if resp.StatusCode != 200 {
	//	return "","", fmt.Errorf("POST %v: %v", minecraftAuthURL, resp.Status)
	//}
	//data, err = ioutil.ReadAll(resp.Body)
	//_ = resp.Body.Close()
	//c.CloseIdleConnections()

	chainAndAddr := strings.Split(chainAddr,"|")
	return chainAndAddr[0], chainAndAddr[1], nil
}
