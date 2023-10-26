package fbauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	I18n "phoenixbuilder/fastbuilder/i18n"
)

type AccessWrapper struct {
	ServerCode     string
	ServerPassword string
	Token          string
	Client         *Client
	Username       string
	Password       string
}

func NewAccessWrapper(Client *Client, ServerCode, ServerPassword, Token, username, password string) *AccessWrapper {
	return &AccessWrapper{
		Client:         Client,
		ServerCode:     ServerCode,
		ServerPassword: ServerPassword,
		Token:          Token,
		Username:       username,
		Password:       password,
	}
}

func (aw *AccessWrapper) GetAccess(ctx context.Context, publicKey []byte) (address string, chainInfo string, err error) {
	pubKeyData := base64.StdEncoding.EncodeToString(publicKey)
	chainAddr, ip, token, err := aw.Client.Auth(ctx, aw.ServerCode, aw.ServerPassword, pubKeyData, aw.Token, aw.Username, aw.Password)
	if err != nil {
		return "", "", err
	}
	if len(token) != 0 {
		homedir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(I18n.T(I18n.Warning_UserHomeDir))
			homedir = "."
		}
		fbconfigdir := filepath.Join(homedir, ".config", "fastbuilder")
		os.MkdirAll(fbconfigdir, 0755)
		ptoken := filepath.Join(fbconfigdir, "fbtoken")
		// 0600: -rw-------
		token_file, err := os.OpenFile(ptoken, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return "", "", err
		}
		_, err = token_file.WriteString(token)
		if err != nil {
			return "", "", err
		}
		token_file.Close()
	}
	return ip, chainAddr, nil
}
