package fbauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	I18n "phoenixbuilder/fastbuilder/i18n"

	"github.com/pterm/pterm"
)

type secretLoadingTransport struct {
	secret string
}

func (s secretLoadingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.secret))
	return http.DefaultTransport.RoundTrip(req)
}

type ClientOptions struct {
	AuthServer          string
	RespondUserOverride string
}

func MakeDefaultClientOptions() *ClientOptions {
	return &ClientOptions{}
}

type ClientInfo struct {
	FBUCUsername string
	RespondUser  string
	Uid          string
	CertSigning  bool
	LocalKey     string
	LocalCert    string
}

type Client struct {
	ClientInfo
	client http.Client
	*ClientOptions
}

func parseAndPanic(message string) {
	error_regex := regexp.MustCompile("^(\\d{3} [a-zA-Z ]+)\n\n(.*?)($|\n)")
	err_matches := error_regex.FindAllStringSubmatch(message, 1)
	if len(err_matches) == 0 {
		panic(fmt.Errorf("Unknown error"))
	}
	panic(fmt.Errorf("%s: %s", err_matches[0][1], err_matches[0][2]))
}

func assertAndParse(resp *http.Response) map[string]interface{} {
	if resp.StatusCode == 503 {
		panic("API server is down")
	}
	_body, _ := io.ReadAll(resp.Body)
	body := string(_body)
	if resp.StatusCode != 200 {
		parseAndPanic(body)
	}
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(body), &ret)
	if err != nil {
		panic(fmt.Errorf("Error parsing API response: %v", err))
	}
	return ret
}

func CreateClient(options *ClientOptions) *Client {
	secret_res, err := http.Get(fmt.Sprintf("%s/api/new", options.AuthServer))
	if err != nil {
		panic(fmt.Errorf("Failed to contact with API"))
	}
	_secret_body, _ := io.ReadAll(secret_res.Body)
	secret_body := string(_secret_body)
	if secret_res.StatusCode == 503 {
		panic("API server is down")
	} else if secret_res.StatusCode != 200 {
		parseAndPanic(secret_body)
	}
	authclient := &Client{
		client: http.Client{Transport: secretLoadingTransport{
			secret: secret_body,
		}},
		ClientOptions: options,
		ClientInfo: ClientInfo{
			RespondUser: options.RespondUserOverride,
		},
	}
	return authclient
}

type AuthRequest struct {
	Action         string `json:"action"`
	ServerCode     string `json:"serverCode"`
	ServerPassword string `json:"serverPassword"`
	Key            string `json:"publicKey"`
	FBToken        string
	VersionId      int64 `json:"version_id"`
	//IgnoreVersionCheck bool `json:"ignore_version_check"`
}

// Ret: chain, ip, token, error
func (client *Client) Auth(ctx context.Context, serverCode string, serverPassword string, key string, fbtoken string, username string, password string) (string, string, string, error) {
	authreq := map[string]interface{}{}
	if len(fbtoken) != 0 {
		authreq["login_token"] = fbtoken
	} else if len(username) != 0 {
		authreq["username"] = username
		authreq["password"] = password
	}
	authreq["server_code"] = serverCode
	authreq["server_passcode"] = serverPassword
	authreq["client_public_key"] = key
	req_content, _ := json.Marshal(&authreq)
	r, err := client.client.Post(fmt.Sprintf("%s/api/phoenix/login", client.AuthServer), "application/json", bytes.NewBuffer(req_content))
	if err != nil {
		panic(err)
	}
	resp := assertAndParse(r)
	succ, _ := resp["success"].(bool)
	if !succ {
		err, _ := resp["message"].(string)
		trans, hasTranslation := resp["translation"].(float64)
		if hasTranslation && int(trans) != -1 {
			err = I18n.T(uint16(trans))
		}
		return "", "", "", fmt.Errorf("%s", err)
	}
	uc_username, _ := resp["username"].(string)
	u_uid, _ := resp["uid"].(string)
	client.FBUCUsername = uc_username
	client.Uid = u_uid
	str, _ := resp["chainInfo"].(string)
	client.CertSigning = true
	if signingKey, success := resp["privateSigningKey"].(string); success {
		client.LocalKey = signingKey
	} else {
		pterm.Error.Println("Failed to fetch privateSigningKey from server")
		client.CertSigning = false
		client.LocalKey = ""
	}
	if keyProve, success := resp["prove"].(string); success {
		client.LocalCert = keyProve
	} else {
		pterm.Error.Println("Failed to fetch keyProve from server")
		client.CertSigning = false
		client.LocalCert = ""
	}
	if !client.CertSigning {
		pterm.Error.Println("CertSigning is disabled for errors above.")
	}
	// If logged in by token, this field'd be empty
	token, _ := resp["token"].(string)
	respond_to, _ := resp["respond_to"].(string)
	if len(respond_to) != 0 && client.RespondUser == "" {
		client.RespondUser = respond_to
	}
	ip, _ := resp["ip_address"].(string)
	return str, ip, token, nil
}

func (client *Client) TransferData(content string) string {
	r, err := client.client.Get(fmt.Sprintf("%s/api/phoenix/transfer_start_type?content=%s", client.AuthServer, content))
	if err != nil {
		panic(err)
	}
	resp := assertAndParse(r)
	succ, _ := resp["success"].(bool)
	if !succ {
		err_m, _ := resp["message"].(string)
		panic(fmt.Errorf("Failed to transfer start type: %s", err_m))
	}
	data, _ := resp["data"].(string)
	return data
}

type FNumRequest struct {
	Data string `json:"data"`
}

func (client *Client) TransferCheckNum(data string) string {
	rspreq := &FNumRequest{
		Data: data,
	}
	msg, err := json.Marshal(rspreq)
	if err != nil {
		panic("Failed to encode json")
	}
	r, err := client.client.Post(fmt.Sprintf("%s/api/phoenix/transfer_check_num", client.AuthServer), "application/json", bytes.NewBuffer(msg))
	if err != nil {
		panic(err)
	}
	resp := assertAndParse(r)
	succ, _ := resp["success"].(bool)
	if !succ {
		err_m, _ := resp["message"].(string)
		panic(fmt.Errorf("Failed to transfer check num: %s", err_m))
	}
	val, _ := resp["value"].(string)
	return val
}
