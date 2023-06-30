package fbauth

import (
	"context"
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

func GetTokenByPassword(connectCtx context.Context, client *Client, userName, userPassword string) (writeBackToken string, err error) {
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
	resp, err := client.SendMessageAndGetResponseWithDeadline(msg, connectCtx)
	if err != nil {
		return "", fmt.Errorf("user auth fail: may be incorrect username or password (%v)", err)
	}
	code, _ := resp["code"].(float64)
	if code != 0 {
		return "", fmt.Errorf("user auth fail: %v", resp["message"].(string))
	}
	FBToken, ok := resp["token"].(string)
	if !ok {
		return "", fmt.Errorf("user auth fail: may be incorrect username or password (invalid server token response)")
	}
	return FBToken, nil
}

func (aw *AccessWrapper) SetServerInfo(ServerCode, Password string) {
	aw.ServerCode = ServerCode
	aw.ServerPassword = Password
}

func (aw *AccessWrapper) GetFBUid() string {
	return aw.ucUID
}

type AuthRequest struct {
	Action            string `json:"action"`
	ServerCode        string `json:"serverCode"`
	ServerPassword    string `json:"serverPassword"`
	Key               string `json:"publicKey"`
	FBToken           string
	ProtocolVersionId int64 `json:"version_id"`
}

func (aw *AccessWrapper) auth(ctx context.Context, publicKey []byte) (resp string, err error) {
	authreq := &AuthRequest{
		Action:            "phoenix::login",
		ServerCode:        aw.ServerCode,
		ServerPassword:    aw.ServerPassword,
		Key:               base64.StdEncoding.EncodeToString(publicKey),
		FBToken:           aw.FBToken,
		ProtocolVersionId: 2,
	}
	msg, err := json.Marshal(authreq)
	if err != nil {
		return "", err
	}
	response, err := aw.Client.SendMessageAndGetResponseWithDeadline(msg, ctx)
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

func (aw *AccessWrapper) getAccess(ctx context.Context, publicKey []byte) (address string, chainInfo string, err error) {
	chainAddr, err := aw.auth(ctx, publicKey)
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

func (aw *AccessWrapper) GetAccess(ctx context.Context, publicKey []byte) (address string, chainInfo string, err error) {
	// TODO make it configurable
	maxRetryTimes := 3
	fastRetryDelay := time.Second
	for retryTimes := 0; retryTimes < maxRetryTimes; retryTimes++ {
		address, chainInfo, err = aw.getAccess(ctx, publicKey)
		if err != nil && strings.Contains(err.Error(), "Link server processing error") {
			fmt.Println("Link server processing error, retrying...")
			time.Sleep(fastRetryDelay)
			continue
		} else {
			break
		}
	}
	return address, chainInfo, err
}

func (aw *AccessWrapper) BotOwner(ctx context.Context) (name string, err error) {
	rspreq := struct {
		Action string `json:"action"`
	}{
		Action: "phoenix::get-user",
	}
	msg, _ := json.Marshal(rspreq)
	resp, err := aw.Client.SendMessageAndGetResponseWithDeadline(msg, ctx)
	if err != nil {
		return "", err
	}
	shouldRespond, _ := resp["username"].(string)
	return shouldRespond, nil
}

type RPCEncRequest struct {
	Action  string `json:"action"`
	Content string `json:"content"`
	Uid     string `json:"uid"`
}

func (aw *AccessWrapper) TransferData(ctx context.Context, content string, uid string) (string, error) {
	rspreq := &RPCEncRequest{
		Action:  "phoenix::transfer-data",
		Content: content,
		Uid:     uid,
	}
	msg, err := json.Marshal(rspreq)
	if err != nil {
		panic("Failed to encode json")
	}
	resp, err := aw.Client.SendMessageAndGetResponseWithDeadline(msg, ctx)
	if err != nil {
		return "", err
	}
	code, _ := resp["code"].(float64)
	if code != 0 {
		panic("Failed to transfer start type")
	}
	data, _ := resp["data"].(string)
	return data, nil
}

type RPCNumRequest struct {
	Action string `json:"action"`
	First  string `json:"1st"`
	Second string `json:"2nd"`
	Third  int64  `json:"3rd"`
}

func (aw *AccessWrapper) TransferCheckNum(ctx context.Context, first string, second string, third int64) (string, string, string, error) {
	rspreq := &RPCNumRequest{
		Action: "phoenix::transfer-check-num",
		First:  first,
		Second: second,
		Third:  third,
	}
	msg, err := json.Marshal(rspreq)
	if err != nil {
		panic("Failed to encode json")
	}
	resp, err := aw.Client.SendMessageAndGetResponseWithDeadline(msg, ctx)
	if err != nil {
		return "", "", "", err
	}
	code, _ := resp["code"].(float64)
	if code != 0 {
		return "", "", "", fmt.Errorf("failed to transfer checknum %v", resp["message"])
	}
	valM, _ := resp["valM"].(string)
	valS, _ := resp["valS"].(string)
	valT, _ := resp["valT"].(string)
	return valM, valS, valT, nil
}
