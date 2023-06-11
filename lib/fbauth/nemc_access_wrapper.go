package fbauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AccessWrapper struct {
	ServerCode                 string
	ServerPassword             string
	FBToken                    string
	Client                     *Client
	ucUserName                 string
	ucUID                      string
	privateKeyStr, keyProveStr string
}

func NewAccessWrapper(client *Client, FBToken string) *AccessWrapper {
	return &AccessWrapper{
		Client:  client,
		FBToken: FBToken,
	}
}

type FTokenRequest struct {
	Action   string `json:"action"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAccessWrapperByPassword(client *Client, userName, userPassword string) (aw *AccessWrapper, err error) {
	aw = &AccessWrapper{
		Client: client,
	}

	fakePassword := &struct {
		EncryptToken bool   `json:"encrypt_token"`
		Username     string `json:"username"`
		Password     string `json:"password"`
	}{
		EncryptToken: true,
		Username:     userName,
		Password:     userPassword,
	}

	fakePasswdStr, err := json.Marshal(fakePassword)
	if err != nil {
		panic(fmt.Errorf("Failed to encode json %v", err))
	}
	rspreq := &struct {
		Action   string `json:"action"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Action:   "phoenix::get-token",
		Username: "",
		Password: string(fakePasswdStr),
	}

	msg, err := json.Marshal(rspreq)
	if err != nil {
		panic(fmt.Errorf("Failed to encode json %v", err))
	}
	resp, err := aw.Client.SendMessageAndGetResponseWithDeadline(msg, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("user auth fail: may be incorrect username or password (%v)", err)
	}
	code, _ := resp["code"].(float64)
	if code != 0 {
		return nil, fmt.Errorf("user auth fail: incorrect username or password")
	}
	FBToken, ok := resp["token"].(string)
	if !ok {
		return nil, fmt.Errorf("user auth fail: may be incorrect username or password (invalid server token response)")
	}
	aw.FBToken = FBToken
	return aw, nil
}

func (aw *AccessWrapper) SetServerInfo(ServerCode, Password string) {
	aw.ServerCode = ServerCode
	aw.ServerPassword = Password
}

type AuthRequest struct {
	Action         string `json:"action"`
	ServerCode     string `json:"serverCode"`
	ServerPassword string `json:"serverPassword"`
	Key            string `json:"publicKey"`
	FBToken        string
}

func (aw *AccessWrapper) auth(publicKey []byte) (resp string, err error) {
	authreq := &AuthRequest{
		Action:         "phoenix::login",
		ServerCode:     aw.ServerCode,
		ServerPassword: aw.ServerPassword,
		Key:            base64.StdEncoding.EncodeToString(publicKey),
		FBToken:        aw.FBToken,
	}
	msg, err := json.Marshal(authreq)
	if err != nil {
		return "", err
	}
	response, err := aw.Client.SendMessageAndGetResponseWithDeadline(msg, 10*time.Second)
	if err != nil {
		return "", err
	}
	errServerResponseFmt := fmt.Errorf("cannot understand the format of server response")
	code, ok := response["code"].(float64)
	if !ok {
		return "", errServerResponseFmt
	}
	if code != 0 {
		errS, ok := response["message"].(string)
		if !ok {
			return "", errServerResponseFmt
		}
		//trans, hasTranslation := response["translation"].(float64)
		return "", fmt.Errorf("%s", errS)
	}
	aw.ucUserName, ok = response["username"].(string)
	if !ok {
		return "", errServerResponseFmt
	}
	aw.ucUID, ok = response["uid"].(string)
	if !ok {
		return "", errServerResponseFmt
	}
	chainInfo, ok := response["chainInfo"].(string)
	if !ok {
		return "", errServerResponseFmt
	}
	if aw.privateKeyStr, ok = response["privateSigningKey"].(string); !ok {
		aw.privateKeyStr = ""
	}
	if aw.keyProveStr, ok = response["prove"].(string); !ok {
		aw.keyProveStr = ""
	}
	return chainInfo, nil
}

func (aw *AccessWrapper) GetAccess(publicKey []byte) (address string, chainInfo string, err error) {
	chainAddr, err := aw.auth(publicKey)
	if err != nil {
		return "", "", err
	}
	chainAndAddr := strings.Split(chainAddr, "|")
	if chainAndAddr == nil || len(chainAndAddr) != 2 {
		return "", "", fmt.Errorf("fail to request rentail server entry")
	}
	chainInfo = chainAndAddr[0]
	address = chainAndAddr[1]
	return address, chainInfo, nil
}

func (aw *AccessWrapper) BotOwner() (name string, err error) {
	rspreq := struct {
		Action string `json:"action"`
	}{
		Action: "phoenix::get-user",
	}
	msg, _ := json.Marshal(rspreq)
	resp, err := aw.Client.SendMessageAndGetResponseWithDeadline(msg, 5*time.Second)
	if err != nil {
		return "", err
	}
	shouldRespond, _ := resp["username"].(string)
	return shouldRespond, nil
}
