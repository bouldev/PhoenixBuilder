package fbauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"strings"
)

type AccessWrapper struct {
	ServerCode string
	Password   string
	Token      string
	Client     *Client
}

func NewAccessWrapper(Client *Client, ServerCode, Password, Token string) *AccessWrapper {
	return &AccessWrapper{
		Client:     Client,
		ServerCode: ServerCode,
		Password:   Password,
		Token:      Token,
	}
}

func (aw *AccessWrapper) GetAccess(ctx context.Context, publicKey []byte) (address string, chainInfo string, err error) {
	pubKeyData := base64.StdEncoding.EncodeToString(publicKey)
	chainAddr, code, err := aw.Client.Auth(ctx, aw.ServerCode, aw.Password, pubKeyData, aw.Token)
	chainAndAddr := strings.Split(chainAddr, "|")
	if err != nil {
		if code == -3 {
			homedir, err := os.UserHomeDir()
			if err != nil {
				fmt.Println(I18n.T(I18n.Warning_UserHomeDir))
				homedir = "."
			}
			fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
			os.MkdirAll(fbconfigdir, 0755)
			token := filepath.Join(fbconfigdir, "fbtoken")
			os.Remove(token)
		}
		return "", "", err
	}
	if chainAndAddr == nil || len(chainAndAddr) != 2 {
		return "", "", fmt.Errorf(I18n.T(I18n.Auth_FailedToRequestEntry_TryAgain))
	}
	chainInfo = chainAndAddr[0]
	address = chainAndAddr[1]
	return address, chainInfo, nil
}
