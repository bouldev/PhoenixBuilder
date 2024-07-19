package fbauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/pterm/pterm"

	I18n "phoenixbuilder/fastbuilder/i18n"
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
	GrowthLevel  int
	RespondTo    string
	Uid          string
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

func assertAndParse[T any](resp *http.Response) T {
	if resp.StatusCode == 503 {
		panic("API server is down")
	}
	_body, _ := io.ReadAll(resp.Body)
	body := string(_body)
	if resp.StatusCode != 200 {
		parseAndPanic(body)
	}
	var ret T
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
			RespondTo: options.RespondUserOverride,
		},
	}
	return authclient
}

// 客户端向验证服务器发送的请求体，
// 用于获得 FBToken，
// 或使客户端登录到网易租赁服。
// AuthResponse 是该请求体对应的响应体
type AuthRequest struct {
	/*
		此字段非空时，则下方 UserName 和 Password 为空，
		否则反之。

		当 FBToken 或 UserName、Password 二者中任意一个
		填写的值正确时，用户将登录到用户中心，然后进入租赁服
	*/
	FBToken string `json:"login_token,omitempty"`

	UserName string `json:"username,omitempty"` // 用户在用户中心的用户名
	Password string `json:"password,omitempty"` // 用户在用户中心的密码

	ServerCode     string `json:"server_code"`     // 要进入的租赁服的 服务器号
	ServerPassword string `json:"server_passcode"` // 该租赁服的 密码

	ClientPublicKey string `json:"client_public_key"` // ...
}

// 验证服务器对 AuthRequest 的响应体
type AuthResponse struct {
	/*
		描述请求的成功状态。

		如果成功，则其余的所有非可选字段都将有值，
		这也包括 Message 本身。

		如果失败，除了本字段和 Message 以外，
		其余所有字段都为默认的零值，或空字符串，
		同时 Message 会展示对应的失败原因
	*/
	SuccessStates bool   `json:"success"`
	ServerMessage string `json:"server_msg,omitempty"` // 来自验证服务器的消息 (可能不存在；为可选功能)
	Message

	BotName  string   `json:"username"`            // 机器人的 游戏名称
	BotUID   string   `json:"uid"`                 // 机器人的 UID
	BotLevel int      `json:"growth_level"`        // 机器人的等级
	SkinInfo SkinInfo `json:"skin_info,omitempty"` // 机器人的皮肤信息 (可能不存在；需要验证服务器支持)

	FBToken    string `json:"token"`      // ...
	MasterName string `json:"respond_to"` // 机器人主人的游戏名称

	RentalServerIP string `json:"ip_address"` // 目标租赁服的 IP 地址
	ChainInfo      string `json:"chainInfo"`  // ...
}

// 描述 AuthResponse 所附带的额外信息
type Message struct {
	/*
		若 AuthRequest 成功，
		则对于原生的 FastBuilder 的验证服务器(mv4)，
		此字段为 "正常返回"；
		否则，对于 咕咕酱及其开发团队 的验证服务器，
		此字段为 "well down"。

		当 AuthRequest 失败时，
		若此字段非空，则它将阐明对应的失败原因，
		否则，由下方的 Translation 揭示具体的原因
	*/
	Information string `json:"message,omitempty"`
	// 表示错误码，且可以与 i18n 中所记的映射对应。
	// 如果不存在，则其默认值为 0，
	// 如果未使用，则其默认值为 -1
	Translation int `json:"translation,omitempty"`
}

// 描述 AuthResponse 中附带的 机器人 的皮肤信息
type SkinInfo struct {
	ItemID          string `json:"entity_id"` // 皮肤的资源 ID
	SkinDownloadURL string `json:"res_url"`   // 皮肤的下载地址 [需要验证]
}

// Ret: chain, ip, token, error
func (client *Client) Auth(
	ctx context.Context,
	serverCode string, serverPassword string,
	key string,
	fbtoken string, username string, password string,
) (AuthResponse, error) {
	var authResponse AuthResponse
	// prepare
	request := AuthRequest{
		FBToken:         fbtoken,
		UserName:        username,
		Password:        password,
		ServerCode:      serverCode,
		ServerPassword:  serverPassword,
		ClientPublicKey: key,
	}
	authRequest, _ := json.Marshal(request)
	// pack request and marshal to binary
	httpResponse, err := client.client.Post(
		fmt.Sprintf("%s/api/phoenix/login", client.AuthServer),
		"application/json",
		bytes.NewBuffer(authRequest),
	)
	if err != nil {
		panic(fmt.Sprintf("Auth: %v", err))
	}
	authResponse = assertAndParse[AuthResponse](httpResponse)
	// get response
	if !authResponse.SuccessStates {
		failedReason := authResponse.Message
		err := failedReason.Information
		if t := failedReason.Translation; t != -1 && t != 0 {
			err = I18n.T(uint16(t))
		}
		return AuthResponse{}, fmt.Errorf("%s", err)
	}
	// if reuqest failed
	if len(authResponse.ServerMessage) > 0 {
		pterm.Println(pterm.LightGreen(I18n.T(I18n.Auth_MessageFromAuthServer)))
		pterm.Println(pterm.LightGreen(authResponse.ServerMessage))
	}
	// server message
	client.FBUCUsername = authResponse.BotName
	client.Uid = authResponse.BotUID
	client.GrowthLevel = authResponse.BotLevel
	if len(authResponse.MasterName) > 0 && client.RespondTo == "" {
		client.RespondTo = authResponse.MasterName
	}
	// set value
	return authResponse, nil
	// return
}

func (client *Client) TransferData(content string) string {
	r, err := client.client.Get(fmt.Sprintf("%s/api/phoenix/transfer_start_type?content=%s", client.AuthServer, content))
	if err != nil {
		panic(err)
	}
	resp := assertAndParse[map[string]any](r)
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
	resp := assertAndParse[map[string]any](r)
	succ, _ := resp["success"].(bool)
	if !succ {
		err_m, _ := resp["message"].(string)
		panic(fmt.Errorf("Failed to transfer check num: %s", err_m))
	}
	val, _ := resp["value"].(string)
	return val
}
